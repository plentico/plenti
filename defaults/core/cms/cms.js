/* Plenti CMS - a dependency-free editor for Plenti content.
   Loads its data from /admin/data.json and saves through the configured
   provider (local dev server or a git host). */
(function () {
  "use strict";

  var IMAGE_EXTS = [".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg", ".avif", ".apng"];

  var state = {
    data: null,        // payload from data.json
    trayItem: null,    // data.content node for the page shown in the tray
    pendingFilename: null, // staged filename change, applied on Save
    trayFields: null,  // working copy of the tray's field values
    trayJsonMode: false, // tray shows raw JSON source instead of the form
    trayInfoMode: false, // tray shows page info instead of the form
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

  function isMediaPath(value) {
    return typeof value === "string" && /^\/?media\/.+\.(jpe?g|png|webp|gif|svg|avif|apng|pdf|msword)$/i.test(value);
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

  function fieldLabel(schema, key, path) {
    var options = fieldOptions(schema, key, path);
    return options.label || key;
  }

  function fieldOptions(schema, key, path) {
    var def = schema[path || key] || schema[key] || {};
    var options = def.options;
    return isPlainObject(options) ? options : (isPlainObject(def) ? def : {});
  }

  function fieldChoices(schema, key, path) {
    var options = fieldOptions(schema, key, path);
    return options.choices || [];
  }

  function referenceValue(item, result) {
    if (!result || result === "path") return item.path;
    if (result === "*") return item;
    var value = result.split(".").reduce(function (value, key) { return value == null ? null : value[key]; }, item);
    if (value === undefined && result.indexOf(".") === -1 && item.fields) value = item.fields[result];
    return value;
  }

  function referenceItems(schema, key, path, query) {
    var options = fieldOptions(schema, key, path);
    var types = options.types || [];
    var search = options.search || ["path"];
    query = (query || "").toLowerCase();
    return (state.data.content || []).filter(function (item) {
      if (item.pendingCreate || (types.length && types.indexOf(item.type) === -1)) return false;
      if (!query) return true;
      return search.some(function (field) {
        return String(referenceValue(item, field) || "").toLowerCase().indexOf(query) !== -1;
      });
    });
  }

  function referenceInput(key, value, onChange, schema, path, multiple) {
    var options = fieldOptions(schema, key, path);
    var values = multiple ? (Array.isArray(value) ? value : []) : [value];
    var search = h("input", { type: "search", placeholder: "Search content…", class: "plenti-reference-search" });
    var list = h("div", { class: "plenti-reference-results" });
    function matches(item) { return values.some(function (v) { return JSON.stringify(v) === JSON.stringify(referenceValue(item, options.result)); }); }
    function refresh() {
      list.innerHTML = "";
      var items = referenceItems(schema, key, path, search.value).slice(0, 50);
      if (!items.length) list.appendChild(h("div", { class: "plenti-help" }, ["No matching content"]));
      items.forEach(function (item) {
        var result = referenceValue(item, options.result);
        list.appendChild(h("label", { class: "plenti-reference-option" }, [
          h("input", { type: multiple ? "checkbox" : "radio", name: "reference-" + path, checked: matches(item), onchange: function () {
            if (multiple) {
              values = values.filter(function (v) { return JSON.stringify(v) !== JSON.stringify(result); });
              if (matches(item)) values.push(result);
              onChange(values);
            } else { values = [result]; onChange(result); }
            refresh();
          } }),
          h("span", {}, [(item.fields && (item.fields.title || item.fields.name)) || item.path])
        ]));
      });
    }
    search.addEventListener("input", refresh);
    refresh();
    return h("div", { class: "plenti-reference" }, [search, list]);
  }

  function richTextInput(value, onChange, options) {
    var toolbar = options.toolbar || "all";
    var commands = toolbar === "all" ? ["bold", "italic", "link", "heading2", "bullets"] : toolbar;
    var editor = h("div", { class: "plenti-rich-editor", contenteditable: "true" });
    editor.innerHTML = value || "";
    editor.addEventListener("input", function () { onChange(editor.innerHTML); });
    var buttons = commands.map(function (command) {
      var names = { bold: "B", italic: "I", link: "Link", heading2: "H2", bullets: "• List" };
      return h("button", { type: "button", class: "plenti-rich-command", title: command, onclick: function (e) {
        e.preventDefault(); editor.focus();
        if (command === "link") {
          var url = window.prompt("Link URL", "https://");
          if (url) document.execCommand("createLink", false, url);
        } else if (command === "heading2") document.execCommand("formatBlock", false, "h2");
        else if (command === "bullets") document.execCommand("insertUnorderedList", false, null);
        else document.execCommand(command, false, null);
        onChange(editor.innerHTML);
      } }, [names[command] || command]);
    });
    return h("div", { class: "plenti-rich-text" }, [h("div", { class: "plenti-rich-toolbar" }, buttons), editor]);
  }

  function componentInput(value, onChange, schema, path) {
    var allowed = fieldOptions(schema, path, path).allowed || [];
    var components = state.data.components || {};
    var componentSchemas = state.data.componentSchemas || {};
    var selected = value && value._component || "";
    var select = h("select", { disabled: fieldOptions(schema, path, path).disabled, onchange: function (e) { selected = e.target.value; var next = JSON.parse(JSON.stringify(components[selected] || {})); if (selected) next._component = selected; onChange(next); renderTray(document.querySelector(".plenti-admin-tray"), document.querySelector(".plenti-admin-status")); } }, [
      h("option", { value: "" }, ["Choose a component"])
    ].concat(allowed.map(function (name) { return h("option", { value: name, selected: name === selected ? "selected" : null }, [name]); })));
    var values = value || {};
    var componentSchema = componentSchemas[selected] || {};
    var body = h("div", { class: "plenti-component-fields" });
    Object.keys(values).forEach(function (key) {
      if (key === "_component") return;
      body.appendChild(h("div", { class: "plenti-field" }, [h("label", { hidden: schemaType(componentSchema, key, values[key], key) === "group" ? "hidden" : null }, [fieldLabel(componentSchema, key)]), fieldInput(key, schemaType(componentSchema, key, values[key], key), values[key], function (next) { values[key] = next; onChange(values); }, componentSchema, key)]));
    });
    return h("div", { class: "plenti-component" }, [select, body]);
  }

  function transformMedia(selected, options) {
    if (!isImage(selected) || (!options.width && !options.height && !options.convert)) return Promise.resolve(selected);
    return new Promise(function (resolve, reject) {
      var img = new Image();
      img.onload = function () {
        var width = Number(options.width) || img.naturalWidth;
        var height = Number(options.height) || img.naturalHeight;
        if (!options.width) width = Math.round(img.naturalWidth * height / img.naturalHeight);
        if (!options.height) height = Math.round(img.naturalHeight * width / img.naturalWidth);
        var canvas = document.createElement("canvas");
        canvas.width = width; canvas.height = height;
        var ctx = canvas.getContext("2d");
        var sw = img.naturalWidth, sh = img.naturalHeight, sx = 0, sy = 0;
        if (options.crop) {
          var scale = Math.max(width / sw, height / sh);
          sw = width / scale; sh = height / scale;
          sx = (img.naturalWidth - sw) / 2; sy = (img.naturalHeight - sh) / 2;
        }
        ctx.drawImage(img, sx, sy, sw, sh, 0, 0, width, height);
        var format = (options.convert || selected.split(".").pop()).toLowerCase().replace("jpg", "jpeg");
        var name = selected.replace(/^\/+/, "").replace(/\.[^.]+$/, "") + "." + (format === "jpeg" ? "jpg" : format);
        commit([{ file: name, action: "upsert", encoding: "base64", contents: canvas.toDataURL("image/" + format, Number(options.quality) || 0.82) }]).then(function () {
          state.data.media = state.data.media || [];
          if (state.data.media.indexOf("/" + name) === -1) state.data.media.push("/" + name);
          resolve("/" + name);
        }).catch(reject);
      };
      img.onerror = function () { reject(new Error("Could not transform image")); };
      img.src = state.data.env.baseurl.replace(/\/$/, "") + "/" + selected.replace(/^\/+/, "");
    });
  }

  function fieldInput(key, type, value, onChange, schema, path) {
    schema = schema || {};
    path = path || key;
    var options = fieldOptions(schema, key, path);
    var common = {
      disabled: options.disabled === true || type === "id" ? "disabled" : null,
      required: options.required === true ? "required" : null
    };
    if (type === "rich_text") return richTextInput(value, onChange, options);
    if (type === "reference" || type === "references") return referenceInput(key, value, onChange, schema, path, type === "references");
    if (type === "component") return componentInput(value, onChange, schema, path);
    if (type === "list_text") type = "list";

    if (type === "boolean") {
      return h("input", { type: "checkbox", checked: !!value, disabled: common.disabled, required: common.required, onchange: function (e) { onChange(e.target.checked); } });
    }
    if (type === "number") {
      return h("input", { type: "number", disabled: common.disabled, required: common.required, value: value == null ? "" : value, onchange: function (e) { onChange(e.target.value === "" ? null : Number(e.target.value)); } });
    }
    if (type === "date") {
      return h("input", { type: "date", disabled: common.disabled, required: common.required, value: value || "", onchange: function (e) { onChange(e.target.value); } });
    }
    if (type === "time") {
      return h("input", { type: "time", disabled: common.disabled, required: common.required, value: value || "", onchange: function (e) { onChange(e.target.value); } });
    }
    if ((type === "select" || type === "radio" || type === "checkbox")) {
      var choices = fieldChoices(schema, key, path);
      if (type === "select") {
        return h("select", { disabled: common.disabled, required: common.required, onchange: function (e) { onChange(e.target.value); } },
          choices.map(function (opt) {
            return h("option", { value: opt, selected: opt === value ? "selected" : null }, [opt]);
          }));
      }
      return h("div", { class: "plenti-checks" }, choices.map(function (opt) {
        var input = type === "radio"
          ? h("input", { type: "radio", name: key, disabled: common.disabled, required: common.required, value: opt, checked: opt === value, onchange: function () { onChange(opt); } })
          : h("input", { type: "checkbox", value: opt, disabled: common.disabled, required: common.required, checked: Array.isArray(value) && value.indexOf(opt) !== -1,
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
      var preview = h("div", { class: "plenti-media-preview" }, []);
      function renderPreview() {
        preview.innerHTML = "";
        var full = state.data.env.baseurl.replace(/\/$/, "") + "/" + String(value || "").replace(/^\/+/, "");
        preview.appendChild(h("button", {
          type: "button", class: "plenti-media-field", title: "Change media", disabled: common.disabled,
          onclick: function () {
            openMediaModal(function (selected) {
              transformMedia(selected, options).then(function (result) {
                value = result;
                onChange(result);
                renderPreview();
              }).catch(function (err) { notify(err.message, true); });
            });
          }
        }, [
          isImage(value || "")
            ? h("img", { src: full, alt: "Click to change media" })
            : h("span", { class: "plenti-media-doc" }, ["\uD83D\uDCC4"]),
          h("span", { class: "plenti-media-field-name" }, [value || "Choose media"])
        ]));
      }
      renderPreview();
      return preview;
    }
    if (type === "object" || type === "group") {
      var fieldset = h("fieldset", { class: "plenti-fieldset" }, []);
      Object.keys(value || {}).forEach(function (nestedKey) {
        var nestedValue = value[nestedKey];
        var nestedPath = path + "." + nestedKey;
        fieldset.appendChild(h("div", { class: "plenti-field" }, [
          schemaType(schema, nestedKey, nestedValue, nestedPath) === "group" ? null : h("label", {}, [fieldLabel(schema, nestedKey, nestedPath)]),
          fieldInput(nestedKey, schemaType(schema, nestedKey, nestedValue, nestedPath), nestedValue, function (nested) {
            value[nestedKey] = nested;
            onChange(value);
          }, schema, nestedPath)
        ]));
      });
      var groupOpen = options.expand === true;
      var groupChevron = h("button", {
        type: "button",
        class: "plenti-array-chevron" + (groupOpen ? " plenti-open" : ""),
        "aria-label": "Toggle " + fieldLabel(schema, key, path)
      }, [icon("chevron")]);
      fieldset.style.display = groupOpen ? "" : "none";
      var groupHeader = h("div", { class: "plenti-group-header", onclick: function () {
        groupOpen = !groupOpen;
        groupChevron.classList.toggle("plenti-open", groupOpen);
        slideToggle(fieldset, groupOpen);
      } }, [
        groupChevron,
        h("label", {}, [fieldLabel(schema, key, path)]),
        h("span", { class: "plenti-group-count" }, [Object.keys(value || {}).length + " field" + (Object.keys(value || {}).length === 1 ? "" : "s")])
      ]);
      return h("div", { class: "plenti-group" }, [groupHeader, fieldset]);
    }
    if (type === "list") {
      return h("input", { type: "text", value: Array.isArray(value) ? value.join(", ") : (value || ""),
        onchange: function (e) {
          onChange(e.target.value.split(",").map(function (s) { return s.trim(); }).filter(Boolean));
        } });
    }
    if (type === "textarea") {
      return h("textarea", { disabled: common.disabled, required: common.required, oninput: function (e) { onChange(e.target.value); } }, [value || ""]);
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
    return h("input", { type: "text", value: value == null ? "" : value, disabled: common.disabled, required: common.required, oninput: function (e) { onChange(e.target.value); } });
  }

  function schemaType(schema, key, value, path) {
    var def = schema[path || key] || schema[key];
    if (typeof def === "string") return def;
    if (def && def.type) return def.type;
    if (typeof value === "boolean") return "boolean";
    if (typeof value === "number") return "number";
    if (Array.isArray(value)) return "list";
    if (isPlainObject(value)) return "object";
    if (isMediaPath(value)) return "media";
    return "text";
  }

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
    syncAdminBar();
    // Pattr may insert the embedded container after updating its root data.
    setTimeout(syncAdminBar, 0);
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

  /* ---------- Inline admin bar (embedded CMS) ---------- */

  // Find the content node backing the page currently shown in the browser.
  function currentPageItem() {
    if (state.trayItem && state.trayItem.pendingCreate) return state.trayItem;
    var path = window.location.pathname.replace(/\/+$/, "") || "/";
    var items = state.data.content || [];
    for (var i = 0; i < items.length; i++) {
      if ((items[i].path || "").replace(/\/+$/, "") === path || items[i].path === path) return items[i];
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
      // Reactive body p-class bindings may replace CMS-managed classes while
      // applying this content update. Restore them after Pattr's deferred DOM
      // walk, including the admin bar's conditional mount when needed.
      setTimeout(syncAdminBar, 0);
    }
  }

  /* ---------- Tray JSON source view ---------- */

  function escapeHtml(s) {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }

  // Regex tokenizer for JSON: keys, strings, numbers, booleans/null.
  // Input is HTML-escaped first; matches are escaped per-token already.
  function highlightJson(src) {
    var esc = escapeHtml(src);
    return esc.replace(/("(\\u[a-fA-F0-9]{4}|\\[^u]|[^\\"])*")(\s*:)?|\b(true|false|null)\b|-?\d+(\.\d+)?([eE][+-]?\d+)?/g,
      function (m, str, _esc, colon, bool) {
        var cls = bool ? "bool" : (str ? (colon ? "key" : "str") : "num");
        return "<span class=\"plenti-tok-" + cls + "\">" + m + "</span>";
      });
  }

  // Turn a JSON.parse error into a line/column for the error bar.
  function jsonErrorLocation(text, err) {
    var m = /position (\d+)/.exec(err.message);
    if (m) {
      var pos = parseInt(m[1], 10);
      var line = 1, col = 1;
      for (var i = 0; i < pos && i < text.length; i++) {
        if (text[i] === "\n") { line++; col = 1; } else col++;
      }
      return "line " + line + ", col " + col;
    }
    var lc = /line (\d+) column (\d+)/.exec(err.message);
    if (lc) return "line " + lc[1] + ", col " + lc[2];
    return null;
  }

  function renderTrayJson(tray, statusEl, saveBtn) {
    var code = h("code", {}, []);
    var pre = h("pre", { class: "plenti-json-highlight" }, [code]);
    var input = h("textarea", {
      class: "plenti-json-input",
      spellcheck: "false",
      wrap: "off",
      oninput: validate,
      onscroll: function () {
        pre.scrollTop = input.scrollTop;
        pre.scrollLeft = input.scrollLeft;
      }
    }, []);
    input.value = JSON.stringify(state.trayFields, null, 2);
    var errEl = h("div", { class: "plenti-json-error", style: "display:none" }, []);
    var wrap = h("div", { class: "plenti-json-wrap" }, [pre, input]);
    tray.appendChild(wrap);
    tray.appendChild(errEl);

    function validate() {
      var text = input.value;
      code.innerHTML = highlightJson(text) + "\n";
      try {
        state.trayFields = JSON.parse(text);
        errEl.style.display = "none";
        wrap.classList.remove("plenti-json-invalid");
        if (saveBtn) saveBtn.disabled = false;
        // Live-update the page from valid JSON edits.
        Object.keys(state.trayFields).forEach(function (key) {
          liveUpdate(key, state.trayFields[key]);
        });
      } catch (err) {
        var loc = jsonErrorLocation(text, err);
        errEl.textContent = "Invalid JSON" + (loc ? " (" + loc + ")" : "") + ": " + err.message;
        errEl.style.display = "block";
        wrap.classList.add("plenti-json-invalid");
        if (saveBtn) saveBtn.disabled = true;
      }
    }
    validate();
  }

  /* ---------- Tray: generic arrays (repeatable groups) ---------- */

  function isPlainObject(v) {
    return !!v && typeof v === "object" && !Array.isArray(v);
  }

  // Repeatable groups can contain either objects or nested groups (rows,
  // columns, etc.). Scalar arrays remain simple comma-separated fields.
  function isStructuredArray(v) {
    return Array.isArray(v) && v.every(function (item) {
      return isPlainObject(item) || Array.isArray(item);
    });
  }

  // Label for an array item: first non-empty string value, else "Item N".
  // Purely cosmetic — no key carries special meaning.
  function arrayItemLabel(item, i) {
    var keys = Object.keys(item);
    for (var n = 0; n < keys.length; n++) {
      var v = item[keys[n]];
      if (typeof v === "string" && v.trim()) {
        return v.length > 28 ? v.slice(0, 28) + "…" : v;
      }
    }
    return "Item " + (i + 1);
  }

  // Array groups are expanded by default; item cards remain collapsed.
  function arrayGroupOpen(path, schema) {
    state.trayArrayOpen = state.trayArrayOpen || {};
    if (path in state.trayArrayOpen) return state.trayArrayOpen[path];
    var key = path.slice(path.lastIndexOf(".") + 1);
    return fieldOptions(schema, key, path).expand === true;
  }
  function arrayItemOpen(path) {
    state.trayItemOpen = state.trayItemOpen || {};
    return !!state.trayItemOpen[path];
  }

  // Renders one value inside an array item. Objects render as a labeled group
  // with an indented body; arrays of objects recurse as collapsible groups.
  function renderArrayItemField(tray, statusEl, path, item, key, commit, schema, onEdit) {
    var v = item[key];
    if (isStructuredArray(v)) {
      return renderArrayField(tray, statusEl, path + "." + key, key, v, commit, schema);
    }
    if (isPlainObject(v)) {
      var body = h("div", { class: "plenti-nested-body" }, []);
      Object.keys(v).forEach(function (nk) {
        body.appendChild(renderArrayItemField(tray, statusEl, path + "." + key, v, nk, commit, schema, onEdit));
      });
      return h("div", { class: "plenti-field plenti-field-nested" }, [
        h("label", {}, [fieldLabel(schema, key, path + "." + key)]),
        body
      ]);
    }
    var type = schemaType(schema, key, v);
    return h("div", { class: "plenti-field" }, [
      type === "object" || type === "group" ? null : h("label", {}, [fieldLabel(schema, key)]),
      fieldInput(key, type, v, function (val) {
        item[key] = val; commit(); onEdit();
      }, schema)
    ]);
  }

  // Animate expand/collapse by transitioning height between 0 and the
  // element's natural height (in-place; no re-render needed).
  function slideToggle(el, show) {
    if (el._slideAnim) return;
    if (show) {
      el.style.display = "";
      var ht = el.scrollHeight;
      if (!ht) return;
      el._slideAnim = true;
      el.style.overflow = "hidden";
      el.style.height = "0px";
      requestAnimationFrame(function () {
        el.style.transition = "height .2s ease";
        el.style.height = ht + "px";
        el.addEventListener("transitionend", function done() {
          el.removeEventListener("transitionend", done);
          el.style.transition = ""; el.style.height = ""; el.style.overflow = "";
          el._slideAnim = false;
        });
      });
    } else {
      var ht2 = el.scrollHeight;
      if (!ht2) { el.style.display = "none"; return; }
      el._slideAnim = true;
      el.style.overflow = "hidden";
      el.style.height = ht2 + "px";
      requestAnimationFrame(function () {
        el.style.transition = "height .2s ease";
        el.style.height = "0px";
        el.addEventListener("transitionend", function done() {
          el.removeEventListener("transitionend", done);
          el.style.display = "none";
          el.style.transition = ""; el.style.height = ""; el.style.overflow = "";
          el._slideAnim = false;
        });
      });
    }
  }

  // An array of objects = repeatable group. Top-level groups start expanded,
  // nested groups and individual items start collapsed. Items can be
  // reordered via up/down arrows or drag-and-drop, and removed via trash.
  function renderArrayField(tray, statusEl, path, label, arr, commit, schema, itemsOnly) {
    var open = arrayGroupOpen(path, schema);
    var chevron = h("button", {
      class: "plenti-array-chevron" + (open ? " plenti-open" : ""),
      title: open ? "Collapse" : "Expand",
      onclick: function () { toggleGroup(); }
    }, [icon("chevron")]);
    var body = h("div", { class: "plenti-array-items", style: open ? "" : "display:none" }, []);
    // Items are always rendered (hidden when collapsed) so toggles can
    // animate in place without re-rendering the whole tray.
    function toggleGroup() {
      open = !open;
      state.trayArrayOpen[path] = open;
      chevron.classList.toggle("plenti-open", open);
      chevron.title = open ? "Collapse" : "Expand";
      slideToggle(body, open);
    }
    // FLIP animation helper: records each item card's position before a
    // mutation, re-renders, then slides cards from old to new positions.
    function flip(mutate) {
      var cards = body.querySelectorAll(":scope > .plenti-array-item");
      var before = arr.map(function (it, j) {
        return { item: it, top: cards[j] ? cards[j].getBoundingClientRect().top : 0 };
      });
      mutate();
      renderTray(tray, statusEl);
      var grp = tray.querySelector('.plenti-array[data-path="' + path + '"]');
      if (!grp) return;
      var newCards = grp.querySelectorAll(":scope > .plenti-array-items > .plenti-array-item");
      arr.forEach(function (it, j) {
        var oldTop = null;
        for (var k = 0; k < before.length; k++) if (before[k].item === it) oldTop = before[k].top;
        if (oldTop == null || !newCards[j]) return;
        var delta = oldTop - newCards[j].getBoundingClientRect().top;
        if (!delta) return;
        var c = newCards[j];
        c.style.transition = "none";
        c.style.transform = "translateY(" + delta + "px)";
        requestAnimationFrame(function () {
          requestAnimationFrame(function () {
            c.style.transition = "transform .2s ease";
            c.style.transform = "";
          });
        });
      });
    }
    // Drag state shared by the items in this group. Reordering happens LIVE
    // during dragover by physically moving card DOM nodes (the dragged node
    // itself is never removed, so the drag session survives); the array is
    // kept in sync and committed on each move. dragend just cleans up.
    var dragFrom = -1;
    var dragCard = null;
    var justDragged = false;
    function itemCards() {
      return Array.prototype.slice.call(body.querySelectorAll(":scope > .plenti-array-item"));
    }
    function slideIn(el, delta) {
      if (!delta) return;
      el.style.transition = "none";
      el.style.transform = "translateY(" + delta + "px)";
      requestAnimationFrame(function () {
        requestAnimationFrame(function () {
          el.style.transition = "transform .18s ease";
          el.style.transform = "";
        });
      });
    }
    if (!arr.length) body.appendChild(h("p", { class: "plenti-admin-empty" }, ["Empty."]));
    arr.forEach(function (item, i) {
        var itemPath = path + "." + i;
        var itemOpen = arrayItemOpen(itemPath);
        function rerender() { renderTray(tray, statusEl); }
        function move(to) {
          flip(function () {
            var x = arr.splice(i, 1)[0];
            arr.splice(to, 0, x);
            commit();
          });
        }
        var labelEl = h("span", { class: "plenti-array-item-label" }, [arrayItemLabel(item, i)]);
        function onEdit() { labelEl.textContent = arrayItemLabel(item, i); }
        var itemChevron = h("button", {
          class: "plenti-array-chevron plenti-sm" + (itemOpen ? " plenti-open" : ""),
          title: itemOpen ? "Collapse" : "Expand",
          onclick: function () { toggleItem(); }
        }, [icon("chevron")]);
        var fieldsEl = h("div", {
          class: "plenti-array-item-fields",
          style: itemOpen ? "" : "display:none"
        }, []);
        if (Array.isArray(item)) {
          // An array inside an array renders its cards directly inside the
          // existing outer card, without a second group header.
          fieldsEl.appendChild(renderArrayField(tray, statusEl, itemPath, "Items", item, commit, schema, true));
        } else {
          Object.keys(item).forEach(function (key) {
            fieldsEl.appendChild(renderArrayItemField(tray, statusEl, itemPath, item, key, commit, schema, onEdit));
          });
        }
        // Toggle in place (no re-render) so the fields slide open/closed.
        function toggleItem() {
          itemOpen = !itemOpen;
          state.trayItemOpen[itemPath] = itemOpen;
          itemChevron.classList.toggle("plenti-open", itemOpen);
          itemChevron.title = itemOpen ? "Collapse" : "Expand";
          slideToggle(fieldsEl, itemOpen);
        }
        // Drag-and-drop reordering (drag by the item header).
        var card = h("div", { class: "plenti-array-item" }, []);
        var header = h("div", { class: "plenti-array-item-header", draggable: "true" }, [
          itemChevron,
          h("span", { class: "plenti-array-grip", title: "Drag to reorder" }, [icon("grip")]),
          labelEl,
          h("span", { class: "plenti-spacer" }, []),
          h("button", {
            class: "plenti-array-move", title: "Move up",
            disabled: i === 0 ? "disabled" : null,
            onclick: function () { move(i - 1); }
          }, [icon("up")]),
          h("button", {
            class: "plenti-array-move", title: "Move down",
            disabled: i === arr.length - 1 ? "disabled" : null,
            onclick: function () { move(i + 1); }
          }, [icon("down")]),
          h("button", {
            class: "plenti-array-move plenti-array-remove", title: "Remove",
            onclick: function () {
              flip(function () {
                arr.splice(i, 1);
                commit();
              });
            }
          }, [icon("trash")])
        ]);
        // Click the empty middle of the header to toggle expand/collapse.
        // (A real drag suppresses the click; justDragged covers edge cases.)
        header.addEventListener("click", function (e) {
          if (justDragged) return;
          if (e.target.closest("button") || e.target.closest(".plenti-array-grip")) return;
          toggleItem();
        });
        header.addEventListener("dragstart", function (e) {
          e.dataTransfer.effectAllowed = "move";
          e.dataTransfer.setData("text/plain", String(i));
          dragFrom = i;
          dragCard = card;
          card.classList.add("plenti-dragging");
        });
        header.addEventListener("dragend", function () {
          card.classList.remove("plenti-dragging");
          dragFrom = -1;
          dragCard = null;
          justDragged = true;
          setTimeout(function () { justDragged = false; }, 100);
          // Array was already reordered live; re-render to sync labels
          // and arrow disabled states with the new positions.
          rerender();
        });
        card.addEventListener("dragover", function (e) {
          if (dragFrom < 0 || !dragCard || card === dragCard) return;
          e.preventDefault();
          e.dataTransfer.dropEffect = "move";
          // Current position of the hovered card (closure i is stale once
          // cards start moving, so compute it live).
          var cards = itemCards();
          var to = cards.indexOf(card);
          if (to < 0 || to === dragFrom) return;
          // Only swap once the cursor crosses the target card's midpoint in
          // the direction of travel — otherwise the displaced card slides
          // back under the cursor and immediately re-triggers a swap
          // (visible as the item "jumping back" when dragging upward).
          var rect = card.getBoundingClientRect();
          var mid = rect.top + rect.height / 2;
          if (to > dragFrom && e.clientY < mid) return;
          if (to < dragFrom && e.clientY > mid) return;
          // Swap preview: physically move the dragged card past this one
          // and slide this card into the vacated slot.
          var top = card.getBoundingClientRect().top;
          if (to > dragFrom) body.insertBefore(dragCard, card.nextSibling);
          else body.insertBefore(dragCard, card);
          slideIn(card, top - card.getBoundingClientRect().top);
          var x = arr.splice(dragFrom, 1)[0];
          arr.splice(to, 0, x);
          dragFrom = to;
          commit();
        });
        card.appendChild(header);
        card.appendChild(fieldsEl);
        body.appendChild(card);
      });
    // Allow dropping anywhere over the tray; the actual reorder already
    // happened live during dragover, so drop just needs to not be rejected.
    tray.addEventListener("dragover", function (e) {
      if (dragFrom < 0) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = "move";
    });
    tray.addEventListener("drop", function (e) {
      if (dragFrom < 0) return;
      e.preventDefault();
    });
    // Array-valued items already have a collapsible outer card. Only render
    // their nested cards, without a redundant nested group header/wrapper.
    if (itemsOnly) {
      body.style.display = "";
      return body;
    }
    label = fieldLabel(schema, label, path);
    return h("div", { class: "plenti-array", "data-path": path }, [
      h("div", { class: "plenti-array-header", onclick: function (e) {
        if (e.target.closest(".plenti-array-chevron")) return;
        chevron.click();
      } }, [
        chevron,
        h("label", {}, [label]),
        h("span", { class: "plenti-array-count" }, [arr.length + " item" + (arr.length === 1 ? "" : "s")])
      ]),
      body
    ]);
  }

  function renderTray(tray, statusEl) {
    // Clear previous content but keep the drag-resize handle in place.
    var resizeHandle = tray.querySelector(".plenti-tray-resize");
    tray.innerHTML = "";
    if (resizeHandle) tray.appendChild(resizeHandle);
    // Mark as re-rendered so field entrance animations only play once
    // (on tray open), not on every in-place re-render (reorders etc).
    tray.classList.add("plenti-rendered");
    var item = state.trayItem;
    if (!item) {
      tray.appendChild(h("p", { class: "plenti-admin-empty" }, ["No content source found for this page."]));
      return;
    }
    var schema = (state.data.schemas && state.data.schemas[item.type]) || {};
    var saveBtn = h("button", {
      class: "plenti-btn primary",
      onclick: function () { saveTray(statusEl); }
    }, ["Save"]);
    var mode = state.trayJsonMode ? "json" : state.trayInfoMode ? "info" : "form";
    var header = h("div", { class: "plenti-tray-header" }, [
      h("h3", {}, [mode === "json" ? "Code" : mode === "info" ? "Info" : "Form"]),
      h("div", { class: "plenti-tray-tabs" }, [
        h("button", {
          class: "plenti-json-toggle" + (mode === "form" ? " plenti-active" : ""),
          title: "Form",
          onclick: function () { setTrayMode(tray, statusEl, "form"); }
        }, [icon("form")]),
        h("button", {
          class: "plenti-json-toggle" + (mode === "info" ? " plenti-active" : ""),
          title: "Page info",
          onclick: function () { setTrayMode(tray, statusEl, "info"); }
        }, [icon("info")]),
        h("button", {
          class: "plenti-json-toggle" + (mode === "json" ? " plenti-active" : ""),
          title: "JSON source",
          onclick: function () { setTrayMode(tray, statusEl, "json"); }
        }, [icon("braces")])
      ])
    ]);
    tray.appendChild(header);
    if (state.trayJsonMode) {
      renderTrayJson(tray, statusEl, saveBtn);
    } else if (state.trayInfoMode) {
      renderTrayInfo(tray, item, statusEl);
    } else {
      Object.keys(state.trayFields).forEach(function (key) {
        var value = state.trayFields[key];
        if (fieldOptions(schema, key).hidden) return;
        if (isStructuredArray(value) && schemaType(schema, key, value) !== "references") {
          tray.appendChild(renderArrayField(tray, statusEl, key, key, value, function () {
            liveUpdate(key, value);
          }, schema));
          return;
        }
        var type = schemaType(schema, key, state.trayFields[key]);
        tray.appendChild(h("div", { class: "plenti-field" }, [
          type === "object" || type === "group" ? null : h("label", {}, [fieldLabel(schema, key)]),
          fieldInput(key, type, state.trayFields[key], function (v) { liveUpdate(key, v); }, schema)
        ]));
      });
    }
    tray.appendChild(h("div", { class: "plenti-admin-actions" }, [
      h("button", {
        class: "plenti-btn danger",
        title: "Delete this page",
        onclick: function () { deleteTrayPage(statusEl); }
      }, [icon("trash"), h("span", {}, ["Delete"])]),
      saveBtn,
      h("button", {
        class: "plenti-btn",
        onclick: function () { setTrayOpen(false); }
      }, ["Close"])
    ]));
  }

  // Delete the page shown in the tray: custom confirm modal, then a delete
  // commit (skipped for stubs that were never saved), local index cleanup,
  // and a redirect home since the current URL no longer exists.
  function deleteTrayPage(statusEl) {
    var item = state.trayItem;
    if (!item) return;
    confirmModal(
      "Delete this page?",
      [
        "This permanently deletes ",
        h("code", { class: "plenti-inline-code" }, ["content/" + item.filepath]),
        ". This cannot be undone."
      ],
      "Delete",
      true
    ).then(function (ok) {
      if (!ok) return;
      statusEl.textContent = "Deleting...";
      statusEl.classList.remove("success", "error");
      var finish = function () {
        sessionStorage.removeItem("plenti_cms_draft");
        var idx = state.data.content.indexOf(item);
        if (idx >= 0) state.data.content.splice(idx, 1);
        if (state.trayItem === item) state.trayItem = null;
        setTrayOpen(false);
        statusEl.innerHTML = ICONS.check + "Deleted";
        statusEl.classList.add("success");
        setTimeout(function () {
          statusEl.textContent = "";
          statusEl.classList.remove("success");
        }, 3000);
        setTimeout(function () { window.location.href = "/"; }, 400);
      };
      // A stubbed new page was never written to disk, so there's nothing to commit.
      if (item.pendingCreate) { finish(); return; }
      commit([{ file: "content/" + item.filepath, action: "delete" }])
        .then(finish)
        .catch(function (err) {
          statusEl.textContent = "Error: " + err.message;
          statusEl.classList.add("error");
        });
    });
  }

  // Mirror of cmd/build/data_source.go path resolution: route overrides from
  // plenti.json (":fields(name)" tokens replaced with slugified field values),
  // else the default type-folder/file-name path, with "_index" files mapping
  // to the root of their folder.
  function computePath(type, relPathNoExt, fields) {
    var slugify = function (s) { return String(s).toLowerCase().replace(/[^a-z0-9./*]+/g, "-"); };
    var routes = (state.data.env && state.data.env.routes) || {};
    var path = null;
    var route = routes[type];
    if (route != null) {
      path = route.replace(/:fields\(([^)]*)\)/g, function (m, name) {
        var v = fields ? fields[name] : null;
        return slugify(v == null ? "" : v);
      }).replace(/:paginate\([^)]*\)/g, "").replace(/\/$/, "");
    }
    if (path == null) {
      var parts = relPathNoExt.split("/");
      if (parts[parts.length - 1] === "_index") {
        path = "/" + parts.slice(0, -1).join("/");
      } else {
        path = parts.length > 1 ? "/" + parts.slice(1).join("/") : "/" + parts[0];
      }
    }
    return path;
  }

  // Generic custom modal confirm (never the browser default), styled like the
  // media library delete confirm. msgChildren is an array of nodes/strings so
  // paths can be rendered as inline code. Resolves true on OK, false on
  // Cancel/Escape/backdrop click.
  function confirmModal(title, msgChildren, okLabel, danger) {
    return new Promise(function (resolve) {
      var overlay, onKey;
      function done(val) {
        document.removeEventListener("keydown", onKey, true);
        overlay.classList.remove("open");
        setTimeout(function () { overlay.remove(); }, 200);
        resolve(val);
      }
      onKey = function (e) {
        if (e.key === "Escape") { e.stopPropagation(); done(false); }
        if (e.key === "Enter") { done(true); }
      };
      var okBtn = h("button", { class: danger ? "plenti-btn plenti-confirm-danger" : "plenti-btn primary", onclick: function () { done(true); } }, [okLabel || "OK"]);
      overlay = h("div", { class: "plenti-media-overlay plenti-confirm-overlay", onclick: function (e) { if (e.target === overlay) done(false); } }, [
        h("div", { class: "plenti-media-modal plenti-confirm-dialog" }, [
          h("p", { class: "plenti-confirm-msg" }, [h("strong", {}, [title])].concat([" "]).concat(msgChildren)),
          h("div", { class: "plenti-confirm-actions" }, [
            h("button", { class: "plenti-btn", onclick: function () { done(false); } }, ["Cancel"]),
            okBtn
          ])
        ])
      ]);
      document.body.appendChild(overlay);
      document.addEventListener("keydown", onKey, true);
      void overlay.offsetWidth;
      overlay.classList.add("open");
      okBtn.focus();
    });
  }

  // If the page's route is about to change, warn and give the user a chance
  // to cancel. Resolves true when it's OK to proceed (including no change).
  function confirmRouteChange(oldPath, newPath) {
    if (oldPath === newPath) return Promise.resolve(true);
    return confirmModal(
      "Page URL will change.",
      [
        "This change moves the page from ",
        h("code", { class: "plenti-inline-code" }, [oldPath]),
        " to ",
        h("code", { class: "plenti-inline-code" }, [newPath]),
        ". You'll be redirected to the new URL after saving."
      ]
    );
  }

  // Built-in info about the current page (type, path, file location). The
  // filename is editable via a small edit icon or by clicking the name;
  // typing previews the new filepath/path in realtime and stages the rename,
  // which is only applied when the user saves the page.
  function renderTrayInfo(tray, item, statusEl) {
    var fields = state.trayFields || item.fields;
    var dir = item.filepath.indexOf("/") >= 0 ? item.filepath.slice(0, item.filepath.lastIndexOf("/")) : "";
    var list = h("div", { class: "plenti-info" }, []);
    list.appendChild(h("div", { class: "plenti-info-row" }, [
      h("span", { class: "plenti-info-key" }, ["Type"]),
      h("span", { class: "plenti-info-val" }, [item.type || ""])
    ]));
    var pathVal = h("span", { class: "plenti-info-val" }, [item.path]);
    list.appendChild(h("div", { class: "plenti-info-row" }, [
      h("span", { class: "plenti-info-key" }, ["Path"]),
      pathVal
    ]));
    var filepathVal = h("span", { class: "plenti-info-val" }, [item.filepath]);
    list.appendChild(h("div", { class: "plenti-info-row" }, [
      h("span", { class: "plenti-info-key" }, ["Filepath"]),
      filepathVal
    ]));
    var nameView = h("span", { class: "plenti-info-val" }, [state.pendingFilename || item.filename || ""]);
    var input = h("input", {
      class: "plenti-info-input",
      type: "text",
      value: state.pendingFilename || item.filename || "",
      style: "display:none"
    }, []);
    function startEdit() {
      viewWrap.style.display = "none";
      input.style.display = "";
      input.focus();
      input.select();
    }
    var viewWrap = h("span", { class: "plenti-info-name", title: "Rename file", onclick: startEdit }, [
      h("button", { class: "plenti-json-toggle plenti-info-edit", onclick: function (e) { e.stopPropagation(); startEdit(); } }, [icon("edit")]),
      nameView
    ]);
    // Show the current (or staged) name back in view mode.
    function showView() {
      input.style.display = "none";
      viewWrap.style.display = "";
      var shown = state.pendingFilename || item.filename || "";
      nameView.textContent = shown;
      input.value = shown;
    }
    // Cancel the edit: discard the staged rename and restore the real values.
    function revertPreview() {
      state.pendingFilename = null;
      showView();
      filepathVal.textContent = item.filepath;
      pathVal.textContent = item.path;
    }
    // Realtime preview of the new filepath and route while typing; a valid,
    // changed name is staged and applied on Save.
    input.addEventListener("input", function () {
      var name = input.value.trim();
      filepathVal.textContent = (dir ? dir + "/" : "") + name + ".json";
      pathVal.textContent = computePath(item.type, (dir ? dir + "/" : "") + name, fields);
      state.pendingFilename = (name && name !== item.filename && /^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(name)) ? name : null;
    });
    input.addEventListener("keydown", function (e) {
      if (e.key === "Enter") { e.preventDefault(); input.blur(); }
      if (e.key === "Escape") { revertPreview(); }
    });
    input.addEventListener("blur", function () {
      var name = input.value.trim();
      if (name && name !== item.filename && !/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(name)) {
        notify("Invalid filename — letters, numbers, dashes and underscores only.", true);
        revertPreview();
        return;
      }
      showView();
    });
    list.appendChild(h("div", { class: "plenti-info-row" }, [
      h("span", { class: "plenti-info-key" }, ["Filename"]),
      h("span", { class: "plenti-info-cell" }, [viewWrap, input])
    ]));
    tray.appendChild(list);
  }

  // Rename the content file: handled in saveTray — the Info tab filename edit
  // stages state.pendingFilename, and Save applies the create + delete pair
  // (the commit API has no move action) after warning about route changes.

  // Switch the tray between the form, page-info, and JSON source views.
  // Leaving the JSON source requires valid JSON — a broken document can't
  // be left behind.
  function setTrayMode(tray, statusEl, mode) {
    if (state.trayJsonMode && mode !== "json") {
      var errEl = tray.querySelector(".plenti-json-error");
      if (errEl && errEl.style.display !== "none") {
        errEl.classList.add("plenti-shake");
        setTimeout(function () { errEl.classList.remove("plenti-shake"); }, 400);
        return;
      }
    }
    state.trayJsonMode = mode === "json";
    state.trayInfoMode = mode === "info";
    renderTray(tray, statusEl);
    syncBodyClasses();
  }

  function saveTray(statusEl) {
    var item = state.trayItem;
    if (!item) return;
    var dir = item.filepath.indexOf("/") >= 0 ? item.filepath.slice(0, item.filepath.lastIndexOf("/")) : "";
    var pending = state.pendingFilename;
    var filepath = pending ? (dir ? dir + "/" : "") + pending + ".json" : item.filepath;
    // The route may depend on field values (routes in plenti.json) and on the
    // filename; warn and redirect when a save would move the page.
    var newPath = computePath(item.type, filepath.replace(/\.json$/, ""), state.trayFields);
    confirmRouteChange(item.path, newPath).then(function (ok) {
      if (!ok) return;
      var creating = !!item.pendingCreate;
      statusEl.textContent = "Saving...";
      statusEl.classList.remove("success", "error");
      var contents = JSON.stringify(orderedFields(item, state.trayFields), null, 2);
      // Renames are a create + delete pair (the commit API has no move
      // action); a stubbed new page is a plain create.
      var files = pending
        ? [
            { file: "content/" + filepath, action: "create", contents: contents },
            { file: "content/" + item.filepath, action: "delete" }
          ]
        : [{ file: "content/" + filepath, action: item.pendingCreate ? "create" : "update", contents: contents }];
      commit(files)
        .then(function () {
          item.fields = JSON.parse(JSON.stringify(state.trayFields));
          sessionStorage.removeItem("plenti_cms_draft");
          delete item.pendingCreate;
          if (pending) {
            item.filepath = filepath;
            item.filename = pending;
            state.pendingFilename = null;
          }
          var pathChanged = newPath !== item.path;
          item.path = newPath;
          statusEl.innerHTML = ICONS.check + "Saved";
          statusEl.classList.add("success");
          setTimeout(function () {
            statusEl.textContent = "";
            statusEl.classList.remove("success");
          }, 3000);
          if (creating || pathChanged) {
            setTimeout(function () { window.location.href = newPath; }, 400);
          }
        })
        .catch(function (err) {
          statusEl.textContent = "Error: " + err.message;
          statusEl.classList.add("error");
        });
    });
  }

  function setTrayOpen(open) {
    state.trayOpen = open;
    if (open) sessionStorage.setItem("plenti_cms_open_tray", "1");
    else sessionStorage.removeItem("plenti_cms_open_tray");
    if (open) { state.trayJsonMode = false; state.trayInfoMode = false; state.trayArrayOpen = {}; state.trayItemOpen = {}; state.pendingFilename = null; } // always reopen in form mode, groups/items default
    var tray = document.querySelector(".plenti-admin-tray");
    var toggle = document.querySelector(".plenti-admin-bar .plenti-edit-toggle");
    var label = toggle && toggle.querySelector(".plenti-label");
    var iconEl = toggle && toggle.querySelector(".plenti-icon");
    if (open && tray) {
      renderTray(tray, document.querySelector(".plenti-admin-status"));
      // Fresh open: let fields animate in (renderTray marks the tray as
      // re-rendered to suppress animations on later in-place re-renders;
      // undo that marker now, before the first paint).
      tray.classList.remove("plenti-rendered");
    }
    if (tray) {
      // Force a reflow so the slide transition runs on first open.
      if (open) void tray.offsetWidth;
      tray.classList.toggle("open", open);
    }
    if (label) label.textContent = open ? "View" : "Edit";
    if (iconEl) iconEl.innerHTML = open ? ICONS.eye : ICONS.edit;
    syncBodyClasses();
  }

  var ICONS = {
    edit: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><g class=\"plenti-pencil\"><path d=\"M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z\"/></g></svg>",
    eye: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><g class=\"plenti-eye\"><path d=\"M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z\"/><circle cx=\"12\" cy=\"12\" r=\"3\"/></g></svg>",
    logout: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4\"/><path d=\"m16 17 5-5-5-5\"/><path d=\"M21 12H9\"/></svg>",
    check: "<svg class=\"plenti-check\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M20 6 9 17l-5-5\"/></svg>",
    alert: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M10.3 3.7 2.2 18a2 2 0 0 0 1.7 3h16.2a2 2 0 0 0 1.7-3L13.7 3.7a2 2 0 0 0-3.4 0Z\"/><path d=\"M12 9v4\"/><path d=\"M12 17h.01\"/></svg>",
    image: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><rect x=\"3\" y=\"3\" width=\"18\" height=\"18\" rx=\"2\"/><circle class=\"plenti-sun\" cx=\"9\" cy=\"9\" r=\"2\"/><path class=\"plenti-hills\" d=\"m21 15-3.1-3.1a2 2 0 0 0-2.8 0L6 21\"/></svg>",
    file: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z\"/><path d=\"M14 2v4a2 2 0 0 0 2 2h4\"/><path class=\"plenti-write-line plenti-w1\" d=\"M8 12h8\"/><path class=\"plenti-write-line plenti-w2\" d=\"M8 15h8\"/><path class=\"plenti-write-line plenti-w3\" d=\"M8 18h5\"/></svg>",
    plus: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M5 12h14\"/><path d=\"M12 5v14\"/></svg>",
    upload: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4\"/><path d=\"m17 8-5-5-5 5\"/><path d=\"M12 3v12\"/></svg>",
    download: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4\"/><path d=\"m7 10 5 5 5-5\"/><path d=\"M12 15V3\"/></svg>",
    trash: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M3 6h18\"/><path d=\"M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6\"/><path d=\"M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2\"/><path d=\"M10 11v6\"/><path d=\"M14 11v6\"/></svg>",
    link: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71\"/><path d=\"M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71\"/></svg>",
    crop: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M6 2v14a2 2 0 0 0 2 2h14\"/><path d=\"M18 22V8a2 2 0 0 0-2-2H2\"/></svg>",
    close: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M18 6 6 18\"/><path d=\"m6 6 12 12\"/></svg>",
    braces: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M8 3H7a2 2 0 0 0-2 2v5a2 2 0 0 1-2 2 2 2 0 0 1 2 2v5c0 1.1.9 2 2 2h1\"/><path d=\"M16 21h1a2 2 0 0 0 2-2v-5c0-1.1.9-2 2-2a2 2 0 0 1-2-2V5a2 2 0 0 0-2-2h-1\"/></svg>",
    info: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><circle cx=\"12\" cy=\"12\" r=\"10\"/><path d=\"M12 16v-4\"/><path d=\"M12 8h.01\"/></svg>",
    list: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M8 6h13\"/><path d=\"M8 12h13\"/><path d=\"M8 18h13\"/><path d=\"M3 6h.01\"/><path d=\"M3 12h.01\"/><path d=\"M3 18h.01\"/></svg>",
    form: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M21 6H3\"/><path d=\"M15 12H3\"/><path d=\"M17 18H3\"/></svg>",
    chevron: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><polyline points=\"9 18 15 12 9 6\"/></svg>",
    up: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M12 19V5\"/><polyline points=\"5 12 12 5 19 12\"/></svg>",
    down: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M12 5v14\"/><polyline points=\"19 12 12 19 5 12\"/></svg>",
    grip: "<svg viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><circle cx=\"9\" cy=\"6\" r=\"1\"/><circle cx=\"15\" cy=\"6\" r=\"1\"/><circle cx=\"9\" cy=\"12\" r=\"1\"/><circle cx=\"15\" cy=\"12\" r=\"1\"/><circle cx=\"9\" cy=\"18\" r=\"1\"/><circle cx=\"15\" cy=\"18\" r=\"1\"/></svg>"
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

  /* ---------- Content modal (embedded CMS) ---------- */

  function missingDefaultsModal(type, hasExisting, onUseExisting) {
    var actions = [];
    if (hasExisting) {
      actions.push(h("button", { class: "plenti-btn", onclick: function () { close(); } }, ["Cancel"]));
    } else {
      actions.push(h("button", { class: "plenti-btn primary", onclick: function () { close(); } }, ["Close"]));
    }
    var children = [
      h("h3", { class: "plenti-newitem-title" }, ["New " + type]),
      h("div", { class: "plenti-newitem-warning", role: "alert" }, [icon("alert"), h("span", {}, ["Missing _defaults.json"])])
    ];
    if (hasExisting) {
      children.push(h("button", {
        class: "plenti-existing-choice",
        onclick: function () { close(); if (onUseExisting) onUseExisting(); }
      }, [
        h("span", { class: "plenti-existing-check", "aria-hidden": "true" }, [icon("check")]),
        h("span", {}, ["Use existing content as starting point"])
      ]));
    } else {
      children.push(h("p", { class: "plenti-newitem-message" }, ["No existing content of type exists"]));
    }
    children.push(h("div", { class: "plenti-confirm-actions" }, actions));
    var overlay = h("div", {
      class: "plenti-media-overlay plenti-confirm-overlay",
      onclick: function (e) { if (e.target === overlay) close(); }
    }, [h("div", { class: "plenti-media-modal plenti-confirm-dialog plenti-newitem-dialog" }, children)]);
    function onKey(e) { if (e.key === "Escape") { e.preventDefault(); close(); } }
    function close() {
      document.removeEventListener("keydown", onKey);
      overlay.classList.remove("open");
      setTimeout(function () { overlay.remove(); }, 200);
    }
    document.addEventListener("keydown", onKey);
    document.body.appendChild(overlay);
    void overlay.offsetWidth;
    overlay.classList.add("open");
    var primary = overlay.querySelector("button");
    if (primary) primary.focus();
  }


  // Choose the initial fields for a draft. Prefer an explicit _defaults.json
  // blueprint. Without one, an existing content node may be used after the
  // editor opts in. Nothing is committed until Save in the tray.
  function startingContent(type) {
    var defaults = state.data.defaults || {};
    if (Object.prototype.hasOwnProperty.call(defaults, type)) {
      return { fields: defaults[type] || {}, fieldOrder: Object.keys(defaults[type] || {}), source: "defaults" };
    }
    var items = state.data.content || [];
    for (var i = 0; i < items.length; i++) {
      if (items[i].type === type) {
        return { fields: items[i].fields || {}, fieldOrder: Object.keys(orderedFields(items[i])), source: "existing" };
      }
    }
    return null;
  }

  function applyFieldDefaults(type, fields) {
    var result = JSON.parse(JSON.stringify(fields || {}));
    var schema = (state.data.schemas && state.data.schemas[type]) || {};
    Object.keys(schema).forEach(function (key) {
      if (key.indexOf(".") !== -1 || Object.prototype.hasOwnProperty.call(result, key)) return;
      var options = fieldOptions(schema, key);
      if (options.default !== undefined) {
        var value = options.default;
        var fieldType = schemaType(schema, key, value);
        if ((fieldType === "date" || fieldType === "time") && typeof value === "string") {
          var match = value.match(/^(today|now)(?:([+-])(\\d+))?$/);
          if (match) {
            var date = new Date();
            if (match[2]) date.setDate(date.getDate() + (match[2] === "+" ? 1 : -1) * Number(match[3]));
            value = fieldType === "date" ? date.toISOString().slice(0, 10) : date.toTimeString().slice(0, 5);
          }
        }
        result[key] = JSON.parse(JSON.stringify(value));
      } else if (schemaType(schema, key, null) === "id") result[key] = Date.now().toString(36);
    });
    return result;
  }

  function createItemStub(type, name, startingFields, startingFieldOrder) {
    var defaults = startingFields || {};
    var host = (state.data.content || []).find(function (content) {
      return content.type === type && !content.pendingCreate;
    });
    if (!host) {
      notify("No existing page is available to preview this content type", true);
      return;
    }
    var item = {
      type: type,
      path: computePath(type, type + "/" + name, defaults),
      filename: name,
      filepath: type + "/" + name + ".json",
      fields: JSON.parse(JSON.stringify(defaults)),
      fieldOrder: (startingFieldOrder || Object.keys(defaults)).slice(),
      previewPath: host.path,
      pendingCreate: true
    };
    sessionStorage.setItem("plenti_cms_draft", JSON.stringify(item));
    sessionStorage.setItem("plenti_cms_open_tray", "1");
    // Load a real page of this type so its compiled layout is available. The
    // hash identifies the draft without pretending its final route exists.
    window.location.href = host.path + "#add/" + encodeURIComponent(type) + "/" + encodeURIComponent(name);
  }

  function restoreDraft() {
    var match = (window.location.hash || "").match(/^#add\/([^/]+)\/([^/]+)$/);
    if (!match) return;
    var raw = sessionStorage.getItem("plenti_cms_draft");
    if (!raw) return;
    try {
      var item = JSON.parse(raw);
      if (item.type !== decodeURIComponent(match[1]) || item.filename !== decodeURIComponent(match[2])) return;
      var currentPath = window.location.pathname.replace(/\/+$/, "") || "/";
      var previewPath = (item.previewPath || "").replace(/\/+$/, "") || "/";
      if (previewPath !== currentPath) {
        window.location.href = item.previewPath + window.location.hash;
        return;
      }
      item.pendingCreate = true;
      state.data.content.push(item);
      state.trayItem = item;
      state.trayFields = JSON.parse(JSON.stringify(orderedFields(item)));
      if (window.Pattr && window.Pattr.data) {
        try { window.Pattr.data.content = item; } catch (e) { /* read-only scope */ }
      }
      setTrayOpen(true);
      // Keep the host's compiled template, but show a draft-specific URL that
      // cannot be mistaken for the existing page's editing URL.
      window.history.replaceState({}, "", "/#add/" + encodeURIComponent(item.type) + "/" + encodeURIComponent(item.filename));
    } catch (e) {
      sessionStorage.removeItem("plenti_cms_draft");
    }
  }

  // Custom "new item" modal (replaces window.prompt): filename input with the
  // .json extension ghosted inline and a live preview of the saved filepath.
  function newItemModal(type, onCreate, startingFields, startingFieldOrder) {
    var starting = startingContent(type);
    if (!starting) {
      missingDefaultsModal(type, false);
      return;
    }
    if (starting.source === "existing" && !startingFields) {
      missingDefaultsModal(type, true, function () {
        newItemModal(type, onCreate, starting.fields, starting.fieldOrder);
      });
      return;
    }
    startingFields = startingFields || starting.fields;
    startingFieldOrder = startingFieldOrder || starting.fieldOrder;
    startingFields = applyFieldDefaults(type, startingFields);
    startingFieldOrder = Object.keys(startingFields);
    function cleanName() {
      var name = input.value.trim().toLowerCase().replace(/[^a-z0-9-_]+/g, "-").replace(/^-+|-+$/g, "");
      return /^[a-z0-9][a-z0-9-_]*$/.test(name) ? name : null;
    }
    var input = h("input", { type: "text", placeholder: "my-page", autocomplete: "off", spellcheck: "false" });
    var preview = h("code", { class: "plenti-inline-code" }, [type + "/\u2026.json"]);
    var okBtn = h("button", {
      class: "plenti-btn primary", disabled: true,
      onclick: function () { submit(); }
    }, ["Create"]);
    input.addEventListener("input", function () {
      var name = cleanName();
      okBtn.disabled = !name;
      preview.textContent = type + "/" + (name || "\u2026") + ".json";
    });
    input.addEventListener("keydown", function (e) {
      if (e.key === "Enter") { e.preventDefault(); submit(); }
      if (e.key === "Escape") { e.preventDefault(); close(); }
    });
    function submit() {
      var name = cleanName();
      if (!name) return;
      close();
      if (onCreate) onCreate();
      createItemStub(type, name, startingFields, startingFieldOrder);
    }
    var overlay = h("div", {
      class: "plenti-media-overlay plenti-confirm-overlay",
      onclick: function (e) { if (e.target === overlay) close(); }
    }, [
      h("div", { class: "plenti-media-modal plenti-confirm-dialog plenti-newitem-dialog" }, [
        h("h3", { class: "plenti-newitem-title" }, ["New " + type]),
        h("label", { class: "plenti-newitem-label" }, ["Filename"]),
        h("div", { class: "plenti-newitem-input" }, [
          input,
          h("span", { class: "plenti-newitem-ext" }, [".json"])
        ]),
        h("div", { class: "plenti-newitem-preview" }, ["Saved as ", preview]),
        h("div", { class: "plenti-confirm-actions" }, [
          h("button", { class: "plenti-btn", onclick: function () { close(); } }, ["Cancel"]),
          okBtn
        ])
      ])
    ]);
    function close() {
      overlay.classList.remove("open");
      setTimeout(function () { overlay.remove(); }, 200);
    }
    document.body.appendChild(overlay);
    void overlay.offsetWidth; // reflow so the entrance transition runs
    overlay.classList.add("open");
    input.focus();
  }

  function matchesContentSearch(item, query) {
    var terms = query.toLowerCase().trim().split(/\s+/).filter(Boolean);
    var fields = item.fields || {};
    var text = [item.type, item.path, item.filepath, fields.title, fields.name]
      .filter(function (value) { return value != null; }).join(" ").toLowerCase();
    return terms.every(function (term) { return text.indexOf(term) !== -1; });
  }

  function isSingleType(item) {
    return item.filepath.indexOf("/") === -1;
  }

  // Content browser: all content grouped by type, styled like the media
  // modal. Clicking an item navigates to its page for editing in the tray.
  function openContentModal() {
    var byType = {};
    var searchableRows = [];
    (state.data.content || []).forEach(function (c) {
      (byType[c.type] = byType[c.type] || []).push(c);
    });
    var noResults = h("div", { class: "plenti-admin-empty", style: "display:none" }, ["No matching content"]);
    var loading = h("div", { class: "plenti-admin-empty" }, ["Loading content\u2026"]);
    var spacer = h("div", { class: "plenti-content-loading-spacer", "aria-hidden": "true" }, []);
    var body = h("div", { class: "plenti-content-body" }, [noResults, loading, spacer]);
    var search = h("input", {
      class: "plenti-content-search", type: "search",
      placeholder: "Search content\u2026", "aria-label": "Search content",
      autocomplete: "off", spellcheck: "false"
    });
    function filterContent() {
      var matches = 0;
      searchableRows.forEach(function (entry) {
        var visible = matchesContentSearch(entry.item, search.value);
        entry.row.style.display = visible ? "" : "none";
        if (visible) matches++;
      });
      noResults.style.display = matches || !searchableRows.length ? "none" : "";
    }
    search.addEventListener("input", filterContent);
    search.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && search.value) {
        e.preventDefault();
        e.stopPropagation();
        search.value = "";
        search.dispatchEvent(new Event("input"));
      }
    });
    var overlay = h("div", {
      class: "plenti-media-overlay",
      onclick: function (e) { if (e.target === overlay) close(); }
    }, [
      h("div", { class: "plenti-media-modal plenti-content-modal" }, [
        h("div", { class: "plenti-media-header" }, [
          h("h3", {}, ["Content"]),
          h("button", { class: "plenti-media-close", onclick: function () { close(); } }, ["\u00D7"])
        ]),
        search,
        body
      ])
    ]);
    function close() {
      overlay.classList.remove("open");
      setTimeout(function () { overlay.remove(); }, 200);
    }
    document.body.appendChild(overlay);
    void overlay.offsetWidth; // reflow so the entrance transition runs
    overlay.classList.add("open");
    search.focus();
    // Fill the viewport with an initial chunk, then load the remaining rows in
    // small tasks so a large content index never blocks modal interaction.
    var types = Object.keys(byType).sort();
    var typeIndex = 0;
    var rowIndex = 0;
    var list = null;
    // Reserve the final list's height so batch rendering does not continually
    // enlarge the scrollbar track. Rows are 30px plus a 2px gap; each section
    // contributes a 22px heading and 14px bottom margin.
    function reserveRemainingHeight() {
      var height = 0;
      for (var i = typeIndex; i < types.length; i++) {
        var items = byType[types[i]];
        var unloaded = items.length - (i === typeIndex ? rowIndex : 0);
        height += unloaded * 32;
        if (i > typeIndex || rowIndex === 0) height += 36;
        if (i > typeIndex && !isSingleType(items[0])) height += 32;
      }
      spacer.style.height = height + "px";
    }
    function renderContentChunk() {
      if (!overlay.isConnected) return;
      var remaining = 80;
      while (typeIndex < types.length && remaining > 0) {
        var type = types[typeIndex];
        var items = byType[type];
        if (!list) {
          list = h("div", { class: "plenti-content-list" }, []);
          body.insertBefore(h("div", { class: "plenti-content-section" }, [
            h("h4", {}, [type]), list
          ]), spacer);
          loading.remove();
        }
        while (rowIndex < items.length && remaining > 0) {
          var c = items[rowIndex++];
          var row = (function (item) {
            return h("button", {
              class: "plenti-content-item",
              onclick: function () {
                sessionStorage.setItem("plenti_cms_open_tray", "1");
                window.location.href = item.path;
              }
            }, [
              icon("file"),
              h("span", { class: "plenti-content-path" }, [item.path]),
              h("span", { class: "plenti-content-file" }, [item.filepath])
            ]);
          })(c);
          list.appendChild(row);
          searchableRows.push({ item: c, row: row });
          remaining--;
        }
        if (rowIndex === items.length) {
          if (!isSingleType(items[0])) {
            list.appendChild(h("button", {
              class: "plenti-content-item plenti-content-add",
              // Leave Content open underneath so Cancel returns to it unchanged.
              onclick: (function (contentType) {
                return function () { newItemModal(contentType, close); };
              })(type)
            }, [icon("plus"), h("span", { class: "plenti-content-path" }, ["add " + type])]));
          }
          typeIndex++;
          rowIndex = 0;
          list = null;
        }
      }
      reserveRemainingHeight();
      filterContent();
      if (typeIndex < types.length) {
        setTimeout(renderContentChunk, 0);
      } else {
        loading.remove();
        if (!searchableRows.length) body.appendChild(h("div", { class: "plenti-admin-empty" }, ["No content yet"]));
      }
    }
    reserveRemainingHeight();
    setTimeout(renderContentChunk, 0);
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

  function openMediaModal(onSelect) {
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
      if (onSelect) {
        close();
        onSelect(m);
        return;
      }
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
            if (onSelect) pick("/media/" + files[files.length - 1].name, "");
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
          h("h3", {}, [onSelect ? "Select Media" : "Media"]),
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
        class: "plenti-btn plenti-content-toggle",
        onclick: function () { openContentModal(); }
      }, [icon("file"), h("span", { class: "plenti-label" }, ["Content"])]),
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

  function syncBodyClasses() {
    var authenticated = isAuthenticated();
    var open = authenticated && state.trayOpen;
    var mode = state.trayJsonMode ? "code" : state.trayInfoMode ? "info" : "form";
    var classes = document.body.classList;
    classes.toggle("plenti-admin-active", authenticated && !!document.querySelector("[plenti-cms]"));
    classes.toggle("plenti-tray-open", open);
    ["form", "info", "code"].forEach(function (name) {
      classes.toggle("plenti-tray-" + name, open && mode === name);
    });
  }

  function syncAdminBar() {
    var el = document.querySelector("[plenti-cms]");
    if (!isAuthenticated() || !el) {
      syncBodyClasses();
      return;
    }
    document.body.classList.add("plenti-admin-active");
    syncBodyClasses();
    if (!el.querySelector(".plenti-admin-bar")) {
      mountAdminBar(el);
      // Keep the tray open across page navigation, including content browser links.
      if (sessionStorage.getItem("plenti_cms_open_tray")) setTrayOpen(true);
    }
  }

  function boot() {
    loadData().then(function () {
      restoreDraft();
      publishUser();
    }).catch(function (err) {
      var el = document.querySelector("[plenti-cms]");
      if (el) el.textContent = "Could not load CMS data: " + err.message;
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
