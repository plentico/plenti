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
      .then(function (data) {
        state.data = data;
        // In local dev, /plenti-media lists files on disk so the library
        // reflects uploads made after the last build. Silently ignored
        // elsewhere (endpoint only exists under `plenti serve`).
        if (currentProvider() === "local") {
          return fetch("/plenti-media")
            .then(function (res) { return res.ok ? res.json() : []; })
            .then(function (live) {
              // Dedupe across sources: data.json uses "media/x" while
              // /plenti-media and uploads use "/media/x" — normalize.
              var merged = {};
              (data.media || []).concat(live || []).forEach(function (m) {
                merged["/" + m.replace(/^\/+/, "")] = true;
              });
              data.media = Object.keys(merged).sort();
              return data;
            })
            .catch(function () { return data; });
        }
        return data;
      });
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
    check: "<svg class=\"plenti-check\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M20 6 9 17l-5-5\"/></svg>",
    image: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><rect x=\"3\" y=\"3\" width=\"18\" height=\"18\" rx=\"2\"/><circle cx=\"9\" cy=\"9\" r=\"2\"/><path d=\"m21 15-3.1-3.1a2 2 0 0 0-2.8 0L6 21\"/></svg>",
    upload: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4\"/><path d=\"m17 8-5-5-5 5\"/><path d=\"M12 3v12\"/></svg>",
    download: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4\"/><path d=\"m7 10 5 5 5-5\"/><path d=\"M12 15V3\"/></svg>",
    trash: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M3 6h18\"/><path d=\"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6\"/><path d=\"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2\"/><path d=\"M10 11v6\"/><path d=\"M14 11v6\"/></svg>",
    link: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71\"/><path d=\"M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71\"/></svg>",
    crop: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M6 2v14a2 2 0 0 0 2 2h14\"/><path d=\"M18 22V8a2 2 0 0 0-2-2H2\"/></svg>",
    close: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M18 6 6 18\"/><path d=\"m6 6 12 12\"/></svg>"
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

  /* ---------- Media modal (embedded CMS) ---------- */

  function uploadMediaFiles(files, onStatus, onDone) {
    var queue = Array.prototype.slice.call(files);
    var uploaded = 0;
    function next() {
      if (!queue.length) { onDone(uploaded); return; }
      var file = queue.shift();
      onStatus("Uploading " + file.name + "...");
      var reader = new FileReader();
      reader.onload = function () {
        commit([{ file: "media/" + file.name, action: "upsert", encoding: "base64", contents: reader.result }])
          .then(function () {
            state.data.media = state.data.media || [];
            if (state.data.media.indexOf("/media/" + file.name) === -1) {
              state.data.media.push("/media/" + file.name);
            }
            uploaded++;
            next();
          })
          .catch(function (err) { onStatus("Error: " + err.message, true); onDone(uploaded); });
      };
      reader.onerror = function () { onStatus("Could not read " + file.name, true); onDone(uploaded); };
      reader.readAsDataURL(file);
    }
    next();
  }

  function renderMediaLibrary(grid, onPick, onDelete, onCopy, ver) {
    grid.innerHTML = "";
    var items = state.data.media || [];
    if (!items.length) {
      grid.appendChild(h("p", { class: "plenti-media-empty" }, ["No media yet — drop files above to upload."]));
      return;
    }
    items.forEach(function (m) {
      var full = state.data.env.baseurl.replace(/\/$/, "") + "/" + m.replace(/^\/+/, "") + (ver ? "?v=" + ver : "");
      var name = m.replace(/^\/+media\/+/, "");
      grid.appendChild(h("div", {
        class: "plenti-media-item",
        title: m,
        onclick: function () { onPick(m, full); }
      }, [
        isImage(m)
          ? h("img", { src: full, alt: m, loading: "lazy" })
          : h("div", { class: "plenti-media-doc" }, ["\uD83D\uDCC4"]),
        h("div", { class: "plenti-media-name" }, [name]),
        h("div", { class: "plenti-media-actions" }, [
          h("button", {
            class: "plenti-media-action",
            title: "Copy path",
            onclick: function (e) { e.stopPropagation(); onCopy(m); }
          }, [icon("link")]),
          h("a", {
            class: "plenti-media-action",
            href: full,
            download: name,
            title: "Download",
            onclick: function (e) { e.stopPropagation(); }
          }, [icon("download")]),
          h("button", {
            class: "plenti-media-action plenti-media-delete",
            title: "Delete",
            onclick: function (e) { e.stopPropagation(); onDelete(m, name); }
          }, [icon("trash")])
        ])
      ]));
    });
  }

  /* Custom styled confirm dialog (replaces window.confirm). */
  function confirmDialog(message, confirmLabel) {
    return new Promise(function (resolve) {
      var overlay = h("div", { class: "plenti-media-overlay plenti-confirm-overlay" }, [
        h("div", { class: "plenti-media-modal plenti-confirm-dialog" }, [
          h("p", { class: "plenti-confirm-msg" }, [message]),
          h("div", { class: "plenti-confirm-actions" }, [
            h("button", { class: "plenti-btn", onclick: function () { done(false); } }, ["Cancel"]),
            h("button", { class: "plenti-btn plenti-confirm-danger", onclick: function () { done(true); } }, [confirmLabel || "Delete"])
          ])
        ])
      ]);
      overlay.addEventListener("click", function (e) { if (e.target === overlay) done(false); });
      function done(val) {
        overlay.classList.remove("open");
        setTimeout(function () { overlay.remove(); }, 200);
        resolve(val);
      }
      document.body.appendChild(overlay);
      void overlay.offsetWidth;
      overlay.classList.add("open");
    });
  }

  /* ---------- Image editor modal ---------- */

  function openImageEditor(m, full, name, onSaved) {
    if (document.querySelector(".plenti-editor-overlay")) return;
    var statusEl = h("span", { class: "plenti-admin-status" }, []);
    var canvas = h("canvas", { class: "plenti-editor-canvas" }, []);
    var selection = h("div", { class: "plenti-editor-selection" }, []);
    var stage = h("div", { class: "plenti-editor-stage" }, [canvas, selection]);

    var img = new Image();
    var origCanvas = null, workCanvas = null;
    var scale = 100, format = (name.split(".").pop() || "png").toLowerCase();
    if (format === "jpg") format = "jpeg";
    if (["png", "jpeg", "webp"].indexOf(format) === -1) format = "png";
    var sel = null; // {x, y, w, h} in workCanvas coords

    function flash(msg, isError) {
      statusEl.classList.remove("success", "error");
      if (isError) { statusEl.textContent = "Error: " + msg; statusEl.classList.add("error"); return; }
      statusEl.innerHTML = ICONS.check + msg;
      statusEl.classList.add("success");
      setTimeout(function () { statusEl.textContent = ""; statusEl.classList.remove("success"); }, 2500);
    }

    function render() {
      var w = Math.max(1, Math.round(workCanvas.width * scale / 100));
      var hgt = Math.max(1, Math.round(workCanvas.height * scale / 100));
      canvas.width = w;
      canvas.height = hgt;
      canvas.getContext("2d").drawImage(workCanvas, 0, 0, w, hgt);
      clearSel();
    }
    function clearSel() {
      dsel = null;
      sel = null;
      selection.style.display = "none";
      cropBtn.disabled = true;
    }
    function sizeLabel() {
      return workCanvas.width + " \u00D7 " + workCanvas.height + " \u2192 " +
        Math.round(workCanvas.width * scale / 100) + " \u00D7 " + Math.round(workCanvas.height * scale / 100) + "px";
    }
    // Interactive crop box: draw on the canvas, then drag the box to move
    // it or pull one of the 8 handles to resize (coords tracked in display
    // space, converted to source pixels when applying).
    var dsel = null; // {x, y, w, h} relative to canvas top-left, in display px
    var HANDLE_DIRS = ["nw", "n", "ne", "e", "se", "s", "sw", "w"];
    HANDLE_DIRS.forEach(function (dir) {
      var hd = h("div", { class: "plenti-editor-handle plenti-h-" + dir }, []);
      hd.addEventListener("mousedown", function (e) {
        e.preventDefault();
        e.stopPropagation();
        beginDrag(dir, e);
      });
      selection.appendChild(hd);
    });
    selection.addEventListener("mousedown", function (e) {
      e.preventDefault();
      e.stopPropagation();
      beginDrag("move", e);
    });

    var mode = null, dragStart = null, boxStart = null;
    canvas.addEventListener("mousedown", function (e) {
      e.preventDefault();
      var rect = canvas.getBoundingClientRect();
      dsel = { x: e.clientX - rect.left, y: e.clientY - rect.top, w: 0, h: 0 };
      mode = "draw";
      dragStart = { x: dsel.x, y: dsel.y };
      showSel();
    });
    function beginDrag(m, e) {
      if (!dsel) return;
      mode = m;
      dragStart = { x: e.clientX, y: e.clientY };
      boxStart = { x: dsel.x, y: dsel.y, w: dsel.w, h: dsel.h };
    }
    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onUp);
    function onMove(e) {
      if (!mode) return;
      var rect = canvas.getBoundingClientRect();
      if (mode === "draw") {
        var x = Math.min(Math.max(e.clientX - rect.left, 0), rect.width);
        var y = Math.min(Math.max(e.clientY - rect.top, 0), rect.height);
        dsel = {
          x: Math.min(dragStart.x, x), y: Math.min(dragStart.y, y),
          w: Math.abs(x - dragStart.x), h: Math.abs(y - dragStart.y)
        };
      } else {
        var dx = e.clientX - dragStart.x, dy = e.clientY - dragStart.y;
        if (mode === "move") {
          dsel.x = Math.min(Math.max(boxStart.x + dx, 0), rect.width - boxStart.w);
          dsel.y = Math.min(Math.max(boxStart.y + dy, 0), rect.height - boxStart.h);
        } else {
          var nx = boxStart.x, ny = boxStart.y, nw = boxStart.w, nh = boxStart.h;
          if (mode.indexOf("w") !== -1) { nx = Math.min(boxStart.x + dx, boxStart.x + boxStart.w - 2); nw = boxStart.w - (nx - boxStart.x); }
          if (mode.indexOf("e") !== -1) { nw = Math.max(2, boxStart.w + dx); }
          if (mode.indexOf("n") !== -1) { ny = Math.min(boxStart.y + dy, boxStart.y + boxStart.h - 2); nh = boxStart.h - (ny - boxStart.y); }
          if (mode.indexOf("s") !== -1) { nh = Math.max(2, boxStart.h + dy); }
          nx = Math.max(0, nx); ny = Math.max(0, ny);
          nw = Math.min(nw, rect.width - nx); nh = Math.min(nh, rect.height - ny);
          dsel = { x: nx, y: ny, w: nw, h: nh };
        }
      }
      showSel();
    }
    function onUp() {
      if (mode === "draw" && dsel && (dsel.w < 3 || dsel.h < 3)) dsel = null;
      mode = null;
      showSel();
    }
    function showSel() {
      if (!dsel) { selection.style.display = "none"; syncSel(); return; }
      selection.style.display = "block";
      selection.style.left = (canvas.offsetLeft + dsel.x) + "px";
      selection.style.top = (canvas.offsetTop + dsel.y) + "px";
      selection.style.width = dsel.w + "px";
      selection.style.height = dsel.h + "px";
      syncSel();
    }
    function syncSel() {
      if (!dsel || dsel.w < 2 || dsel.h < 2) { sel = null; cropBtn.disabled = true; return; }
      var rect = canvas.getBoundingClientRect();
      var kx = workCanvas.width / rect.width, ky = workCanvas.height / rect.height;
      sel = { x: Math.round(dsel.x * kx), y: Math.round(dsel.y * ky), w: Math.round(dsel.w * kx), h: Math.round(dsel.h * ky) };
      cropBtn.disabled = false;
    }

    var cropBtn = h("button", { class: "plenti-btn", disabled: "disabled" }, [icon("crop"), h("span", { class: "plenti-label" }, ["Apply crop"])]);
    cropBtn.addEventListener("click", applyCrop);
    function applyCrop() {
      if (!sel || sel.w < 2 || sel.h < 2) return;
      var c = document.createElement("canvas");
      c.width = sel.w; c.height = sel.h;
      c.getContext("2d").drawImage(workCanvas, sel.x, sel.y, sel.w, sel.h, 0, 0, sel.w, sel.h);
      workCanvas = c;
      render();
      dimsEl.textContent = sizeLabel();
    }

    var scaleSlider = h("input", { type: "range", min: "10", max: "300", value: "100", class: "plenti-editor-scale" }, []);
    var scaleLabel = h("span", { class: "plenti-editor-scaleval" }, ["100%"]);
    scaleSlider.addEventListener("input", function () {
      scale = parseInt(scaleSlider.value, 10);
      scaleLabel.textContent = scale + "%";
      render();
      dimsEl.textContent = sizeLabel();
    });

    var formatSel = h("select", { class: "plenti-editor-format" }, ["png", "jpeg", "webp"].map(function (f) {
      return h("option", { value: f, selected: f === format ? "selected" : null }, [f === "jpeg" ? "jpg" : f]);
    }));
    formatSel.addEventListener("change", function () {
      format = formatSel.value;
      nameInput.value = nameInput.value.replace(/\.[a-z0-9]+$/i, "") + "." + (format === "jpeg" ? "jpg" : format);
    });

    var dimsEl = h("span", { class: "plenti-editor-dims" }, []);
    var nameInput = h("input", { type: "text", class: "plenti-input plenti-editor-name", value: name }, []);
    function exportDataURL() {
      var w = Math.max(1, Math.round(workCanvas.width * scale / 100));
      var hgt = Math.max(1, Math.round(workCanvas.height * scale / 100));
      var c = document.createElement("canvas");
      c.width = w; c.height = hgt;
      var ctx = c.getContext("2d");
      if (format === "jpeg") { ctx.fillStyle = "#ffffff"; ctx.fillRect(0, 0, w, hgt); }
      ctx.drawImage(workCanvas, 0, 0, w, hgt);
      return c.toDataURL("image/" + format, 0.92);
    }

    function save(targetName, thenDelete) {
      targetName = (targetName || "").trim().replace(/^\/+/, "").replace(/^media\/+/, "");
      if (!targetName) { flash("Filename required", true); return; }
      if (!/\.(png|jpe?g|webp)$/i.test(targetName)) {
        targetName += "." + (format === "jpeg" ? "jpg" : format);
      }
      statusEl.classList.remove("success", "error");
      statusEl.textContent = "Saving...";
      var ops = [{ file: "media/" + targetName, action: "upsert", encoding: "base64", contents: exportDataURL() }];
      commit(ops)
        .then(function () {
          if (thenDelete) return commit([{ file: "media/" + thenDelete, action: "delete" }]);
        })
        .then(function () {
          var entry = "/media/" + targetName;
          state.data.media = state.data.media || [];
          if (thenDelete) {
            var old = ("media/" + thenDelete).replace(/^\/+/, "").replace(/^media\/+/, "media/");
            state.data.media = state.data.media.filter(function (x) {
              return x.replace(/^\/+/, "") !== old;
            });
          }
          if (state.data.media.map(function (x) { return x.replace(/^\/+/, ""); }).indexOf(entry.replace(/^\/+/, "")) === -1) {
            state.data.media.push(entry);
            state.data.media.sort();
          }
          flash("Saved " + targetName);
          onSaved();
        })
        .catch(function (err) { flash(err.message, true); });
    }
    // Find a non-clobbering copy name: name-2.ext, name-3.ext, ...
    function copyName(base) {
      var m2 = base.match(/^(.*?)(\.[a-z0-9]+)$/i);
      var stem = m2 ? m2[1] : base, ext = m2 ? m2[2] : "";
      var taken = (state.data.media || []).map(function (x) { return x.replace(/^\/+/, ""); });
      var i = 2, cand;
      do { cand = stem + "-" + i + ext; i++; } while (taken.indexOf("media/" + cand) !== -1);
      return cand;
    }

    var overlay = h("div", {
      class: "plenti-media-overlay plenti-editor-overlay",
      onclick: function (e) { if (e.target === overlay) close(); }
    }, [
      h("div", { class: "plenti-media-modal plenti-editor-modal" }, [
        h("div", { class: "plenti-media-header" }, [
          h("h3", {}, [name]),
          dimsEl,
          statusEl,
          h("button", { class: "plenti-media-close", onclick: function () { close(); } }, ["×"])
        ]),
        stage,
        h("div", { class: "plenti-editor-toolbar" }, [
          cropBtn,
          h("button", { class: "plenti-btn", onclick: function () { workCanvas = origCanvas; scale = 100; scaleSlider.value = "100"; scaleLabel.textContent = "100%"; render(); dimsEl.textContent = sizeLabel(); } }, ["Reset"]),
          h("span", { class: "plenti-spacer" }, []),
          h("label", {}, ["Scale"]), scaleSlider, scaleLabel,
          h("label", {}, ["Format"]), formatSel
        ]),
        h("div", { class: "plenti-editor-savebar" }, [
          nameInput,
          h("button", { class: "plenti-btn", onclick: function () {
            var target = nameInput.value;
            // Never silently clobber the original from "Save as copy".
            if (target.trim().toLowerCase() === name.toLowerCase()) {
              target = copyName(nameInput.value);
              nameInput.value = target;
            }
            save(target, null);
          } }, ["Save as copy"]),
          h("button", { class: "plenti-btn primary", onclick: function () {
            var ext = format === "jpeg" ? "jpg" : format;
            var target = name.replace(/\.[a-z0-9]+$/i, "") + "." + ext;
            save(target, target.toLowerCase() === name.toLowerCase() ? null : name);
          } }, ["Overwrite original"])
        ])
      ])
    ]);
    function close() {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup", onUp);
      overlay.classList.remove("open");
      setTimeout(function () { overlay.remove(); }, 200);
    }
    img.onload = function () {
      origCanvas = document.createElement("canvas");
      origCanvas.width = img.naturalWidth;
      origCanvas.height = img.naturalHeight;
      origCanvas.getContext("2d").drawImage(img, 0, 0);
      workCanvas = origCanvas;
      render();
      dimsEl.textContent = sizeLabel();
    };
    img.onerror = function () { flash("Could not load image", true); };
    img.src = full;
    document.body.appendChild(overlay);
    void overlay.offsetWidth;
    overlay.classList.add("open");
  }

  function openMediaModal() {
    if (document.querySelector(".plenti-media-overlay")) return;
    var statusEl = h("span", { class: "plenti-admin-status" }, []);
    var grid = h("div", { class: "plenti-media-grid" }, []);
    var fileInput = h("input", { type: "file", multiple: "multiple", style: "display:none" }, []);

    var thumbVer = Date.now();
    function refresh() { renderMediaLibrary(grid, pick, del, copyPath, thumbVer); }
    function busted() { thumbVer = Date.now(); refresh(); }
    function copyPath(m) {
      if (!navigator.clipboard) return;
      navigator.clipboard.writeText(m).then(function () {
        statusEl.innerHTML = ICONS.check + "Copied " + m;
        statusEl.classList.add("success");
        setTimeout(function () { statusEl.textContent = ""; statusEl.classList.remove("success"); }, 2500);
      }).catch(function () { /* clipboard permission denied */ });
    }
    function pick(m, full) {
      var name = m.replace(/^\/+media\/+/, "");
      if (isImage(m)) openImageEditor(m, full, name, busted);
      else window.open(full, "_blank");
    }
    function del(m, name) {
      confirmDialog("Delete " + name + "? This cannot be undone.").then(function (ok) {
        if (!ok) return;
        statusEl.classList.remove("success", "error");
        commit([{ file: "media/" + name, action: "delete" }])
          .then(function () {
            state.data.media = (state.data.media || []).filter(function (x) { return x !== m; });
            busted();
            statusEl.innerHTML = ICONS.check + "Deleted " + name;
            statusEl.classList.add("success");
            setTimeout(function () { statusEl.textContent = ""; statusEl.classList.remove("success"); }, 2500);
          })
          .catch(function (err) {
            statusEl.textContent = "Error: " + err.message;
            statusEl.classList.add("error");
          });
      });
    }
    function doUpload(files) {
      if (!files || !files.length) return;
      statusEl.classList.remove("success", "error");
      uploadMediaFiles(files,
        function (msg, isError) { statusEl.textContent = msg; statusEl.classList.toggle("error", !!isError); },
        function (n) {
          if (n) {
            statusEl.innerHTML = ICONS.check + "Uploaded " + n + " file" + (n > 1 ? "s" : "");
            statusEl.classList.add("success");
            busted();
          }
        });
    }

    var dropzone = h("div", {
      class: "plenti-media-dropzone",
      onclick: function () { fileInput.click(); }
    }, [
      icon("upload"),
      h("p", {}, ["Drag & drop files here, or click to browse"])
    ]);
    ["dragenter", "dragover"].forEach(function (ev) {
      dropzone.addEventListener(ev, function (e) { e.preventDefault(); dropzone.classList.add("plenti-dragover"); });
    });
    ["dragleave", "drop"].forEach(function (ev) {
      dropzone.addEventListener(ev, function (e) { e.preventDefault(); dropzone.classList.remove("plenti-dragover"); });
    });
    dropzone.addEventListener("drop", function (e) { doUpload(e.dataTransfer.files); });
    fileInput.addEventListener("change", function () { doUpload(fileInput.files); fileInput.value = ""; });

    var overlay = h("div", {
      class: "plenti-media-overlay",
      onclick: function (e) { if (e.target === overlay) close(); }
    }, [
      h("div", { class: "plenti-media-modal" }, [
        h("div", { class: "plenti-media-header" }, [
          h("h3", {}, ["Media"]),
          statusEl,
          h("button", { class: "plenti-media-close", onclick: function () { close(); } }, ["\u00D7"])
        ]),
        dropzone,
        fileInput,
        grid
      ])
    ]);
    function close() {
      overlay.classList.remove("open");
      setTimeout(function () { overlay.remove(); }, 200);
    }
    document.body.appendChild(overlay);
    refresh();
    void overlay.offsetWidth; // reflow so the entrance transition runs
    overlay.classList.add("open");
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
      h("button", {
        class: "plenti-btn plenti-media-toggle",
        onclick: function () { openMediaModal(); }
      }, [icon("image"), h("span", { class: "plenti-label" }, ["Media"])]),
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
