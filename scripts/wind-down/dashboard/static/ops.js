// Ops tab: the dated wind-down checklist. Unlike the other tabs it has no collector snapshot: it polls
// /api/ops ({plan, status, today}) itself and ticks through POST /api/ops/check.
(() => {

const OPS_POLL_MS = 60 * 1000;
const NAME_KEY = 'ops-name';
const ESTIMATED_HINT = 'Estimated: this date follows from unbonding times, not from a decision.';

let plan = null;
let status = {};
let today = '';
let lastBody = ''; // the raw /api/ops answer last drawn, so an unchanged poll does not re-render
const sectionOpen = new Map(); // block key -> open, once the user has toggled it by hand
const detailsOpen = new Set(); // step ids whose "details" the user opened

registerSelfPollingTab('ops', startOps);

function startOps() {
  const root = document.getElementById('view-ops');
  root.innerHTML = `<div class="panel ops-bar"><div id="opsHeadline"></div>
      <label class="ops-name">you are <input id="opsName" type="text" maxlength="80" placeholder="your name" value="${escapeHtml(savedName())}"></label></div>
    <div id="opsMessage"></div><div id="opsBlocks"></div>`;

  document.getElementById('opsName').oninput = (event) => saveName(event.target.value);
  root.addEventListener('change', onTick);
  // `toggle` does not bubble, so listen in the capture phase to remember what the user opened or closed.
  root.addEventListener('toggle', onToggle, true);

  pollOps();
  setInterval(pollOps, OPS_POLL_MS);
}

async function pollOps() {
  const response = await fetch('/api/ops').catch(() => null);
  const raw = response ? await response.text() : '';
  if (!response || response.status !== 200) {
    showMessage(response ? `could not read the plan: ${errorText(raw)}` : 'server unreachable');
    return;
  }
  showMessage('');
  if (raw === lastBody) return;

  lastBody = raw;
  const body = JSON.parse(raw);
  ({ plan, status, today } = body);
  drawOps();
}

function errorText(raw) {
  try {
    return JSON.parse(raw).error;
  } catch (error) {
    return raw.slice(0, 200);
  }
}

function showMessage(text) {
  document.getElementById('opsMessage').innerHTML = text ? `<div class="errors">${escapeHtml(text)}</div>` : '';
}

// ---- ticking

async function onTick(event) {
  const box = event.target.closest('input[data-id]');
  if (!box) return;

  const by = document.getElementById('opsName').value.trim();
  if (!by) {
    box.checked = !box.checked;
    showMessage('Enter your name in the "you are" field before ticking.');
    return;
  }

  const response = await fetch('/api/ops/check', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id: box.dataset.id, done: box.checked, by }),
  }).catch(() => null);
  if (!response || response.status !== 200) {
    box.checked = !box.checked;
    showMessage(response ? `tick failed: ${errorText(await response.text())}` : 'server unreachable');
    return;
  }

  showMessage('');
  status = await response.json();
  lastBody = ''; // the stored answer is stale now; the next poll redraws from disk
  drawOps();
}

function onToggle(event) {
  const target = event.target;
  if (target.dataset.block) sectionOpen.set(target.dataset.block, target.open);
  if (target.dataset.detail) target.open ? detailsOpen.add(target.dataset.detail) : detailsOpen.delete(target.dataset.detail);
}

function savedName() {
  try {
    return localStorage.getItem(NAME_KEY) || '';
  } catch (error) {
    return '';
  }
}

function saveName(name) {
  try {
    localStorage.setItem(NAME_KEY, name);
  } catch (error) {
    // private mode: the name just lasts until reload
  }
}

// ---- page

function drawOps() {
  const blocks = [...plan.days].sort((first, second) => first.date.localeCompare(second.date));
  const current = blocks.filter((block, index) => isCurrent(block, blocks[index + 1]));
  const next = blocks.find((block) => block.date > today);

  document.getElementById('opsHeadline').innerHTML = headline(current[current.length - 1], next);
  document.getElementById('opsBlocks').innerHTML = blocks
    .map((block) => blockHtml(block, { current: current.includes(block), next: block === next }))
    .join('');
}

// A block is current from its date until the next block starts, or through `until` for a multi-day block.
function isCurrent(block, followingBlock) {
  if (block.date > today) return false;
  if (block.until && today <= block.until) return true;
  return !followingBlock || followingBlock.date > today;
}

function daysBetween(fromDate, toDate) {
  return Math.round((Date.parse(`${toDate}T00:00:00Z`) - Date.parse(`${fromDate}T00:00:00Z`)) / 86400000);
}

