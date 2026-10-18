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
    jsonText: ""
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

  function fieldInput(key, type, value, onChange) {
    var schema = state.data.schemas[currentType()] || {};
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
      return h("textarea", { onchange: function (e) { onChange(e.target.value); } }, [value || ""]);
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
    return h("input", { type: "text", value: value == null ? "" : value, onchange: function (e) { onChange(e.target.value); } });
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
          fieldInput(key, type, state.fields[key], function (v) { state.fields[key] = v; })
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
    var root = document.getElementById("plenti-cms-app");
    root.innerHTML = "";
    var main = state.view === "media" ? renderMedia() : renderEditor();
    root.appendChild(h("div", { class: "plenti-layout" }, [renderSidebar(), main]));
  }

  /* ---------- Actions ---------- */

  function openItem(index) {
    state.view = "content";
    state.current = index;
    state.fields = JSON.parse(JSON.stringify(state.data.content[index].fields || {}));
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
    commit([{ file: "content/" + item.filepath, action: "update", contents: JSON.stringify(state.fields, null, 2) }])
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

  /* ---------- Boot ---------- */

  fetch("/admin/data.json")
    .then(function (res) { return res.json(); })
    .then(function (data) {
      state.data = data;
      var cms = data.env.cms || {};
      var provider = (cms.provider || "local").toLowerCase();
      if (!data.env.local && provider !== "local" && !getToken()) {
        var token = prompt("Enter a personal access token for " + provider + " (stored in this browser):");
        if (token) localStorage.setItem("plenti_cms_token", token);
      }
      if (data.content.length) openItem(0);
      else { state.view = "media"; render(); }
    })
    .catch(function (err) {
      document.getElementById("plenti-cms-app").textContent = "Could not load CMS data: " + err.message;
    });
})();
