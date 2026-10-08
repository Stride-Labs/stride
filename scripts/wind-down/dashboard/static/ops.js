// Ops tab: the dated wind-down checklist. Unlike the other tabs it has no collector snapshot: it polls
// /api/ops ({plan, status, today}) itself and ticks through POST /api/ops/check. A step's `auto` check reads another
// tab's snapshot live, but only once its window has started (the window's `start`, else its day): before that the
// marks read n/a. A step's `multisig` names a tx set on the Multisig tab (`<set-id>`, or `<set-id>/<zone>` for one
// zone's heading in it) and renders as a link; a step with zones and a bare set id links to each zone's heading.
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
const foldOpen = new Set(); // step ids whose "more" fold (rest of the text, command, expect, details) is open

registerSelfPollingTab('ops', startOps);

function startOps() {
  const root = document.getElementById('view-ops');
  root.innerHTML = `<div class="panel ops-bar"><div id="opsHeadline"></div>
      <label class="ops-name">you are <input id="opsName" type="text" maxlength="80" placeholder="your name" value="${escapeHtml(savedName())}"></label></div>
    <div id="opsMessage"></div><div id="opsBlocks"></div>`;

  document.getElementById('opsName').oninput = (event) => saveName(event.target.value);
  root.addEventListener('change', onTick);
  root.addEventListener('click', onFold);
  // `toggle` does not bubble, so listen in the capture phase to remember what the user opened or closed.
  root.addEventListener('toggle', onToggle, true);

  pollOps();
  setInterval(pollOps, OPS_POLL_MS);
  pollLive();
  setInterval(pollLive, OPS_POLL_MS);
}

// Live snapshots the plan's `auto` checks read ({channels, validators, funds}: {fetched_at, data}), plus `multisig`
// (its tx sets, for the titles of the steps' "→ Multisig tab" links) when the plan references a set.
const live = {};

async function pollLive() {
  const tabs = new Set(liveSources());
  let changed = false;
  for (const tab of tabs) {
    const response = await fetch(`/api/${tab}`).catch(() => null);
    if (!response || response.status !== 200) continue;
    const body = await response.json();
    if (live[tab] && live[tab].fetched_at === body.fetched_at) continue;
    live[tab] = { fetched_at: body.fetched_at, data: body.data };
    changed = true;
  }
  if (changed && plan) drawOps();
}

function liveSources() {
  if (!plan) return [];
  const steps = plan.days.flatMap((day) => day.windows.flatMap((window) => window.steps));
  const autoTabs = steps.filter((step) => step.auto).map((step) => step.auto.tab);
  return steps.some((step) => step.multisig) ? [...autoTabs, 'multisig'] : autoTabs;
}

// The title of a multisig tx set, once /api/multisig has answered; the set id until then.
function multisigSetTitle(setId) {
  const set = live.multisig && live.multisig.data.sets.find((candidate) => candidate.id === setId);
  return set ? set.title : setId;
}

// The live value behind a step's `auto` check for one zone: true / false / null (n/a) / undefined (no snapshot yet).
// The rows a live check reports on: the step's zones, or, for a check on a Multisig set (`auto.set`), that set's bundles
// (by their "bundle N" label), optionally only the bundles with a member matching `auto.members`.
function autoEntries(step) {
  if (!step.auto || !step.auto.set) return step.zones || [];
  const txs = bundleTxs(step);
  return txs ? txs.map(bundleLabel) : [];
}

function bundleTxs(step) {
  const snapshot = live.multisig;
  if (!snapshot) return null;
  const set = snapshot.data.sets.find((candidate) => candidate.id === step.auto.set);
  if (!set) return [];
  return set.txs.filter((tx) => tx.members && tx.members.length && (!step.auto.members || tx.members.some((member) => member.includes(step.auto.members))));
}

function bundleLabel(tx) {
  return tx.title.split(' · ')[0];
}

function autoValue(step, zone) {
  if (step.auto.set) return bundleValue(step, zone);
  const snapshot = live[step.auto.tab];
  if (!snapshot) return undefined;
  const entry = snapshot.data.zones.find((candidate) => candidate.chain_id === zone);
  if (!entry || entry.error) return null;
  const value = step.auto.path.split('.').reduce((inner, key) => (inner == null ? inner : inner[key]), entry);
  if (value === null || value === undefined) return null;
  // `equals` / `at_least` turn a number (an over count, a record count, a drained count) into a pass/fail; without
  // either the value is a boolean already.
  if (step.auto.equals === undefined && step.auto.at_least === undefined) return value;
  const passes = step.auto.at_least === undefined ? String(value) === String(step.auto.equals) : Number(value) >= step.auto.at_least;
  if (passes) return true;
  return (step.auto.allow || []).includes(zone) ? 'allowed' : false;
}

// A bundle's value for the check: 'blocked' while the bundle waits on something outside our hands (it then stays out
// of the step's count), else its `seeded` / `done` flag.
function bundleValue(step, label) {
  const txs = bundleTxs(step);
  if (!txs) return undefined;
  const tx = txs.find((candidate) => bundleLabel(candidate) === label);
  if (!tx) return null;
  if (tx.blocked) return 'blocked';
  const value = tx[step.auto.path];
  return value === null || value === undefined ? null : value;
}

