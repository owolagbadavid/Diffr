const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);
const show = (el) => el.classList.remove("hidden");
const hide = (el) => el.classList.add("hidden");

let currentUser = null;
let currentOwner = "";
let currentRepo = "";
let allRepos = [];
let strategies = [];
const fileCache = {};

// --- Boot ---

async function init() {
  // Load strategies
  try {
    const res = await fetch("/api/strategies");
    strategies = await res.json();
    const sel = $("#strategy-select");
    strategies.forEach((s) => {
      const opt = document.createElement("option");
      opt.value = s.name;
      opt.textContent = `${s.name} — ${s.description}`;
      sel.appendChild(opt);
    });
  } catch {}

  // Check auth
  try {
    const res = await fetch("/api/user");
    const data = await res.json();
    if (data.logged_in) {
      currentUser = data;
      renderUserArea();
    }
  } catch {}

  // Route from hash
  if (location.hash.length > 1) {
    const parts = location.hash.slice(1).split("/");
    if (parts.length >= 2) {
      currentOwner = parts[0];
      currentRepo = parts[1];
      if (parts.length === 3) {
        await navigateToPRDetail(parseInt(parts[2], 10));
        return;
      }
      await navigateToPRs();
      return;
    }
  }

  // Default view
  if (currentUser) {
    await navigateToRepos();
  } else {
    navigateToLanding();
  }
}

// --- Navigation ---

function hideAllSections() {
  ["#landing", "#repo-list", "#pr-list", "#pr-detail"].forEach((s) => hide($(s)));
  hideError();
}

function navigateToLanding() {
  hideAllSections();
  show($("#landing"));
  renderBreadcrumbs([]);
  location.hash = "";
}

async function navigateToRepos() {
  hideAllSections();
  showLoading();
  renderBreadcrumbs([{ label: "Repos" }]);

  try {
    const res = await fetch("/api/user/repos");
    if (!res.ok) throw new Error((await res.json()).error || res.statusText);
    allRepos = await res.json();
    renderRepoList(allRepos);
  } catch (err) {
    showError(err.message);
  } finally {
    hideLoading();
  }
  location.hash = "";
}

async function navigateToPRs() {
  hideAllSections();
  showLoading();
  renderBreadcrumbs([
    { label: "Repos", action: () => navigateToRepos() },
    { label: `${currentOwner}/${currentRepo}` },
  ]);

  try {
    const res = await fetch(`/api/repos/${currentOwner}/${currentRepo}/pulls`);
    if (!res.ok) throw new Error((await res.json()).error || res.statusText);
    const prs = await res.json();
    renderPRList(prs);
    location.hash = `${currentOwner}/${currentRepo}`;
  } catch (err) {
    showError(err.message);
  } finally {
    hideLoading();
  }
}

async function navigateToPRDetail(number) {
  hideAllSections();
  showLoading();
  renderBreadcrumbs([
    { label: "Repos", action: () => navigateToRepos() },
    { label: `${currentOwner}/${currentRepo}`, action: () => navigateToPRs() },
    { label: `#${number}` },
  ]);

  const strat = $("#strategy-select").value || "by-size";

  try {
    const res = await fetch(
      `/api/repos/${currentOwner}/${currentRepo}/pulls/${number}/files?strategy=${strat}`
    );
    if (!res.ok) throw new Error((await res.json()).error || res.statusText);
    const data = await res.json();
    renderPRDetail(data, number);
    location.hash = `${currentOwner}/${currentRepo}/${number}`;
  } catch (err) {
    showError(err.message);
  } finally {
    hideLoading();
  }
}

// --- Events ---

$("#repo-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const val = $("#repo-input").value.trim();
  const parts = val.split("/");
  if (parts.length !== 2 || !parts[0] || !parts[1]) {
    showError("Enter a valid owner/repo (e.g. facebook/react)");
    return;
  }
  currentOwner = parts[0];
  currentRepo = parts[1];
  await navigateToPRs();
});

