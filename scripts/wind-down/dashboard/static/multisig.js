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
const openCards = new Set(); // tx cards the user opened, keyed `<set-id>/<chain_id>/<index>`
const shownCommands = new Set(); // command rows whose text is shown, keyed `<card key>/<row index>`
let lastSets = null; // the sets last drawn, so a card or row toggle can redraw without a fetch

registerSelfPollingTab('multisig', startMultisig);

function startMultisig() {
  const root = document.getElementById('view-multisig');
  root.innerHTML = '<div id="multisigMessage"></div><div id="multisigSets"></div>';
  // `toggle` does not bubble, so listen in the capture phase to remember what the user closed.
  root.addEventListener('toggle', onToggle, true);
  root.addEventListener('click', onCardClick);
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

// A card's title row opens or closes it; a row's "show" reveals its command text. Both redraw from the last answer,
// so one source of truth decides what is open. Copy buttons are handled by the page-wide copyOnClick.
function onCardClick(event) {
  const head = event.target.closest('[data-card]');
  const show = event.target.closest('button[data-show]');
  if (!head && !show) return;
  if (show) toggleMember(shownCommands, show.dataset.show);
  else toggleMember(openCards, head.dataset.card);
  drawSets(lastSets);
}

function toggleMember(set, key) {
  set.has(key) ? set.delete(key) : set.add(key);
}

// ---- page

function drawSets(body) {
  lastSets = body;
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
    <div class="ms-intro">${escapeHtml(set.description)}<a href="#ops">→ Ops step</a></div>${setNotesHtml(set)}
    ${zones.map(([chainId, txs]) => zoneHtml(set.id, chainId, txs)).join('')}</details>`;
}

// The keyring note and the files to share are the same shape for every tx of a set, so they are said once here
// rather than on every card. Both come from the first tx that has commands (a tx waiting on a snapshot has none).
function setNotesHtml(set) {
  const first = set.txs.find((tx) => tx.commands.length);
  if (!first) return '';
  const signLabel = (first.commands.find(isSignCommand) || { label: '' }).label;
  const keyring = keyringNote(signLabel);
  const note = keyring ? `<div class="ms-note muted">signing ${escapeHtml(keyring)}</div>` : '';
  const perTx = set.txs.filter((tx) => tx.files.length).length > 1 ? 'per tx, e.g. ' : '';
  const files = first.files.length
    ? `<div class="ms-note muted">share between people (Slack), ${perTx}${first.files.map((file) => `<span class="mono">${escapeHtml(file)}</span>`).join(', ')}</div>`
    : '';
  return note || files ? `<div class="ms-notes">${note}${files}</div>` : '';
}

// The "needs the F5 multisig key in your keyring: ..." clause the server appends to every sign label.
function keyringNote(label) {
  const match = label.match(/needs the .*\)$/);
  return match ? match[0].slice(0, -1) : '';
}

function stripKeyringNote(label) {
  return label.replace(/ \(needs the [^)]*\)$/, '').replace(/; needs the [^)]*\)$/, ')');
}

function isSignCommand(command) {
  return Boolean(WHO_CLASS[command.tag]) && /^(Sign|Backup)/.test(command.label);
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
    ${txs.map((tx, index) => txHtml(tx, `${setId}/${chainId}/${index}`)).join('')}</div>`;
}

// A card is one title row until clicked; open, it lists the commands as rows with a Copy button each.
function txHtml(tx, key) {
  const open = openCards.has(key);
  const state = tx.ready ? pill('ok', 'ready') : pill('warn', 'not ready');
  const reason = tx.reason ? `<span class="muted">${escapeHtml(tx.reason)}</span>` : '';
  const body = open && tx.commands.length ? `<table class="ms-cmds">${commandRows(tx.commands, key).join('')}</table>` : '';
  return `<div class="ms-tx ${open ? 'open' : ''}">
    <div class="ms-tx-head" data-card="${escapeHtml(key)}"><span class="ms-caret">${open ? '▾' : '▸'}</span> <b>${escapeHtml(tx.title)}</b> ${state} ${reason}</div>${body}</div>`;
}

// The three sign commands (Sam, Aidan, Riley as backup) fold into one "sign" row with a Copy button per person;
// every other command keeps a row of its own.
function commandRows(commands, key) {
  const rows = [];
  commands.forEach((command, index) => {
    if (!isSignCommand(command)) {
      rows.push(commandRow(command, `${key}/${index}`));
      return;
    }
    const signers = commands.filter(isSignCommand);
    if (command === signers[0]) rows.push(signRow(signers, `${key}/${index}`));
  });
  return rows;
}

function commandRow(command, rowKey) {
  const copy = `<button type="button" class="ms-copy copy" data-copy="${escapeHtml(command.text)}" title="copy to clipboard">Copy</button>`;
  return `<tr><td>${whoPill(command.tag)}</td><td>${escapeHtml(stripKeyringNote(command.label))}</td>
    <td class="ms-actions">${copy} ${showButton(rowKey)}</td></tr>${commandText([command], rowKey)}`;
}

function signRow(signers, rowKey) {
  const copies = signers.map((command) => {
    const backup = /^Backup/.test(command.label);
    return `<button type="button" class="ms-copy copy ${WHO_CLASS[command.tag]}${backup ? ' backup' : ''}" data-copy="${escapeHtml(command.text)}"
      title="${escapeHtml(stripKeyringNote(command.label))} · copy to clipboard">${escapeHtml(command.tag)}${backup ? ' · backup' : ''}</button>`;
  });
  return `<tr><td><span class="who">any 2 of 3</span></td><td>sign</td>
    <td class="ms-actions"><span class="muted">copy for</span> ${copies.join(' ')} ${showButton(rowKey)}</td></tr>${commandText(signers, rowKey)}`;
}

function showButton(rowKey) {
  const shown = shownCommands.has(rowKey);
  return `<button type="button" class="ms-show" data-show="${escapeHtml(rowKey)}">${shown ? 'hide' : 'show'}</button>`;
}

// The command text under its row, only once "show" was clicked (the Copy buttons carry the text regardless).
function commandText(commands, rowKey) {
  if (!shownCommands.has(rowKey)) return '';
  const blocks = commands.map(
    (command) => `<pre class="ms-command mono copy" data-copy="${escapeHtml(command.text)}" title="click to copy">${escapeHtml(command.text)}</pre>`,
  );
  return `<tr class="ms-text"><td colspan="3">${blocks.join('')}</td></tr>`;
}

function whoPill(tag) {
  return `<span class="who ${WHO_CLASS[tag] || ''}">${escapeHtml(tag)}</span>`;
}
})();
