// Page shell: tabs, snapshot polling, header, and the helpers every tab module shares.
//
// A tab module calls registerTab(name, render) once; render(data, root, snapshot) fills `root` with the
// tab's HTML whenever a new snapshot arrives. `data` is the collector's result, `snapshot` is the whole
// API body ({fetched_at, duration_seconds, refreshing, data}).

const TAB_NAMES = ['channels', 'validators', 'funds'];
const POLL_MS = 5000;
const STALE_AFTER_INTERVALS = 3;

const renderers = {};
const snapshots = {}; // tab name -> latest API body
const renderedAt = {}; // tab name -> fetched_at of the snapshot currently drawn
let intervals = { channels: 60, funds: 120, validators: 300 };
let activeTab = TAB_NAMES[0];

function registerTab(name, render) {
  renderers[name] = render;
}

// ---- shared formatters

function escapeHtml(text) {
  return String(text).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
}

// An amount from an exact integer string (or number) and the token's decimals: 1234500000, 6 -> "1,234.50".
function formatAmount(value, decimals, fractionDigits = 2) {
  if (value === null || value === undefined) return 'n/a';
  const raw = BigInt(value);
  const magnitude = raw < 0n ? -raw : raw;
  const scale = 10n ** BigInt(decimals);
  const whole = (magnitude / scale).toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  const fraction = (magnitude % scale).toString().padStart(decimals, '0').slice(0, fractionDigits);
  return `${raw < 0n ? '-' : ''}${whole}${fraction ? '.' + fraction : ''}`;
}

// "45s", "12m", "2h 41m", "5.2d"
function formatDuration(seconds) {
  if (seconds === null || seconds === undefined) return 'n/a';
  const total = Math.max(0, Math.round(seconds));
  if (total < 60) return `${total}s`;
  if (total < 3600) return `${Math.floor(total / 60)}m`;
  if (total < 24 * 3600) return `${Math.floor(total / 3600)}h ${Math.floor((total % 3600) / 60)}m`;
  return `${(total / 86400).toFixed(1)}d`;
}

// "4m ago" for an ISO timestamp, or "n/a" when the lookup had no result.
function formatRelative(iso, now = Date.now()) {
  if (!iso) return 'n/a';
  return `${formatDuration((now - Date.parse(iso)) / 1000)} ago`;
}

function pill(cssClass, text) {
  return `<span class="badge ${cssClass}">${escapeHtml(text)}</span>`;
}

const STATUS_CLASS = { 'closed': 'bad', 'handshake stuck': 'bad', 'stuck': 'bad', 'pending': 'warn', 'ok': 'ok' };

function statusPill(status) {
  return pill(STATUS_CLASS[status] || 'idle', status);
}

// A truncated address that copies the full one when clicked.
function addressCell(address, head = 10, tail = 6) {
  if (!address) return '<span class="muted">n/a</span>';
  const shown = address.length > head + tail + 1 ? `${address.slice(0, head)}…${address.slice(-tail)}` : address;
  return `<span class="mono copy" data-copy="${escapeHtml(address)}" title="${escapeHtml(address)} (click to copy)">${escapeHtml(shown)}</span>`;
}

// Entries that carry an "error" field, in any list of the tab's payload (zones, legs, routes, ...).
function countErrors(data) {
  return Object.values(data || {})
    .filter(Array.isArray)
    .reduce((count, list) => count + list.filter((item) => item && item.error).length, 0);
}

// ---- shell

async function startApp() {
  document.querySelectorAll('.tab').forEach((button) => {
    button.onclick = () => {
      selectTab(button.dataset.tab);
      location.hash = button.dataset.tab;
    };
  });
  document.getElementById('refreshButton').onclick = refreshActiveTab;
  selectTab(TAB_NAMES.includes(location.hash.slice(1)) ? location.hash.slice(1) : TAB_NAMES[0]);
  document.addEventListener('click', copyOnClick);

  TAB_NAMES.filter((name) => !renderers[name]).forEach((name) => {
    document.getElementById(`view-${name}`).innerHTML = '<div class="placeholder">not built</div>';
  });
  TAB_NAMES.filter((name) => renderers[name]).forEach((name) => {
    document.getElementById(`view-${name}`).innerHTML = '<div class="placeholder">loading…</div>';
  });

  intervals = { ...intervals, ...(await fetchJson('/api/config')).body?.intervals };
  TAB_NAMES.forEach(pollTab);
  setInterval(() => TAB_NAMES.forEach(pollTab), POLL_MS);
  setInterval(updateHeader, 1000);
}

