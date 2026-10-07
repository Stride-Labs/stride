// Multisig tab: the F5 protocol-admin multisig's tx sets (drains, ICA transfers, pool funding), each grouped by zone
// under a heading (anchor `set-<set-id>-<chain_id>`, the target of `#multisig/<set-id>/<zone>`) with every tx's
// generate / sign / multisign+broadcast commands and who runs each. Like Ops it has no collector: it polls
// /api/multisig ({fetched_at, data: {sets}}), which the server composes from the collectors' snapshots and the plan.
(() => {

const MULTISIG_POLL_MS = 30 * 1000;
const WHO_CLASS = { Sam: 'who-sam', Aidan: 'who-aidan', Riley: 'who-riley' }; // "anyone" keeps the neutral pill

let lastBody = ''; // the raw /api/multisig answer last drawn, so an unchanged poll does not re-render
let drawn = false; // the first draw honours the `#multisig/<set-id>[/<zone>]` hash by scrolling to that set or zone
const closedSets = new Set(); // set ids the user collapsed, so a redraw keeps them collapsed

registerSelfPollingTab('multisig', startMultisig);

function startMultisig() {
  const root = document.getElementById('view-multisig');
  root.innerHTML = '<div id="multisigMessage"></div><div id="multisigSets"></div>';
  // `toggle` does not bubble, so listen in the capture phase to remember what the user closed.
  root.addEventListener('toggle', onToggle, true);
  // A link from the Ops tab lands here after the first draw, so the hash is honoured on every change too.
  window.addEventListener('hashchange', scrollToHashedSet);

  pollMultisig();
  setInterval(pollMultisig, MULTISIG_POLL_MS);
}

async function pollMultisig() {
  const response = await fetch('/api/multisig').catch(() => null);
  const raw = response ? await response.text() : '';
  if (!response || response.status !== 200) {
    showMessage(response ? `could not read the tx sets: ${errorText(raw)}` : 'server unreachable');
    return;
  }
  showMessage('');
  if (raw === lastBody) return;

  lastBody = raw;
  drawSets(JSON.parse(raw));
}

function errorText(raw) {
  try {
    return JSON.parse(raw).error;
  } catch (error) {
    return raw.slice(0, 200);
  }
}

function showMessage(text) {
  document.getElementById('multisigMessage').innerHTML = text ? `<div class="errors">${escapeHtml(text)}</div>` : '';
}

function onToggle(event) {
  const target = event.target;
  if (!target.dataset.set) return;
  if (target.open) closedSets.delete(target.dataset.set);
  else closedSets.add(target.dataset.set);
}

// ---- page

function drawSets(body) {
  document.getElementById('multisigSets').innerHTML = body.data.sets.map(setHtml).join('');
  if (drawn) return;
  drawn = true;
  scrollToHashedSet();
}

// `#multisig/<set-id>` scrolls to the set, `#multisig/<set-id>/<zone>` to that zone's heading in it (the set first,
// when the zone is not in it).
function scrollToHashedSet() {
  const [tab, setId, zone] = location.hash.slice(1).split('/');
  if (tab !== 'multisig' || !setId) return;
  const set = document.getElementById(`set-${setId}`);
  if (!set) return;
  set.open = true;
  const heading = zone ? document.getElementById(`set-${setId}-${zone}`) : null;
  (heading || set).scrollIntoView();
}

function setHtml(set) {
  const zones = txsByZone(set.txs);
  return `<details class="panel" id="set-${escapeHtml(set.id)}" data-set="${escapeHtml(set.id)}" ${closedSets.has(set.id) ? '' : 'open'}>
    <summary><h2>${escapeHtml(set.title)} <span class="sub">${zones.length} zones · ${set.txs.length} txs</span>
      <span class="ops-count">${readyPill(set.txs)}</span></h2></summary>
    <div class="ms-intro">${escapeHtml(set.description)}<a href="#ops">→ Ops step</a></div>
    ${zones.map(([chainId, txs]) => zoneHtml(set.id, chainId, txs)).join('')}</details>`;
}

// [chain id, its txs] in the order the server lists them (config.ZONES order, txs in submission order within a zone).
function txsByZone(txs) {
  const groups = new Map();
  txs.forEach((tx) => {
    if (!groups.has(tx.chain_id)) groups.set(tx.chain_id, []);
    groups.get(tx.chain_id).push(tx);
  });
  return [...groups];
}

function readyPill(txs) {
  const ready = txs.filter((tx) => tx.ready).length;
  return pill(ready === txs.length ? 'ok' : 'idle', `${ready}/${txs.length} ready`);
}

// A zone's heading carries the anchor the Ops tab's `<set-id>/<zone>` links point at.
function zoneHtml(setId, chainId, txs) {
  return `<div class="ms-zone" id="set-${escapeHtml(setId)}-${escapeHtml(chainId)}">
    <div class="ms-zone-head"><b>${escapeHtml(chainId)}</b> ${readyPill(txs)}</div>
    ${txs.map(txHtml).join('')}</div>`;
}

function txHtml(tx) {
  const state = tx.ready ? pill('ok', 'ready') : pill('warn', 'not ready');
  const reason = tx.reason ? `<span class="muted">${escapeHtml(tx.reason)}</span>` : '';
  const files = tx.files.length
    ? `<div class="ms-files muted">share between people (Slack): ${tx.files.map((file) => `<span class="mono">${escapeHtml(file)}</span>`).join(', ')}</div>`
    : '';
  return `<div class="ms-tx">
    <div class="ms-tx-head"><b>${escapeHtml(tx.title)}</b> ${state} ${reason}</div>${files}
    ${tx.commands.map(commandHtml).join('')}</div>`;
}

function commandHtml(command) {
  return `<div class="ms-label">${whoPill(command.tag)} <span>${escapeHtml(command.label)}</span></div>
    <pre class="ms-command mono copy" data-copy="${escapeHtml(command.text)}" title="click to copy">${escapeHtml(command.text)}</pre>`;
}

function whoPill(tag) {
  return `<span class="who ${WHO_CLASS[tag] || ''}">${escapeHtml(tag)}</span>`;
}
})();