$("#repo-search").addEventListener("input", (e) => {
  const q = e.target.value.toLowerCase();
  const filtered = allRepos.filter(
    (r) =>
      r.full_name.toLowerCase().includes(q) ||
      (r.description || "").toLowerCase().includes(q) ||
      (r.language || "").toLowerCase().includes(q)
  );
  renderRepoTable(filtered);
});

$("#strategy-select").addEventListener("change", () => {
  const parts = location.hash.slice(1).split("/");
  if (parts.length === 3) {
    navigateToPRDetail(parseInt(parts[2], 10));
  }
});

$("#file-viewer-close").addEventListener("click", () => hide($("#file-viewer")));

$("#logo").addEventListener("click", () => {
  if (currentUser) navigateToRepos();
  else navigateToLanding();
});

// --- Rendering: chrome ---

function renderUserArea() {
  const area = $("#user-area");
  if (!currentUser) {
    area.innerHTML = '<a href="/auth/login" class="btn btn-sm">Sign in</a>';
    return;
  }
  area.innerHTML = `
    <img class="user-avatar" src="${esc(currentUser.avatar_url)}" alt="" />
    <span class="user-name">${esc(currentUser.login)}</span>
    <a href="/auth/logout" class="link-btn">Logout</a>
  `;
}

function renderBreadcrumbs(crumbs) {
  const nav = $("#breadcrumbs");
  nav.innerHTML = "";
  crumbs.forEach((c, i) => {
    if (i > 0) {
      const sep = document.createElement("span");
      sep.className = "crumb-sep";
      sep.textContent = "/";
      nav.appendChild(sep);
    }
    if (c.action) {
      const a = document.createElement("a");
      a.className = "crumb-link";
      a.textContent = c.label;
      a.href = "#";
      a.addEventListener("click", (e) => { e.preventDefault(); c.action(); });
      nav.appendChild(a);
    } else {
      const span = document.createElement("span");
      span.className = "crumb-current";
      span.textContent = c.label;
      nav.appendChild(span);
    }
  });
}

// --- Rendering: repo list ---

function renderRepoList(repos) {
  $("#repo-search").value = "";
  renderRepoTable(repos);
  show($("#repo-list"));
}

function renderRepoTable(repos) {
  const container = $("#repo-table");
  container.innerHTML = "";

  if (repos.length === 0) {
    container.innerHTML = '<p class="muted" style="padding:1rem">No repositories found.</p>';
    return;
  }

  repos.forEach((repo) => {
    const row = document.createElement("div");
    row.className = "repo-row";
    row.addEventListener("click", () => {
      const parts = repo.full_name.split("/");
      currentOwner = parts[0];
      currentRepo = parts[1];
      navigateToPRs();
    });
    row.innerHTML = `
      <img class="repo-avatar" src="${esc(repo.owner_avatar)}" alt="" />
      <div class="repo-info">
        <span class="repo-name">${esc(repo.full_name)}</span>
        ${repo.private ? '<span class="repo-badge private">Private</span>' : ""}
        <span class="repo-desc">${esc(repo.description || "")}</span>
      </div>
      <div class="repo-meta">
        ${repo.language ? `<span class="repo-lang">${esc(repo.language)}</span>` : ""}
        ${repo.stars > 0 ? `<span class="repo-stars">\u2605 ${repo.stars}</span>` : ""}
        <span class="muted">${timeAgo(repo.updated_at)}</span>
      </div>
    `;
    container.appendChild(row);
  });
}

// --- Rendering: PR list ---

function renderPRList(prs) {
  const container = $("#pr-table");
  container.innerHTML = "";
  $("#pr-list-title").textContent = `${currentOwner}/${currentRepo} — Pull Requests`;

  if (prs.length === 0) {
    show($("#pr-empty"));
  } else {
    hide($("#pr-empty"));
    prs.forEach((pr) => {
      const row = document.createElement("div");
      row.className = "pr-row";
      row.addEventListener("click", () => navigateToPRDetail(pr.number));
      row.innerHTML = `
        <img class="pr-avatar" src="${esc(pr.avatar_url)}" alt="" />
        <span class="pr-number">#${pr.number}</span>
        <div class="pr-info">
          <span class="pr-title">${esc(pr.title)}</span>
          <span class="pr-branch">${esc(pr.branch)}</span>
        </div>
        <span class="pr-meta">
          ${pr.draft ? '<span class="pr-draft">Draft</span>' : ""}
          <span>${esc(pr.user)}</span>
          <span class="muted">${timeAgo(pr.updated_at)}</span>
        </span>
      `;
      container.appendChild(row);
    });
  }
  show($("#pr-list"));
}