function selectTab(name) {
  activeTab = name;
  document.querySelectorAll('.tab').forEach((button) => button.classList.toggle('on', button.dataset.tab === name));
  TAB_NAMES.forEach((tab) => document.getElementById(`view-${tab}`).classList.toggle('on', tab === name));
  updateHeader();
}

async function fetchJson(url, options) {
  const response = await fetch(url, options);
  return { status: response.status, body: await response.json().catch(() => null) };
}

async function pollTab(name) {
  if (!renderers[name]) return;

  const { status, body } = await fetchJson(`/api/${name}`).catch(() => ({ status: 0, body: null }));
  if (status === 200) {
    snapshots[name] = body;
    drawSnapshot(name, body);
  } else if (!snapshots[name]) {
    showWaiting(name, status, body);
  }
  updateHeader();
}

function drawSnapshot(name, body) {
  updateBadge(name, body);
  if (renderedAt[name] === body.fetched_at) return;

  const root = document.getElementById(`view-${name}`);
  try {
    renderers[name](body.data, root, body);
  } catch (error) {
    console.error(error);
    root.innerHTML = `<div class="errors">render failed: ${escapeHtml(error.message)}</div>`;
  }
  renderedAt[name] = body.fetched_at;
}

function showWaiting(name, status, body) {
  const detail = status === 0 ? 'server unreachable' : body && body.last_error ? `last refresh failed: ${body.last_error}` : '';
  document.getElementById(`view-${name}`).innerHTML =
    `<div class="placeholder">loading… ${escapeHtml(detail)}</div>`;
}

function updateBadge(name, body) {
  const badge = document.querySelector(`.tab[data-tab="${name}"] .badge`);
  const errors = countErrors(body.data);
  // `.badge` sets display, so the hidden attribute never applied; show the count and colour it instead.
  badge.hidden = false;
  badge.classList.toggle('bad', errors > 0);
  badge.classList.toggle('ok', errors === 0);
  badge.textContent = `${errors} error${errors === 1 ? '' : 's'}`;
}

function updateHeader() {
  const label = document.getElementById('snapshotAge');
  const button = document.getElementById('refreshButton');
  const body = snapshots[activeTab];

  label.classList.remove('stale');
  if (!renderers[activeTab]) {
    label.textContent = 'not built';
    button.disabled = true;
    return;
  }
  if (!body) {
    button.disabled = false;
    label.textContent = 'loading…';
    return;
  }
  button.disabled = body.refreshing;

  const ageSeconds = (Date.now() - Date.parse(body.fetched_at)) / 1000;
  const interval = intervals[activeTab];
  const stale = ageSeconds > STALE_AFTER_INTERVALS * interval;
  label.classList.toggle('stale', stale);
  label.textContent =
    `${stale ? 'STALE · ' : ''}snapshot ${formatDuration(ageSeconds)} ago · refresh took ${body.duration_seconds}s` +
    ` · auto-refresh ${interval}s${body.refreshing ? ' · refreshing…' : ''}`;
}

async function refreshActiveTab() {
  const name = activeTab;
  const before = snapshots[name] && snapshots[name].fetched_at;
  document.getElementById('refreshButton').disabled = true;
  await fetchJson(`/api/refresh/${name}`, { method: 'POST' });

  // Poll quickly until the new snapshot lands (or give up after a couple of minutes).
  for (let attempt = 0; attempt < 120; attempt++) {
    await new Promise((resolve) => setTimeout(resolve, 1000));
    await pollTab(name);
    if (snapshots[name] && snapshots[name].fetched_at !== before) break;
  }
  updateHeader();
}

function copyOnClick(event) {
  const target = event.target.closest('[data-copy]');
  if (!target) return;

  navigator.clipboard.writeText(target.dataset.copy);
  target.classList.add('copied');
  setTimeout(() => target.classList.remove('copied'), 800);
}