function headline(current, next) {
  const now = current ? `now: <b>${escapeHtml(current.title)}</b>` : 'now: <span class="muted">nothing scheduled</span>';
  if (!next) return `<b>${today}</b> · ${now} · <span class="muted">no later block</span>`;

  const days = daysBetween(today, next.date);
  const countdown = days === 1 ? 'tomorrow' : `in ${days} days`;
  return `<b>${today}</b> · ${now} · next: <b>${escapeHtml(next.title)}</b> <span class="muted">${datesLabel(next)}, ${countdown}</span>`;
}

function datesLabel(block) {
  const range = block.until ? `${block.date} → ${block.until}` : block.date;
  return block.estimated ? `<span title="${ESTIMATED_HINT}">~${range}</span>` : range;
}

function stepProgress(block) {
  const steps = block.windows.flatMap((window) => window.steps);
  return { done: steps.filter(stepDone).length, total: steps.length };
}

// A step with zones is done when every zone is; its own id is not ticked directly.
function stepDone(step) {
  if (step.zones && step.zones.length) return step.zones.every((zone) => isDone(`${step.id}:${zone}`));
  return isDone(step.id);
}

function isDone(id) {
  return Boolean(status[id] && status[id].done);
}

function blockHtml(block, { current, next }) {
  const key = `${block.date}|${block.title}`;
  const progress = stepProgress(block);
  const complete = progress.total > 0 && progress.done === progress.total;
  // Default: past and later blocks collapsed, current and next open; a manual toggle wins from then on.
  const open = sectionOpen.has(key) ? sectionOpen.get(key) : current || next;
  const marker = current ? pill('warn', 'now') : next ? pill('idle', 'next') : '';
  const avoid = (block.avoid || []).length
    ? `<div class="ops-avoid"><b>Do not</b><ul>${block.avoid.map((item) => `<li>${escapeHtml(item)}</li>`).join('')}</ul></div>`
    : '';

  return `<details class="panel ops-block ${current ? 'current' : ''}" data-block="${escapeHtml(key)}" ${open ? 'open' : ''}>
    <summary><h2>${datesLabel(block)} <span>${escapeHtml(block.title)}</span> ${marker}
      <span class="sub">${escapeHtml(block.note || '')}</span>
      <span class="ops-count">${pill(complete ? 'ok' : 'idle', `${progress.done}/${progress.total}`)}</span></h2></summary>
    ${avoid}${block.windows.map(windowHtml).join('')}</details>`;
}

function windowHtml(window) {
  const label = window.label ? `<div class="ops-window">${escapeHtml(window.label)}</div>` : '';
  return `${label}<ul class="ops-steps">${window.steps.map(stepHtml).join('')}</ul>`;
}

function stepHtml(step) {
  const zones = step.zones || [];
  const tag = step.conditional ? pill('warn', step.conditional) : '';
  const ref = step.ref ? `<span class="muted mono">${escapeHtml(step.ref)}</span>` : '';
  const detail = step.detail
    ? `<details class="ops-detail" data-detail="${escapeHtml(step.id)}" ${detailsOpen.has(step.id) ? 'open' : ''}>
        <summary class="muted">details</summary><div class="note">${escapeHtml(step.detail)}</div></details>`
    : '';
  const command = step.command
    ? `<pre class="ops-command mono copy" data-copy="${escapeHtml(step.command)}" title="click to copy">${escapeHtml(step.command)}</pre>`
    : '';
  const expect = step.expect ? `<div class="muted ops-expect">expect: ${escapeHtml(step.expect)}</div>` : '';
  const checkbox = zones.length
    ? `<input type="checkbox" disabled ${stepDone(step) ? 'checked' : ''} title="done when every zone is">`
    : `<input type="checkbox" data-id="${escapeHtml(step.id)}" ${isDone(step.id) ? 'checked' : ''}>`;
  const zoneRows = zones.length
    ? `<ul class="ops-zones">${zones.map((zone) => zoneHtml(`${step.id}:${zone}`, zone)).join('')}</ul>`
    : '';

  return `<li class="ops-step ${stepDone(step) ? 'done' : ''}">
    <div class="ops-row"><label>${checkbox} <span class="ops-text">${escapeHtml(step.text)}</span></label>
      ${tag} ${ref} ${zones.length ? '' : tickedBy(step.id)}</div>${command}${expect}${detail}${zoneRows}</li>`;
}

function zoneHtml(id, zone) {
  return `<li class="ops-row"><label><input type="checkbox" data-id="${escapeHtml(id)}" ${isDone(id) ? 'checked' : ''}> ${escapeHtml(zone)}</label>
    ${tickedBy(id)}</li>`;
}

function tickedBy(id) {
  if (!isDone(id)) return '';
  const tick = status[id];
  return `<span class="muted ops-by" title="${escapeHtml(tick.at)}">${escapeHtml(tick.by)} · ${formatRelative(tick.at)}</span>`;
}
})();