// --- Rendering: PR detail ---

function renderPRDetail(data, number) {
  $("#pr-detail-title").textContent = `#${number} — ${data.total_files} files`;

  const container = $("#file-groups");
  container.innerHTML = "";

  (data.groups || []).forEach((group) => {
    const groupDiv = document.createElement("div");
    groupDiv.className = "group";
    groupDiv.innerHTML = `<div class="group-header">${esc(group.name)} (${group.files.length})</div>`;

    group.files.forEach((f) => {
      const card = document.createElement("div");
      card.className = "file-card collapsed";

      const header = document.createElement("div");
      header.className = "file-header";
      header.innerHTML = `
        <span class="file-toggle">&#9654;</span>
        <span class="file-status ${f.status}">${f.status}</span>
        <span class="file-name">${esc(f.filename)}</span>
        <span class="file-stat pr-additions">+${f.additions}</span>
        <span class="file-stat pr-deletions">-${f.deletions}</span>
        ${f.contents_url ? `<button class="file-view-btn" data-url="${esc(f.contents_url)}" data-name="${esc(f.filename)}">View full</button>` : ""}
      `;
      header.addEventListener("click", (e) => {
        if (e.target.closest(".file-view-btn")) {
          const btn = e.target.closest(".file-view-btn");
          openFileViewer(btn.dataset.url, btn.dataset.name);
          return;
        }
        card.classList.toggle("collapsed");
      });

      const diff = document.createElement("div");
      diff.className = "file-diff";
      diff.appendChild(buildDiffView(f.patch, f.contents_url, f.status));

      card.appendChild(header);
      card.appendChild(diff);
      groupDiv.appendChild(card);
    });

    container.appendChild(groupDiv);
  });

  show($("#pr-detail"));
}

// --- Diff rendering with expandable gaps ---

function parseHunks(patch) {
  const lines = patch.split("\n");
  const hunks = [];
  let current = null;
  let oldLine = 0;
  let newLine = 0;

  for (const line of lines) {
    const m = line.match(/^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)/);
    if (m) {
      const oldStart = parseInt(m[1], 10);
      const oldCount = m[2] !== undefined ? parseInt(m[2], 10) : 1;
      const newStart = parseInt(m[3], 10);
      const newCount = m[4] !== undefined ? parseInt(m[4], 10) : 1;
      current = {
        header: line,
        oldStart,
        oldEnd: oldStart + oldCount - 1,
        newStart,
        newEnd: newStart + newCount - 1,
        lines: [],
      };
      hunks.push(current);
      oldLine = oldStart;
      newLine = newStart;
      continue;
    }
    if (!current) continue;

    if (line.startsWith("+")) {
      current.lines.push({ type: "add", oldLn: null, newLn: newLine, text: line });
      newLine++;
    } else if (line.startsWith("-")) {
      current.lines.push({ type: "del", oldLn: oldLine, newLn: null, text: line });
      oldLine++;
    } else {
      current.lines.push({ type: "ctx", oldLn: oldLine, newLn: newLine, text: line });
      oldLine++;
      newLine++;
    }
  }
  return hunks;
}

