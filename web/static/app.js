(function () {
  "use strict";

  const CSRF_HEADER = "X-Nostromo-Web";

  const state = {
    meta: { version: "", modes: [], languages: [] },
    manifests: [],
    manifest: null,     // selected manifest name
    tree: null,         // manifestDetail for the selected manifest
    selected: null,     // keyPath of the selected command
    command: null,      // commandNode currently shown in the editor
    collapsed: new Set(),
    query: "",
    dirty: false,
  };

  const $ = (sel) => document.querySelector(sel);
  const els = {
    version: $("#version"),
    search: $("#search"),
    searchClear: $("#search-clear"),
    refresh: $("#refresh"),
    newCommand: $("#new-command"),
    manifests: $("#manifests"),
    treeTitle: $("#tree-title"),
    tree: $("#tree"),
    results: $("#results"),
    expandAll: $("#expand-all"),
    collapseAll: $("#collapse-all"),
    empty: $("#empty"),
    detail: $("#detail"),
    keypath: $("#detail-keypath"),
    badges: $("#detail-badges"),
    readonlyNote: $("#readonly-note"),
    save: $("#save"),
    reset: $("#reset"),
    addChild: $("#add-child"),
    del: $("#delete"),
    subsCount: $("#subs-count"),
    subsBody: $("#subs-table tbody"),
    subAlias: $("#sub-alias"),
    subName: $("#sub-name"),
    subAdd: $("#sub-add-btn"),
    dialog: $("#add-dialog"),
    addForm: $("#add-form"),
    addCancel: $("#add-cancel"),
    toast: $("#toast"),
  };

  // ---------------------------------------------------------------------------
  // helpers

  function h(tag, attrs, ...children) {
    const el = document.createElement(tag);
    if (attrs) {
      for (const [k, v] of Object.entries(attrs)) {
        if (v === null || v === undefined || v === false) continue;
        if (k === "class") el.className = v;
        else if (k === "text") el.textContent = v;
        else if (k.startsWith("on")) el.addEventListener(k.slice(2), v);
        else el.setAttribute(k, v === true ? "" : v);
      }
    }
    for (const c of children) {
      if (c === null || c === undefined || c === false) continue;
      el.append(c instanceof Node ? c : document.createTextNode(String(c)));
    }
    return el;
  }

  let toastTimer = null;
  function toast(msg, kind) {
    els.toast.textContent = msg;
    els.toast.className = "toast" + (kind ? " " + kind : "");
    els.toast.hidden = false;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => { els.toast.hidden = true; }, kind === "error" ? 6000 : 2500);
  }

  async function api(method, path, body) {
    const opts = { method, headers: { Accept: "application/json" } };
    if (method !== "GET") {
      opts.headers[CSRF_HEADER] = "1";
      if (body !== undefined) {
        opts.headers["Content-Type"] = "application/json";
        opts.body = JSON.stringify(body);
      }
    }
    const res = await fetch(path, opts);
    let data = null;
    try { data = await res.json(); } catch (e) { data = null; }
    if (!res.ok) {
      throw new Error((data && data.error) || res.status + " " + res.statusText);
    }
    return data;
  }

  const qs = (params) => new URLSearchParams(params).toString();

  function countNodes(nodes) {
    let n = 0;
    for (const c of nodes || []) n += 1 + countNodes(c.commands);
    return n;
  }

  function highlight(text, query) {
    if (!query) return document.createTextNode(text || "");
    const frag = document.createDocumentFragment();
    const lower = (text || "").toLowerCase();
    const q = query.toLowerCase();
    let i = 0;
    while (true) {
      const idx = lower.indexOf(q, i);
      if (idx < 0) { frag.append(text.slice(i)); break; }
      frag.append(text.slice(i, idx));
      frag.append(h("mark", { text: text.slice(idx, idx + q.length) }));
      i = idx + q.length;
    }
    return frag;
  }

  function fillSelect(select, values, opts) {
    select.replaceChildren();
    if (opts && opts.blank) select.append(h("option", { value: "", text: opts.blank }));
    for (const v of values) select.append(h("option", { value: v, text: v }));
  }

  // ---------------------------------------------------------------------------
  // loading

  async function loadMeta() {
    state.meta = await api("GET", "/api/meta");
    els.version.textContent = state.meta.version ? "v" + state.meta.version : "";
    fillSelect(els.detail.elements.mode, state.meta.modes);
    fillSelect(els.detail.elements.language, state.meta.languages, { blank: "none" });
    fillSelect(els.addForm.elements.mode, state.meta.modes, { blank: "default" });
    fillSelect(els.addForm.elements.language, state.meta.languages, { blank: "none" });
  }

  async function loadManifests() {
    state.manifests = await api("GET", "/api/manifests");
    if (!state.manifests.some((m) => m.name === state.manifest)) {
      const core = state.manifests.find((m) => m.core) || state.manifests[0];
      state.manifest = core ? core.name : null;
    }
    renderManifests();
    await loadTree();
  }

  async function loadTree() {
    if (!state.manifest) {
      state.tree = null;
      renderTree();
      return;
    }
    state.tree = await api("GET", "/api/manifest?" + qs({ name: state.manifest }));
    renderTree();
  }

  async function loadCommand(keyPath) {
    try {
      state.command = await api("GET", "/api/command?" + qs({ keypath: keyPath }));
      state.selected = keyPath;
      state.dirty = false;
      renderDetail();
      highlightSelected();
    } catch (err) {
      toast(err.message, "error");
      if (state.selected === keyPath) clearSelection();
    }
  }

  async function refreshAll(keepSelection) {
    await loadManifests();
    if (keepSelection && state.selected) {
      await loadCommand(state.selected);
    }
  }

  function clearSelection() {
    state.selected = null;
    state.command = null;
    state.dirty = false;
    renderDetail();
    highlightSelected();
  }

  // ---------------------------------------------------------------------------
  // rendering: manifests

  function renderManifests() {
    els.manifests.replaceChildren();
    for (const m of state.manifests) {
      const btn = h("button", {
        type: "button",
        class: "manifest" + (m.name === state.manifest ? " active" : ""),
        title: m.path + (m.source ? "\n" + m.source : ""),
        onclick: async () => {
          state.manifest = m.name;
          state.query = "";
          els.search.value = "";
          renderManifests();
          await loadTree();
        },
      },
        h("span", { class: "m-name", text: m.name }),
        h("span", { class: "badge" + (m.core ? "" : " docked"), text: m.core ? "core" : "docked" }),
        h("span", { class: "m-meta", text: m.commandCount + (m.commandCount === 1 ? " cmd" : " cmds") }),
      );
      els.manifests.append(h("li", null, btn));
    }
  }

  // ---------------------------------------------------------------------------
  // rendering: tree

  function renderTree() {
    const searching = state.query.length > 0;
    els.tree.hidden = searching;
    els.results.hidden = !searching;
    if (searching) return;

    els.tree.replaceChildren();
    const t = state.tree;
    if (!t) {
      els.treeTitle.textContent = "Commands";
      return;
    }
    els.treeTitle.textContent = t.name + " · " + t.commandCount + (t.commandCount === 1 ? " command" : " commands");
    if (!t.commands.length) {
      els.tree.append(h("div", { class: "tree-empty" },
        t.core ? "No commands yet. Use “New command” to add one." : "This manifest has no commands."));
      return;
    }
    els.tree.append(renderNodes(t.commands));
    highlightSelected();
  }

  function renderNodes(nodes) {
    const ul = h("ul");
    for (const n of nodes) ul.append(renderNode(n));
    return ul;
  }

  function renderNode(n) {
    const hasKids = n.commands && n.commands.length > 0;
    const li = h("li", { class: "node" + (hasKids && state.collapsed.has(n.keyPath) ? " collapsed" : ""), "data-keypath": n.keyPath });
    const twisty = h("span", {
      class: "twisty" + (hasKids ? "" : " leaf"),
      text: "▼",
      onclick: (e) => {
        e.stopPropagation();
        toggleCollapsed(n.keyPath, li);
      },
    });
    const row = h("div", {
      class: "node-row" + (n.disabled ? " disabled" : ""),
      title: n.keyPath + (n.description ? " — " + n.description : ""),
      onclick: () => selectCommand(n.keyPath),
      ondblclick: () => hasKids && toggleCollapsed(n.keyPath, li),
    },
      twisty,
      h("span", { class: "node-alias", text: n.alias }),
      n.name ? h("span", { class: "node-cmd", text: n.name }) : (n.code && n.code.language ? h("span", { class: "node-cmd", text: "<" + n.code.language + " snippet>" }) : null),
      hasKids ? h("span", { class: "node-count", text: String(countNodes(n.commands)) }) : null,
    );
    li.append(row);
    if (hasKids) li.append(renderNodes(n.commands));
    return li;
  }

  function toggleCollapsed(keyPath, li) {
    if (state.collapsed.has(keyPath)) state.collapsed.delete(keyPath);
    else state.collapsed.add(keyPath);
    li.classList.toggle("collapsed", state.collapsed.has(keyPath));
  }

  function setAllCollapsed(collapsed) {
    state.collapsed.clear();
    if (collapsed && state.tree) {
      const walk = (nodes) => {
        for (const n of nodes) {
          if (n.commands && n.commands.length) { state.collapsed.add(n.keyPath); walk(n.commands); }
        }
      };
      walk(state.tree.commands);
    }
    renderTree();
  }

  function expandTo(keyPath) {
    const parts = keyPath.split(".");
    for (let i = 1; i < parts.length; i++) {
      state.collapsed.delete(parts.slice(0, i).join("."));
    }
  }

  function highlightSelected() {
    for (const row of els.tree.querySelectorAll(".node-row.selected")) row.classList.remove("selected");
    if (!state.selected) return;
    const li = els.tree.querySelector('li[data-keypath="' + CSS.escape(state.selected) + '"]');
    if (li) {
      const row = li.querySelector(":scope > .node-row");
      row.classList.add("selected");
      row.scrollIntoView({ block: "nearest" });
    }
  }

  async function selectCommand(keyPath, opts) {
    if (state.dirty && !confirm("Discard unsaved changes?")) return;
    if (opts && opts.fromSearch) {
      const owner = state.manifests.find((m) => m.name === opts.manifest);
      if (owner && owner.name !== state.manifest) {
        state.manifest = owner.name;
        renderManifests();
        state.tree = await api("GET", "/api/manifest?" + qs({ name: state.manifest }));
      }
    }
    expandTo(keyPath);
    if (!state.query) renderTree();
    await loadCommand(keyPath);
  }

  // ---------------------------------------------------------------------------
  // rendering: search

  let searchTimer = null;
  function onSearchInput() {
    const q = els.search.value.trim();
    els.searchClear.hidden = q.length === 0;
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => runSearch(q), 150);
  }

  async function runSearch(q) {
    state.query = q;
    if (!q) {
      renderTree();
      return;
    }
    try {
      const res = await api("GET", "/api/search?" + qs({ q }));
      if (res.query !== state.query) return; // stale
      renderResults(res);
    } catch (err) {
      toast(err.message, "error");
    }
  }

  function renderResults(res) {
    els.tree.hidden = true;
    els.results.hidden = false;
    els.results.replaceChildren();
    els.treeTitle.textContent = "Search results";

    const total = res.commands.length + res.substitutions.length;
    if (total === 0) {
      els.results.append(h("div", { class: "results-empty", text: "No matches for “" + res.query + "”." }));
      return;
    }

    if (res.commands.length) {
      els.results.append(h("h4", { text: "Commands (" + res.commands.length + ")" }));
      for (const c of res.commands) els.results.append(renderResult(c, res.query, c.name || c.description));
    }
    if (res.substitutions.length) {
      els.results.append(h("h4", { text: "Substitutions (" + res.substitutions.length + ")" }));
      for (const c of res.substitutions) {
        const subs = (c.subs || []).filter((s) => s.alias.toLowerCase().includes(res.query.toLowerCase()) || s.name.toLowerCase().includes(res.query.toLowerCase()));
        const desc = subs.map((s) => s.alias + " → " + s.name).join(", ");
        els.results.append(renderResult(c, res.query, desc));
      }
    }
  }

  function renderResult(c, query, meta) {
    return h("button", {
      type: "button",
      class: "result",
      onclick: () => selectCommand(c.keyPath, { fromSearch: true, manifest: c.manifest }),
    },
      h("span", { class: "r-key" }, highlight(c.keyPath, query)),
      " ",
      h("span", { class: "badge" + (c.readOnly ? " docked" : ""), text: c.manifest }),
      h("span", { class: "r-meta" }, highlight(meta || "", query)),
    );
  }

  // ---------------------------------------------------------------------------
  // rendering: detail editor

  function renderDetail() {
    const c = state.command;
    els.empty.hidden = !!c;
    els.detail.hidden = !c;
    if (!c) return;

    const f = els.detail.elements;
    els.keypath.textContent = c.keyPath;

    els.badges.replaceChildren(...[
      h("span", { class: "badge" + (c.readOnly ? " docked" : ""), text: c.manifest + (c.readOnly ? " · docked" : " · core") }),
      c.disabled ? h("span", { class: "badge disabled", text: "disabled" }) : null,
      c.aliasOnly ? h("span", { class: "badge alias", text: "alias only" }) : null,
      c.code && c.code.language ? h("span", { class: "badge code", text: c.code.language }) : null,
    ].filter(Boolean));
    els.readonlyNote.hidden = !c.readOnly;

    f.alias.value = c.alias || "";
    f.name.value = c.name || "";
    f.description.value = c.description || "";
    f.mode.value = c.mode || "";
    f.aliasOnly.checked = !!c.aliasOnly;
    f.disabled.checked = !!c.disabled;
    f.language.value = (c.code && c.code.language) || "";
    f.snippet.value = (c.code && c.code.snippet) || "";

    for (const el of els.detail.querySelectorAll("input, select, textarea, button")) {
      el.disabled = c.readOnly;
    }
    els.reset.disabled = c.readOnly;
    setDirty(false);

    renderSubs(c);
  }

  function renderSubs(c) {
    const subs = c.subs || [];
    els.subsCount.textContent = subs.length ? "(" + subs.length + ")" : "";
    els.subsBody.replaceChildren();
    if (!subs.length) {
      els.subsBody.append(h("tr", null, h("td", { colspan: "3", class: "hint", text: "No substitutions on this command." })));
    }
    for (const s of subs) {
      els.subsBody.append(h("tr", null,
        h("td", { class: "mono", text: s.alias }),
        h("td", { class: "mono", text: s.name }),
        h("td", null, c.readOnly ? null : h("button", {
          type: "button", class: "btn danger", text: "Remove",
          onclick: () => removeSub(c.keyPath, s.alias),
        })),
      ));
    }
  }

  function setDirty(d) {
    state.dirty = d;
    els.detail.querySelector(".form-actions").classList.toggle("dirty", d);
  }

  function formToUpdate() {
    const f = els.detail.elements;
    return {
      alias: f.alias.value.trim(),
      name: f.name.value,
      description: f.description.value,
      mode: f.mode.value,
      aliasOnly: f.aliasOnly.checked,
      disabled: f.disabled.checked,
      code: { language: f.language.value, snippet: f.snippet.value },
    };
  }

  // ---------------------------------------------------------------------------
  // mutations

  async function saveCommand(e) {
    e.preventDefault();
    const c = state.command;
    if (!c || c.readOnly) return;
    if (!els.detail.reportValidity()) return;
    const body = formToUpdate();
    try {
      const updated = await api("PUT", "/api/command?" + qs({ keypath: c.keyPath }), body);
      toast("Saved " + updated.keyPath, "ok");
      state.selected = updated.keyPath;
      state.command = updated;
      await loadManifests();
      renderDetail();
      highlightSelected();
    } catch (err) {
      toast(err.message, "error");
    }
  }

  async function deleteCommand() {
    const c = state.command;
    if (!c || c.readOnly) return;
    const kids = countNodes(c.commands);
    const msg = "Delete " + c.keyPath + (kids ? " and its " + kids + " subcommand" + (kids === 1 ? "" : "s") : "") + "?";
    if (!confirm(msg)) return;
    try {
      await api("DELETE", "/api/command?" + qs({ keypath: c.keyPath }));
      toast("Removed " + c.keyPath, "ok");
      clearSelection();
      await loadManifests();
    } catch (err) {
      toast(err.message, "error");
    }
  }

  async function addSub() {
    const c = state.command;
    if (!c || c.readOnly) return;
    const alias = els.subAlias.value.trim();
    const name = els.subName.value.trim();
    if (!alias || !name) {
      toast("Both alias and original are required", "error");
      return;
    }
    try {
      const updated = await api("POST", "/api/command/sub?" + qs({ keypath: c.keyPath }), { alias, name });
      els.subAlias.value = "";
      els.subName.value = "";
      state.command = updated;
      renderSubs(updated);
      toast("Added substitution " + alias, "ok");
    } catch (err) {
      toast(err.message, "error");
    }
  }

  async function removeSub(keyPath, alias) {
    try {
      const updated = await api("DELETE", "/api/command/sub?" + qs({ keypath: keyPath, alias }));
      state.command = updated;
      renderSubs(updated);
      toast("Removed substitution " + alias, "ok");
    } catch (err) {
      toast(err.message, "error");
    }
  }

  function openAddDialog(parent) {
    const f = els.addForm;
    f.reset();
    f.elements.parent.value = parent || "";
    els.dialog.showModal();
    (parent ? f.elements.alias : f.elements.parent).focus();
  }

  async function submitAdd(e) {
    e.preventDefault();
    const f = els.addForm.elements;
    if (!els.addForm.reportValidity()) return;
    const parent = f.parent.value.trim().replace(/^\.+|\.+$/g, "");
    const alias = f.alias.value.trim();
    const keyPath = parent ? parent + "." + alias : alias;
    const body = {
      keyPath,
      name: f.name.value,
      description: f.description.value,
      mode: f.mode.value,
      aliasOnly: f.aliasOnly.checked,
      code: { language: f.language.value, snippet: f.snippet.value },
    };
    try {
      const created = await api("POST", "/api/command", body);
      els.dialog.close();
      toast("Created " + created.keyPath, "ok");
      state.manifest = created.manifest;
      state.selected = created.keyPath;
      expandTo(created.keyPath);
      els.search.value = "";
      state.query = "";
      els.searchClear.hidden = true;
      await loadManifests();
      await loadCommand(created.keyPath);
    } catch (err) {
      toast(err.message, "error");
    }
  }

  // ---------------------------------------------------------------------------
  // wiring

  els.search.addEventListener("input", onSearchInput);
  els.searchClear.addEventListener("click", () => {
    els.search.value = "";
    onSearchInput();
    els.search.focus();
  });
  els.refresh.addEventListener("click", () => refreshAll(true).then(() => toast("Reloaded", "ok")).catch((e) => toast(e.message, "error")));
  els.newCommand.addEventListener("click", () => openAddDialog(""));
  els.expandAll.addEventListener("click", () => setAllCollapsed(false));
  els.collapseAll.addEventListener("click", () => setAllCollapsed(true));

  els.detail.addEventListener("submit", saveCommand);
  els.detail.addEventListener("input", (e) => {
    if (e.target.closest(".subs")) return;
    setDirty(true);
  });
  els.reset.addEventListener("click", () => renderDetail());
  els.del.addEventListener("click", deleteCommand);
  els.addChild.addEventListener("click", () => state.command && openAddDialog(state.command.keyPath));
  els.subAdd.addEventListener("click", addSub);
  for (const el of [els.subAlias, els.subName]) {
    el.addEventListener("keydown", (e) => {
      if (e.key === "Enter") { e.preventDefault(); addSub(); }
    });
  }

  els.addForm.addEventListener("submit", submitAdd);
  els.addCancel.addEventListener("click", () => els.dialog.close());

  document.addEventListener("keydown", (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "k") {
      e.preventDefault();
      els.search.focus();
      els.search.select();
    }
    if ((e.ctrlKey || e.metaKey) && e.key === "s" && !els.detail.hidden) {
      e.preventDefault();
      els.save.click();
    }
  });

  window.addEventListener("beforeunload", (e) => {
    if (state.dirty) { e.preventDefault(); e.returnValue = ""; }
  });

  (async function init() {
    try {
      await loadMeta();
      await loadManifests();
    } catch (err) {
      toast("Failed to load: " + err.message, "error");
    }
  })();
})();