function autoMark(value) {
  if (value === undefined) return '';
  if (value === true) return `<span class="ops-auto ok" title="verified live on the dashboard">✓ ok</span>`;
  if (value === 'allowed') return `<span class="ops-auto ok" title="not clean, but expected here (see the step text)">✓ allowed</span>`;
  if (value === false) return `<span class="ops-auto bad" title="the live check fails">✗ fail</span>`;
  if (value === 'blocked') return `<span class="ops-auto muted" title="waits on something outside our hands; not counted">blocked</span>`;
  return `<span class="ops-auto muted" title="could not be checked">n/a</span>`;
}

// A live check before its window has started reads n/a; the title says when it applies and what it would read now.
function inactiveMark(value, gate) {
  const wouldRead = value === undefined ? '' : `; would read ${markText(value)}`;
  return `<span class="ops-auto muted" title="applies from ${escapeHtml(gate.label)}${escapeHtml(wouldRead)}">n/a</span>`;
}

function markText(value) {
  if (value === true) return '✓ ok';
  if (value === 'allowed') return '✓ allowed';
  if (value === false) return '✗ fail';
  return 'n/a';
}

function autoSummary(step, gate) {
  if (!gate.active) return `<span class="ops-auto muted" title="applies from ${escapeHtml(gate.label)}">n/a</span>`;
  const values = autoEntries(step).map((zone) => autoValue(step, zone)).filter((value) => value !== 'blocked');
  if (!values.length || values.some((value) => value === undefined)) return '';
  const passing = values.filter((value) => value === true || value === 'allowed').length;
  const cssClass = passing === values.length ? 'ok' : 'bad';
  return `<span class="ops-auto ${cssClass}" title="from the ${escapeHtml(step.auto.tab)} tab's latest snapshot">${passing}/${values.length} ok</span>`;
}