function buildDiffView(patch, rawURL, status) {
  const container = document.createElement("div");
  if (!patch) {
    container.innerHTML = '<div class="diff-empty">Binary file or no diff available</div>';
    return container;
  }

  const hunks = parseHunks(patch);
  const canExpand = rawURL && status !== "removed";
  let prevNewEnd = 0;
  let prevOldEnd = 0;

  hunks.forEach((hunk) => {
    const gapNewStart = prevNewEnd + 1;
    const gapNewEnd = hunk.newStart - 1;
    const gapOldStart = prevOldEnd + 1;
    const gapCount = gapNewEnd - gapNewStart + 1;

    if (gapCount > 0 && canExpand) {
      container.appendChild(createGapBar(gapNewStart, gapNewEnd, gapOldStart, rawURL));
    }

    const hunkHeader = document.createElement("div");
    hunkHeader.className = "diff-hunk-header";
    hunkHeader.textContent = hunk.header;
    container.appendChild(hunkHeader);

    const table = document.createElement("table");
    table.className = "diff-table";
    hunk.lines.forEach((l) => {
      const tr = document.createElement("tr");
      tr.className = l.type === "add" ? "diff-add" : l.type === "del" ? "diff-del" : "diff-ctx";
      tr.innerHTML = `<td class="diff-ln">${l.oldLn ?? ""}</td><td class="diff-ln">${l.newLn ?? ""}</td><td class="diff-code">${esc(l.text)}</td>`;
      table.appendChild(tr);
    });
    container.appendChild(table);

    prevNewEnd = hunk.newEnd;
    prevOldEnd = hunk.oldEnd;
  });

  if (hunks.length > 0 && canExpand) {
    container.appendChild(createGapBar(prevNewEnd + 1, -1, prevOldEnd + 1, rawURL));
  }

  return container;
}

function createGapBar(newStart, newEnd, oldStart, rawURL) {
  const bar = document.createElement("div");
  bar.className = "diff-gap-bar";
  bar.textContent =
    newEnd === -1
      ? `\u2195 Show remaining lines from line ${newStart}`
      : `\u2195 Show ${newEnd - newStart + 1} hidden lines (${newStart}\u2013${newEnd})`;

  bar.addEventListener("click", async () => {
    bar.textContent = "Loading...";
    const fullLines = await fetchFullFile(rawURL);
    if (!fullLines) { bar.textContent = "Failed to load file"; return; }

    const end = newEnd === -1 ? fullLines.length : newEnd;
    if (newStart > end) { bar.remove(); return; }

    const table = document.createElement("table");
    table.className = "diff-table";
    let oldLn = oldStart;
    for (let i = newStart; i <= end; i++) {
      const tr = document.createElement("tr");
      tr.className = "diff-ctx";
      tr.innerHTML = `<td class="diff-ln">${oldLn}</td><td class="diff-ln">${i}</td><td class="diff-code">${esc(fullLines[i - 1] ?? "")}</td>`;
      table.appendChild(tr);
      oldLn++;
    }
    bar.replaceWith(table);
  });

  return bar;
}

// --- Full file viewer ---

async function openFileViewer(rawURL, filename) {
  const viewer = $("#file-viewer");
  $("#file-viewer-title").textContent = filename;
  $("#file-viewer-content").innerHTML = '<div class="diff-empty">Loading...</div>';
  show(viewer);

  const lines = await fetchFullFile(rawURL);
  if (!lines) {
    $("#file-viewer-content").innerHTML = '<div class="diff-empty">Failed to load file</div>';
    return;
  }
  let html = '<table class="diff-table file-table">';
  lines.forEach((line, i) => {
    html += `<tr><td class="diff-ln">${i + 1}</td><td class="diff-code">${esc(line)}</td></tr>`;
  });
  html += "</table>";
  $("#file-viewer-content").innerHTML = html;
}

async function fetchFullFile(rawURL) {
  if (fileCache[rawURL]) return fileCache[rawURL];
  try {
    const res = await fetch(`/api/raw?url=${encodeURIComponent(rawURL)}`);
    if (!res.ok) return null;
    const text = await res.text();
    const lines = text.split("\n");
    fileCache[rawURL] = lines;
    return lines;
  } catch { return null; }
}

// --- Helpers ---

function timeAgo(dateStr) {
  if (!dateStr) return "";
  const diff = Math.floor((new Date() - new Date(dateStr)) / 1000);
  if (diff < 60) return "just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

function showLoading() { show($("#loading")); }
function hideLoading() { hide($("#loading")); }
function showError(msg) { const el = $("#error"); el.textContent = msg; show(el); }
function hideError() { hide($("#error")); }

function esc(s) {
  if (!s) return "";
  const d = document.createElement("div");
  d.textContent = s;
  return d.innerHTML;
}

init();
