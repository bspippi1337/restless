const el = (id) => document.getElementById(id);

function pretty(v) {
  try { return JSON.stringify(v, null, 2); } catch { return String(v); }
}

function parseJSONOrEmpty(text) {
  const t = (text || "").trim();
  if (!t) return null;
  return JSON.parse(t);
}

function setAuthUI() {
  const mode = el("authmode").value;
  const box = el("authbox");
  box.innerHTML = "";

  if (mode === "bearer") {
    box.innerHTML = `
      <label>Token</label>
      <input id="bearer" placeholder="eyJhbGciOi..."/>
    `;
  } else if (mode === "basic") {
    box.innerHTML = `
      <div class="row">
        <div class="col">
          <label>User</label>
          <input id="basicuser" placeholder="user"/>
        </div>
        <div class="col">
          <label>Password</label>
          <input id="basicpass" type="password" placeholder="••••••"/>
        </div>
      </div>
    `;
  }
}

function buildHeaders() {
  let headers = {};
  try {
    headers = parseJSONOrEmpty(el("headers").value) || {};
  } catch (e) {
    alert("Headers JSON is invalid: " + e.message);
    throw e;
  }

  const mode = el("authmode").value;
  if (mode === "bearer") {
    const tok = (document.getElementById("bearer")?.value || "").trim();
    if (tok) headers["Authorization"] = "Bearer " + tok;
  } else if (mode === "basic") {
    const u = (document.getElementById("basicuser")?.value || "");
    const p = (document.getElementById("basicpass")?.value || "");
    if (u || p) headers["Authorization"] = "Basic " + btoa(u + ":" + p);
  }
  return headers;
}

async function refreshHistory() {
  const r = await fetch("/history");
  const items = await r.json();
  el("history").innerHTML = items.slice().reverse().map(v =>
    `<div class="hist">
      <div class="tiny">${escapeHtml(v.at)}</div>
      <div><b>${escapeHtml(v.method)}</b> <span class="muted">${escapeHtml(String(v.status))}</span></div>
      <div class="click" data-url="${escapeHtml(v.url)}">${escapeHtml(v.url)}</div>
    </div>`
  ).join("");

  el("history").querySelectorAll(".click").forEach(n => {
    n.addEventListener("click", () => {
      el("url").value = n.getAttribute("data-url");
      el("method").value = "GET";
    });
  });
}

function escapeHtml(s) {
  return (s || "")
    .replaceAll("&","&amp;")
    .replaceAll("<","&lt;")
    .replaceAll(">","&gt;")
    .replaceAll('"',"&quot;");
}

function renderHints(hints) {
  if (!hints || hints.length === 0) {
    el("hints").innerHTML = `<div class="muted">No obvious next steps found.</div>`;
    return;
  }
  el("hints").innerHTML = hints.map(h =>
    `<div class="hint" data-url="${escapeHtml(h)}">${escapeHtml(h)}</div>`
  ).join("");
  el("hints").querySelectorAll(".hint").forEach(n => {
    n.addEventListener("click", () => {
      el("url").value = n.getAttribute("data-url");
      el("method").value = "GET";
    });
  });
}

async function send() {
  const url = el("url").value.trim();
  if (!url) return;

  let body = null;
  try {
    body = parseJSONOrEmpty(el("body").value);
  } catch (e) {
    alert("Body JSON is invalid: " + e.message);
    return;
  }

  const payload = {
    method: el("method").value,
    url,
    headers: buildHeaders(),
    body
  };

  el("meta").textContent = "Loading…";
  el("json").textContent = "";
  el("raw").textContent = "";

  const r = await fetch("/proxy", {
    method: "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify(payload)
  });

  const data = await r.json();

  el("meta").textContent = `HTTP ${data.status}`;
  el("json").textContent = data.body ? pretty(data.body) : "(not json)";
  el("raw").textContent = data.raw || "";

  renderHints(data.hints);
  refreshHistory();
}

el("authmode").addEventListener("change", setAuthUI);
el("send").addEventListener("click", send);

setAuthUI();
refreshHistory();