// Whether a window's live checks apply yet: from the window's `start` (ISO UTC) when it has one, else from its day
// (the server's ET date). Activation never ends, so a check stays live once its window has started.
function liveGate(day, window) {
  if (window.start) return { active: Date.now() >= Date.parse(window.start), label: window.label };
  return { active: today >= day.date, label: day.date };
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
  pollLive(); // the plan names the tabs the live checks read, so fetch them as soon as it is known
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

// The "more / less" toggle at the end of a step's lead line; a redraw is cheap and keeps one source of truth.
function onFold(event) {
  const button = event.target.closest('button[data-fold]');
  if (!button) return;
  foldOpen.has(button.dataset.fold) ? foldOpen.delete(button.dataset.fold) : foldOpen.add(button.dataset.fold);
  drawOps();
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

// A step with zones is done when every zone is; its own id is not ticked directly. An auto-checked step is the
// exception: the dashboard verifies each zone live, so the step carries one tick of its own.
function stepDone(step) {
  if (step.zones && step.zones.length && !step.auto) return step.zones.every((zone) => isDone(`${step.id}:${zone}`));
  return isDone(step.id);
}

function isDone(id) {
  return Boolean(status[id] && status[id].done);
}

function blockHtml(block, { current, next }) {
  const key = `${block.date}|${block.title}`;
  const progress = stepProgress(block);
  const complete = progress.total > 0 && progress.done === progress.total;
  // Default: later blocks collapsed, current and next open, and a past block open while it still has unticked
  // steps (they are overdue, not history); a manual toggle wins from then on.
  const past = !current && !next && block.date < today;
  const overdue = past && !complete;
  const open = sectionOpen.has(key) ? sectionOpen.get(key) : current || next || overdue;
  const marker = current ? pill('warn', 'now') : next ? pill('idle', 'next') : overdue ? pill('bad', `${progress.total - progress.done} overdue`) : '';
  const avoid = (block.avoid || []).length
    ? `<div class="ops-avoid"><b>Do not</b><ul>${block.avoid.map((item) => `<li>${escapeHtml(item)}</li>`).join('')}</ul></div>`
    : '';

  return `<details class="panel ops-block ${current ? 'current' : ''}" data-block="${escapeHtml(key)}" ${open ? 'open' : ''}>
    <summary><h2>${datesLabel(block)} <span>${escapeHtml(block.title)}</span> ${marker}
      <span class="sub">${escapeHtml(block.note || '')}</span>
      <span class="ops-count">${pill(complete ? 'ok' : 'idle', `${progress.done}/${progress.total}`)}</span></h2></summary>
    ${avoid}${block.windows.map((window) => windowHtml(block, window)).join('')}</details>`;
}

function windowHtml(day, window) {
  const label = window.label ? `<div class="ops-window">${escapeHtml(window.label)}</div>` : '';
  const gate = liveGate(day, window);
  return `${label}<ul class="ops-steps">${window.steps.map((step) => stepHtml(step, gate)).join('')}</ul>`;
}

// A step shows its lead sentence on the checkbox line; the rest of the text, the command, the expectation and
// the details sit behind a "more" fold so the plan reads as a list, not a wall. A done step collapses to its lead
// line and who ticked it: it is history, and unticking it is still possible from the box.
function stepHtml(step, gate) {
  const zones = step.zones || [];
  const perZoneTicks = zones.length > 0 && !step.auto;
  const { lead, rest } = splitLead(step.text);
  const checkbox = perZoneTicks
    ? `<input type="checkbox" disabled ${stepDone(step) ? 'checked' : ''} title="done when every zone is">`
    : `<input type="checkbox" data-id="${escapeHtml(step.id)}" ${isDone(step.id) ? 'checked' : ''}>`;
  const label = `<label>${checkbox} <span class="ops-text">${escapeHtml(lead)}</span></label>`;
  if (stepDone(step)) return `<li class="ops-step done"><div class="ops-row">${label} ${doneBy(step)}</div></li>`;

  const tag = step.conditional ? pill('warn', step.conditional) : '';
  const entries = autoEntries(step);
  const auto = step.auto && entries.length ? autoSummary(step, gate) : '';
  const open = foldOpen.has(step.id);
  const fold = foldHtml(step, rest);
  // No "more" when there is nothing behind it: a one-sentence step with no command, expectation or details.
  const more = fold ? `<button type="button" class="ops-more" data-fold="${escapeHtml(step.id)}">${open ? 'less ▾' : 'more ▸'}</button>` : '';
  const multisig = step.multisig ? multisigLinks(step) : '';
  const zoneMark = (zone) => (gate.active ? autoMark(autoValue(step, zone)) : inactiveMark(autoValue(step, zone), gate));
  const zoneRows = perZoneTicks
    ? `<ul class="ops-zones">${zones.map((zone) => zoneHtml(`${step.id}:${zone}`, zone)).join('')}</ul>`
    : step.auto && entries.length
      ? `<ul class="ops-zones">${entries.map((zone) => `<li class="ops-row">${escapeHtml(zone)} ${zoneMark(zone)}</li>`).join('')}</ul>`
      : '';

  return `<li class="ops-step">
    <div class="ops-row">${label} ${tag} ${auto} ${more}</div>
    ${fold ? `<div class="ops-fold" ${open ? '' : 'hidden'}>${fold}</div>` : ''}${multisig}${zoneRows}</li>`;
}

// The lead sentence (up to the first ". ", keeping the period) and the rest; a text without a break is all lead.
function splitLead(text) {
  const index = text.indexOf('. ');
  if (index === -1) return { lead: text, rest: '' };
  return { lead: text.slice(0, index + 1), rest: text.slice(index + 2) };
}

function foldHtml(step, rest) {
  const text = rest ? `<div class="ops-rest">${escapeHtml(rest)}</div>` : '';
  const command = step.command
    ? `<pre class="ops-command mono copy" data-copy="${escapeHtml(step.command)}" title="click to copy">${escapeHtml(step.command)}</pre>`
    : '';
  const expect = step.expect ? `<div class="muted ops-expect">expect: ${escapeHtml(step.expect)}</div>` : '';
  const detail = step.detail
    ? `<details class="ops-detail" data-detail="${escapeHtml(step.id)}" ${detailsOpen.has(step.id) ? 'open' : ''}>
        <summary class="muted">details</summary><div class="note">${escapeHtml(step.detail)}</div></details>`
    : '';
  return `${text}${command}${expect}${detail}`;
}

// Who finished a step: its own tick, or for a per-zone step the latest zone tick (the one that completed it).
function doneBy(step) {
  const perZoneTicks = step.zones && step.zones.length && !step.auto;
  if (!perZoneTicks) return tickedBy(step.id);
  const ids = step.zones.map((zone) => `${step.id}:${zone}`);
  const latest = ids.reduce((best, id) => (!best || status[id].at > status[best].at ? id : best), null);
  return tickedBy(latest);
}

// `<set-id>/<zone>` links to that zone's heading in the set; a bare set id links to each of the step's zones' headings,
// or to the set itself when the step has no zones.
function multisigLinks(step) {
  const [setId, zone] = step.multisig.split('/');
  const zones = zone ? [zone] : step.zones && step.zones.length ? step.zones.filter((candidate) => setHasZone(setId, candidate)) : [];
  return `<div class="ops-link">${(zones.length ? zones : [null]).map((target) => multisigLink(setId, target)).join('')}</div>`;
}

// A set whose txs are not grouped by zone (the pool-creation bundles span every zone) gets one link to the set.
function setHasZone(setId, zone) {
  const set = live.multisig && live.multisig.data.sets.find((candidate) => candidate.id === setId);
  return Boolean(set) && set.txs.some((tx) => tx.chain_id === zone);
}

function multisigLink(setId, zone) {
  const href = zone ? `#multisig/${setId}/${zone}` : `#multisig/${setId}`;
  const label = zone ? `${multisigSetTitle(setId)} · ${zone}` : multisigSetTitle(setId);
  return `<a href="${escapeHtml(href)}">→ Multisig tab: ${escapeHtml(label)}</a>`;
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
