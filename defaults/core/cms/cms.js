/* Plenti CMS - a dependency-free editor for Plenti content.
   Loads its data from /admin/data.json and saves through the configured
   provider (local dev server or a git host). */
(function () {
  "use strict";

  var IMAGE_EXTS = [".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg", ".avif", ".apng"];

  var state = {
    data: null,        // payload from data.json
    view: "content",   // "content" | "media"
    current: null,     // index into data.content of the item being edited
    fields: null,      // working copy of the fields being edited
    notice: null,
    jsonMode: false,
    jsonText: "",
    rootEl: null,      // element the full CMS app is mounted into
    trayItem: null,    // data.content node for the page shown in the tray
    trayFields: null,  // working copy of the tray's field values
    trayOpen: false
  };

  function h(tag, attrs, children) {
    var el = document.createElement(tag);
    if (attrs) {
      Object.keys(attrs).forEach(function (k) {
        if (k === "class") el.className = attrs[k];
        else if (k.slice(0, 2) === "on") el.addEventListener(k.slice(2), attrs[k]);
        else if (k === "value") el.value = attrs[k];
        else if (k === "checked") el.checked = !!attrs[k];
        else if (attrs[k] != null) el.setAttribute(k, attrs[k]);
      });
    }
    (children || []).forEach(function (c) {
      if (c == null) return;
      el.appendChild(typeof c === "string" ? document.createTextNode(c) : c);
    });
    return el;
  }

  function isImage(path) {
    var ext = path.slice(path.lastIndexOf(".")).toLowerCase();
    return IMAGE_EXTS.indexOf(ext) !== -1;
  }

  /* ---------- Providers ---------- */

  function getToken() {
    return localStorage.getItem("plenti_cms_token") || "";
  }

  function commitLocal(commitList) {
    var body = commitList.map(function (item) {
      return {
        action: item.action === "upsert" ? "create" : item.action,
        encoding: item.encoding || "text",
        file: item.file.replace(/^\/+/, ""),
        contents: item.encoding === "base64" ? item.contents.split(",")[1] : item.contents
      };
    });
    return fetch("/postlocal", {
      method: "POST",
      headers: { "Content-Type": "application/json; charset=utf-8" },
      body: JSON.stringify(body)
    }).then(function (res) {
      if (!res.ok) {
        return res.text().then(function (msg) {
          throw new Error("Save failed (" + res.status + "): " + (msg.trim() || res.statusText));
        });
      }
    });
  }

  function commitGitlab(commitList) {
    var cms = state.data.env.cms;
    var repoUrl = new URL(cms.repo);
    var apiBase = cms.api_base_url || repoUrl.origin + "/api/v4";
    var projectUrl = apiBase + "/projects/" + encodeURIComponent(repoUrl.pathname.slice(1));
    var headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer " + getToken()
    };
    return Promise.all(commitList.map(function (item) {
      var entry = {
        action: item.action,
        file_path: item.file.replace(/^\/+/, ""),
        encoding: item.encoding || "text",
        content: item.encoding === "base64" ? item.contents.split(",")[1] : item.contents
      };
      if (item.action !== "upsert") return entry;
      return fetch(projectUrl + "/repository/files/" + encodeURIComponent(entry.file_path) +
        "?ref=" + encodeURIComponent(cms.branch), { method: "HEAD", headers: headers })
        .then(function (head) {
          entry.action = head.status === 404 ? "create" : "update";
          var lastCommit = head.headers.get("X-Gitlab-Last-Commit-Id");
          if (lastCommit) entry.last_commit_id = lastCommit;
          return entry;
        });
    })).then(function (actions) {
      return fetch(projectUrl + "/repository/commits", {
        method: "POST",
        headers: headers,
        body: JSON.stringify({
          branch: cms.branch,
          commit_message: "Update " + (commitList.length > 1 ? commitList.length + " files" : commitList[0].file),
          actions: actions
        })
      });
    }).then(function (res) {
      if (!res.ok) {
        return res.json().then(function (err) {
          throw new Error("Publish failed: " + (err.error || err.message || res.statusText));
        });
      }
    });
  }

  function commitGitea(commitList) {
    var cms = state.data.env.cms;
    var repoUrl = new URL(cms.repo);
    var apiBase = cms.api_base_url || repoUrl.origin + "/api/v1";
    var repoPath = repoUrl.pathname.slice(1);
    var headers = {
      "Content-Type": "application/json",
      "Authorization": "token " + getToken()
    };
    // Gitea has no batch commit endpoint, so write files sequentially.
    return commitList.reduce(function (chain, item) {
      return chain.then(function () {
        var file = item.file.replace(/^\/+/, "");
        var fileUrl = apiBase + "/repos/" + repoPath + "/contents/" + file;
        var content = item.encoding === "base64" ? item.contents.split(",")[1] : btoa(unescape(encodeURIComponent(item.contents)));
        return fetch(fileUrl + "?ref=" + encodeURIComponent(cms.branch), { headers: headers })
          .then(function (res) { return res.ok ? res.json() : null; })
          .then(function (existing) {
            if (item.action === "delete") {
              if (!existing) return;
              return fetch(fileUrl, {
                method: "DELETE",
                headers: headers,
                body: JSON.stringify({ branch: cms.branch, message: "Delete " + file, sha: existing.sha })
              }).then(function (res) { if (!res.ok) throw new Error("Delete failed for " + file); });
            }
            var body = {
              branch: cms.branch,
              message: (existing ? "Update " : "Create ") + file,
              content: content
            };
            if (existing) body.sha = existing.sha;
            return fetch(fileUrl, {
              method: existing ? "PUT" : "POST",
              headers: headers,
              body: JSON.stringify(body)
            }).then(function (res) { if (!res.ok) throw new Error("Save failed for " + file); });
          });
      });
    }, Promise.resolve());
  }

  function commit(commitList) {
    var cms = state.data.env.cms;
    var provider = (cms.provider || "local").toLowerCase();
    if (state.data.env.local || provider === "local") return commitLocal(commitList);
    if (provider === "gitlab") return commitGitlab(commitList);
    if (provider === "gitea" || provider === "forgejo") return commitGitea(commitList);
    return Promise.reject(new Error("Unknown CMS provider: " + provider));
  }

  /* ---------- Editor fields ---------- */

  function fieldInput(key, type, value, onChange, schema) {
    schema = schema || {};
    var def = schema[key];
    var options = def && def.options;

    if (type === "boolean") {
      return h("input", { type: "checkbox", checked: !!value, onchange: function (e) { onChange(e.target.checked); } });
    }
    if (type === "number") {
      return h("input", { type: "number", value: value == null ? "" : value, onchange: function (e) { onChange(e.target.value === "" ? null : Number(e.target.value)); } });
    }
    if (type === "date") {
      return h("input", { type: "date", value: value || "", onchange: function (e) { onChange(e.target.value); } });
    }
    if (type === "time") {
      return h("input", { type: "time", value: value || "", onchange: function (e) { onChange(e.target.value); } });
    }
    if ((type === "select" || type === "radio" || type === "checkbox") && options) {
      if (type === "select") {
        return h("select", { onchange: function (e) { onChange(e.target.value); } },
          options.map(function (opt) {
            return h("option", { value: opt, selected: opt === value ? "selected" : null }, [opt]);
          }));
      }
      return h("div", { class: "plenti-checks" }, options.map(function (opt) {
        var input = type === "radio"
          ? h("input", { type: "radio", name: key, value: opt, checked: opt === value, onchange: function () { onChange(opt); } })
          : h("input", { type: "checkbox", value: opt, checked: Array.isArray(value) && value.indexOf(opt) !== -1,
              onchange: function (e) {
                var next = Array.isArray(value) ? value.slice() : [];
                if (e.target.checked) next.push(opt);
                else next = next.filter(function (v) { return v !== opt; });
                onChange(next);
              } });
        return h("label", {}, [input, " " + opt]);
      }));
    }
    if (type === "media") {
      return h("select", { onchange: function (e) { onChange(e.target.value); } },
        [h("option", { value: "" }, ["-- choose media --"])].concat(
          (state.data.media || []).map(function (m) {
            return h("option", { value: m, selected: m === value ? "selected" : null }, [m]);
          })));
    }
    if (type === "list") {
      return h("input", { type: "text", value: Array.isArray(value) ? value.join(", ") : (value || ""),
        onchange: function (e) {
          onChange(e.target.value.split(",").map(function (s) { return s.trim(); }).filter(Boolean));
        } });
    }
    if (type === "wysiwyg" || type === "textarea") {
      return h("textarea", { oninput: function (e) { onChange(e.target.value); } }, [value || ""]);
    }
    if (type === "id") {
      return h("input", { type: "text", value: value || "", disabled: "disabled" });
    }
    // Nested objects/unknown types fall back to a JSON textarea.
    if (value !== null && typeof value === "object") {
      return h("textarea", {
        onchange: function (e) {
          try { onChange(JSON.parse(e.target.value)); }
          catch (err) { notify("Invalid JSON in " + key + ": " + err.message, true); }
        }
      }, [JSON.stringify(value, null, 2)]);
    }
    return h("input", { type: "text", value: value == null ? "" : value, oninput: function (e) { onChange(e.target.value); } });
  }

  function schemaType(schema, key, value) {
    var def = schema[key];
    if (typeof def === "string") return def;
    if (def && def.type) return def.type;
    if (typeof value === "boolean") return "boolean";
    if (typeof value === "number") return "number";
    if (Array.isArray(value)) return "list";
    return "text";
  }

  function currentItem() { return state.data.content[state.current]; }
  function currentType() { return currentItem().type; }

  // Rebuild a fields object with keys in the source JSON file's original
  // order (Go marshals the fields map alphabetically in data.json).
  function orderedFields(item, values) {
    var src = values || (item && item.fields) || {};
    var out = {};
    ((item && item.fieldOrder) || []).forEach(function (k) {
      if (Object.prototype.hasOwnProperty.call(src, k)) out[k] = src[k];
    });
    Object.keys(src).forEach(function (k) {
      if (!Object.prototype.hasOwnProperty.call(out, k)) out[k] = src[k];
    });
    return out;
  }

  /* ---------- Rendering ---------- */

  function notify(msg, isError) {
    state.notice = { msg: msg, isError: !!isError };
    render();
  }

  function renderSidebar() {
    var byType = {};
    state.data.content.forEach(function (c, i) {
      (byType[c.type] = byType[c.type] || []).push({ item: c, index: i });
    });

    var sections = Object.keys(byType).sort().map(function (type) {
      return h("div", {}, [
        h("h2", {}, [type]),
        h("div", {}, byType[type].map(function (entry) {
          return h("button", {
            class: "plenti-nav-item" + (state.view === "content" && state.current === entry.index ? " plenti-active" : ""),
            onclick: function () { openItem(entry.index); }
          }, [entry.item.path]);
        }).concat([
          h("button", {
            class: "plenti-nav-item",
            onclick: function () { addItem(type); }
          }, ["+ add"])
        ]))
      ]);
    });

    sections.push(h("div", {}, [
      h("h2", {}, ["media"]),
      h("button", {
        class: "plenti-nav-item" + (state.view === "media" ? " plenti-active" : ""),
        onclick: function () { state.view = "media"; state.jsonMode = false; render(); }
      }, ["browse"])
    ]));

    return h("aside", { class: "plenti-sidebar" }, [h("h1", {}, ["Plenti CMS"])].concat(sections));
  }

  function renderEditor() {
    var item = currentItem();
    var schema = state.data.schemas[item.type] || {};
    var children = [h("h2", {}, [item.filepath])];

    if (state.notice) {
      children.push(h("div", { class: "plenti-notice" + (state.notice.isError ? " plenti-error" : "") }, [state.notice.msg]));
    }

    if (state.jsonMode) {
      children.push(h("textarea", {
        class: "plenti-json-editor",
        onchange: function (e) { state.jsonText = e.target.value; }
      }, [state.jsonText]));
    } else {
      Object.keys(state.fields).forEach(function (key) {
        var type = schemaType(schema, key, state.fields[key]);
        children.push(h("div", { class: "plenti-field" }, [
          h("label", {}, [key]),
          fieldInput(key, type, state.fields[key], function (v) { state.fields[key] = v; }, schema)
        ]));
      });
    }

    children.push(h("div", { class: "plenti-actions" }, [
      h("button", { class: "plenti-btn", onclick: saveItem }, ["Save"]),
      h("button", {
        class: "plenti-btn plenti-btn-secondary",
        onclick: function () { toggleJsonMode(); }
      }, [state.jsonMode ? "Form" : "JSON"]),
      h("button", { class: "plenti-btn plenti-btn-danger", onclick: deleteItem }, ["Delete"])
    ]));

    return h("main", { class: "plenti-main" }, children);
  }

  function renderMedia() {
    var children = [h("h2", {}, ["Media"])];
    if (state.notice) {
      children.push(h("div", { class: "plenti-notice" + (state.notice.isError ? " plenti-error" : "") }, [state.notice.msg]));
    }
    children.push(h("input", {
      type: "file",
      onchange: function (e) {
        var file = e.target.files[0];
        if (!file) return;
        var reader = new FileReader();
        reader.onload = function () {
          commit([{ file: "media/" + file.name, action: "upsert", encoding: "base64", contents: reader.result }])
            .then(function () {
              state.data.media.push("/media/" + file.name);
              notify("Uploaded " + file.name);
            })
            .catch(function (err) { notify(err.message, true); });
        };
        reader.readAsDataURL(file);
      }
    }));
    children.push(h("div", { class: "plenti-media-grid" }, (state.data.media || []).map(function (m) {
      var full = state.data.env.baseurl.replace(/\/$/, "") + "/" + m.replace(/^\/+/, "");
      return h("div", { class: "plenti-media-item", onclick: function () { window.open(full, "_blank"); } }, [
        isImage(m) ? h("img", { src: full, alt: m }) : h("div", {}, ["\uD83D\uDCC4"]),
        h("div", { class: "plenti-media-name" }, [m])
      ]);
    })));
    return h("main", { class: "plenti-main" }, children);
  }

  function render() {
    var root = state.rootEl;
    if (!root) return;
    root.innerHTML = "";
    var main = state.view === "media" ? renderMedia() : renderEditor();
    root.appendChild(h("div", { class: "plenti-layout" }, [renderSidebar(), main]));
  }

  /* ---------- Actions ---------- */

  function openItem(index) {
    state.view = "content";
    state.current = index;
    state.fields = JSON.parse(JSON.stringify(orderedFields(state.data.content[index])));
    state.jsonMode = false;
    state.notice = null;
    render();
  }

  function toggleJsonMode() {
    if (state.jsonMode) {
      try {
        state.fields = JSON.parse(state.jsonText);
        state.jsonMode = false;
      } catch (err) {
        notify("Invalid JSON: " + err.message, true);
        return;
      }
    } else {
      state.jsonText = JSON.stringify(state.fields, null, 2);
      state.jsonMode = true;
    }
    render();
  }

  function addItem(type) {
    var name = prompt("Name for the new " + type + " item (filename without extension):");
    if (!name) return;
    name = name.trim().toLowerCase().replace(/[^a-z0-9-_]+/g, "-");
    if (!name) return;
    var defaults = state.data.defaults[type] || {};
    var item = {
      type: type,
      path: "/" + type + "/" + name,
      filename: name,
      filepath: type + "/" + name + ".json",
      fields: JSON.parse(JSON.stringify(defaults))
    };
    commit([{ file: "content/" + item.filepath, action: "create", contents: JSON.stringify(item.fields, null, 2) }])
      .then(function () {
        state.data.content.push(item);
        openItem(state.data.content.length - 1);
        notify("Created " + item.filepath);
      })
      .catch(function (err) { notify(err.message, true); });
  }

  function saveItem() {
    if (state.jsonMode) {
      try { state.fields = JSON.parse(state.jsonText); }
      catch (err) { notify("Invalid JSON: " + err.message, true); return; }
    }
    var item = currentItem();
    commit([{ file: "content/" + item.filepath, action: "update", contents: JSON.stringify(orderedFields(item, state.fields), null, 2) }])
      .then(function () {
        item.fields = JSON.parse(JSON.stringify(state.fields));
        notify("Saved " + item.filepath);
      })
      .catch(function (err) { notify(err.message, true); });
  }

  function deleteItem() {
    var item = currentItem();
    if (!confirm("Delete " + item.filepath + "?")) return;
    commit([{ file: "content/" + item.filepath, action: "delete" }])
      .then(function () {
        state.data.content.splice(state.current, 1);
        if (state.data.content.length) openItem(Math.max(0, state.current - 1));
        else { state.current = null; state.view = "media"; render(); }
      })
      .catch(function (err) { notify(err.message, true); });
  }

  /* ---------- Auth ---------- */

  function currentProvider() {
    var cms = (state.data && state.data.env.cms) || {};
    return (cms.provider || "local").toLowerCase();
  }

  function isAuthenticated() {
    // The local provider needs no credentials, but only works against the
    // dev server (which sets env.local). An explicit local logout is
    // remembered in localStorage so it survives page reloads.
    if (currentProvider() === "local") {
      return !!(state.data && state.data.env.local) && !localStorage.getItem("plenti_cms_local_logout");
    }
    return !!getToken();
  }

  // Publish the magic `user` value. It is always an object so templates can
  // safely use `user.loggedIn` (false during SSR and when logged out).
  // Updating Pattr's root data makes any `{if user.loggedIn}` blocks in
  // layouts re-render automatically.
  function publishUser() {
    var u = {
      loggedIn: isAuthenticated(),
      provider: currentProvider(),
      login: login,
      logout: logout
    };
    window.user = u;
    if (window.Pattr && window.Pattr.data) {
      try { window.Pattr.data.user = u; } catch (e) { /* scope is read-only */ }
    }
    window.dispatchEvent(new CustomEvent("plenti:cms-auth"));
  }

  function login() {
    var provider = currentProvider();
    if (provider === "local") {
      localStorage.removeItem("plenti_cms_local_logout");
    } else {
      var token = prompt("Enter a personal access token for " + provider + " (stored in this browser):");
      if (token) localStorage.setItem("plenti_cms_token", token);
    }
    publishUser();
  }

  function logout() {
    localStorage.removeItem("plenti_cms_token");
    if (currentProvider() === "local") localStorage.setItem("plenti_cms_local_logout", "1");
    publishUser();
  }

  /* ---------- Boot ---------- */

  function loadData() {
    if (state.data) return Promise.resolve(state.data);
    return fetch("/admin/data.json")
      .then(function (res) { return res.json(); })
      .then(function (data) { state.data = data; return data; });
  }

  function mount(el) {
    if (typeof el === "string") el = document.querySelector(el);
    if (!el) return;
    state.rootEl = el;
    loadData().then(function () {
      if (state.data.content.length) openItem(0);
      else { state.view = "media"; render(); }
    }).catch(function (err) {
      el.textContent = "Could not load CMS data: " + err.message;
    });
  }

  /* ---------- Inline admin bar (embedded CMS) ---------- */

  // Find the content node backing the page currently shown in the browser.
  function currentPageItem() {
    var path = window.location.pathname.replace(/\/+$/, "") || "/";
    var items = state.data.content || [];
    for (var i = 0; i < items.length; i++) {
      if ((items[i].path || "").replace(/\/+$/, "") === path || items[i].path === path) return items[i];
    }
    // A page's path may be "/" while the browser shows "/index" and vice versa.
    if (path === "/index") {
      for (var j = 0; j < items.length; j++) if (items[j].path === "/") return items[j];
    }
    return null;
  }

  // Push an edit straight into Pattr's root data so the live page re-renders.
  function liveUpdate(key, value) {
    state.trayFields[key] = value;
    if (window.Pattr && window.Pattr.data && window.Pattr.data.content) {
      var c = window.Pattr.data.content;
      c.fields = c.fields || {};
      c.fields[key] = value;
      // Reassign so Pattr's Proxy set() fires and walkDom re-renders bindings.
      window.Pattr.data.content = Object.assign({}, c);
    }
  }

  function renderTray(tray, statusEl) {
    // Clear previous content but keep the drag-resize handle in place.
    var resizeHandle = tray.querySelector(".plenti-tray-resize");
    tray.innerHTML = "";
    if (resizeHandle) tray.appendChild(resizeHandle);
    var item = state.trayItem;
    if (!item) {
      tray.appendChild(h("p", { class: "plenti-admin-empty" }, ["No content source found for this page."]));
      return;
    }
    var schema = (state.data.schemas && state.data.schemas[item.type]) || {};
    tray.appendChild(h("h3", {}, ["Edit " + (item.filepath || item.path)]));
    Object.keys(state.trayFields).forEach(function (key) {
      if (key === "components") return; // handled by layouts, not editable here
      var type = schemaType(schema, key, state.trayFields[key]);
      if (type === "media") type = "text"; // keep the tray simple: path inputs
      tray.appendChild(h("div", { class: "plenti-field" }, [
        h("label", {}, [key]),
        fieldInput(key, type, state.trayFields[key], function (v) { liveUpdate(key, v); }, schema)
      ]));
    });
    tray.appendChild(h("div", { class: "plenti-admin-actions" }, [
      h("button", {
        class: "plenti-btn primary",
        onclick: function () { saveTray(statusEl); }
      }, ["Save"]),
      h("button", {
        class: "plenti-btn",
        onclick: function () { setTrayOpen(false); }
      }, ["Close"])
    ]));
  }

  function saveTray(statusEl) {
    var item = state.trayItem;
    if (!item) return;
    statusEl.textContent = "Saving...";
    statusEl.classList.remove("success", "error");
    commit([{ file: "content/" + item.filepath, action: "update", contents: JSON.stringify(orderedFields(item, state.trayFields), null, 2) }])
      .then(function () {
        item.fields = JSON.parse(JSON.stringify(state.trayFields));
        statusEl.innerHTML = ICONS.check + "Saved";
        statusEl.classList.add("success");
        setTimeout(function () {
          statusEl.textContent = "";
          statusEl.classList.remove("success");
        }, 3000);
      })
      .catch(function (err) {
        statusEl.textContent = "Error: " + err.message;
        statusEl.classList.add("error");
      });
  }

  function setTrayOpen(open) {
    state.trayOpen = open;
    var tray = document.querySelector(".plenti-admin-tray");
    var toggle = document.querySelector(".plenti-admin-bar .plenti-edit-toggle");
    var label = toggle && toggle.querySelector(".plenti-label");
    var iconEl = toggle && toggle.querySelector(".plenti-icon");
    if (open && tray) renderTray(tray, document.querySelector(".plenti-admin-status"));
    if (tray) {
      // Force a reflow so the slide transition runs on first open.
      if (open) void tray.offsetWidth;
      tray.classList.toggle("open", open);
    }
    if (label) label.textContent = open ? "View" : "Edit";
    if (iconEl) iconEl.innerHTML = open ? ICONS.eye : ICONS.edit;
    document.body.classList.toggle("plenti-tray-open", open);
  }

  var ICONS = {
    edit: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z\"/></svg>",
    eye: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z\"/><circle cx=\"12\" cy=\"12\" r=\"3\"/></svg>",
    logout: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4\"/><path d=\"m16 17 5-5-5-5\"/><path d=\"M21 12H9\"/></svg>",
    check: "<svg class=\"plenti-check\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M20 6 9 17l-5-5\"/></svg>"
  };

  function icon(name) {
    var span = h("span", { class: "plenti-icon" }, []);
    span.innerHTML = ICONS[name];
    return span;
  }

  function makeTrayResizable(tray) {
    var saved = parseInt(localStorage.getItem("plenti_tray_width"), 10);
    if (saved) tray.style.width = saved + "px";
    var handle = h("div", { class: "plenti-tray-resize" }, []);
    handle.addEventListener("mousedown", function (e) {
      e.preventDefault();
      // Remember where on the handle the user grabbed so the edge stays
      // glued to the cursor (no jump on the first move).
      var grabOffset = e.clientX - tray.getBoundingClientRect().right;
      document.body.classList.add("plenti-tray-resizing");
      function onMove(ev) {
        var w = Math.min(640, Math.max(220, ev.clientX - grabOffset));
        tray.style.width = w + "px";
      }
      function onUp() {
        document.body.classList.remove("plenti-tray-resizing");
        localStorage.setItem("plenti_tray_width", String(parseInt(tray.style.width, 10) || 320));
        document.removeEventListener("mousemove", onMove);
        document.removeEventListener("mouseup", onUp);
      }
      document.addEventListener("mousemove", onMove);
      document.addEventListener("mouseup", onUp);
    });
    tray.appendChild(handle);
  }

  function mountAdminBar(el) {
    if (typeof el === "string") el = document.querySelector(el);
    if (!el) return;
    el.innerHTML = "";
    state.trayItem = currentPageItem();
    state.trayFields = JSON.parse(JSON.stringify(orderedFields(state.trayItem || { fields: {} })));
    var statusEl = h("span", { class: "plenti-admin-status" }, []);
    var bar = h("div", { class: "plenti-admin-bar" }, [
      h("span", { class: "plenti-admin-brand" }, ["Plenti"]),
      h("button", {
        class: "plenti-btn plenti-edit-toggle",
        onclick: function () { setTrayOpen(!state.trayOpen); }
      }, [icon("edit"), h("span", { class: "plenti-label" }, ["Edit"])]),
      statusEl,
      h("span", { class: "plenti-admin-spacer" }, []),
      h("button", {
        class: "plenti-btn plenti-logout",
        onclick: function () { logout(); }
      }, [icon("logout"), h("span", { class: "plenti-label" }, ["Logout"])])
    ]);
    var tray = h("div", { class: "plenti-admin-tray" }, []);
    makeTrayResizable(tray);
    el.appendChild(bar);
    el.appendChild(tray);
    document.body.classList.add("plenti-admin-active");
  }

  function boot() {
    var adminEl = document.getElementById("plenti-cms-app");
    var el = adminEl || document.querySelector("[data-plenti-cms]");
    loadData().then(function () {
      publishUser();
      // Only the dedicated admin page nags for a token; embedded instances
      // wait for an explicit user.login() call.
      if (adminEl && !isAuthenticated() && currentProvider() !== "local") login();
      if (adminEl) mount(adminEl);
      else if (el) mountAdminBar(el);
    }).catch(function (err) {
      if (el) el.textContent = "Could not load CMS data: " + err.message;
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
