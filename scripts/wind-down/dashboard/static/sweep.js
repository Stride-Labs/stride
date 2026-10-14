// Sweep tab: the holder sweep's progress. Like Ops and Multisig it polls its own route, /api/sweep, which composes the
// live holder snapshot (refreshed on demand: the button here, or automatically when the snapshot is missing or older
// than ten minutes while this tab is open) with the plan and ledger read from disk. Every integer is a string.
(() => {

const SWEEP_POLL_MS = 10 * 1000;
const AUTO_REFRESH_AFTER_MS = 10 * 60 * 1000;
const STATE_CLASS = { confirmed: 'ok', submitted: 'warn', failed: 'bad', lost: 'bad', pending: 'idle' };

let lastBody = '';
let refreshRequested = false;

registerSelfPollingTab('sweep', startSweep);

function startSweep() {
  const root = document.getElementById('view-sweep');
  root.innerHTML = '<div id="sweepMessage"></div><div id="sweepHead" class="panel"></div><div id="sweepBody"></div>';
  root.addEventListener('click', (event) => {
    if (event.target.id === 'sweepRefresh') requestRefresh();
  });
  pollSweep();
  setInterval(pollSweep, SWEEP_POLL_MS);
}

async function pollSweep() {
  const response = await fetch('/api/sweep').catch(() => null);
  const raw = response ? await response.text() : '';
  if (!response || response.status !== 200) {
    showMessage(response ? `could not read the sweep state: ${errorText(raw)}` : 'server unreachable');
    return;
  }
  showMessage('');
  const body = JSON.parse(raw);
  maybeAutoRefresh(body);
  if (raw === lastBody) return;
  lastBody = raw;
  draw(body);
}

function maybeAutoRefresh(body) {
  if (location.hash.slice(1).split('/')[0] !== 'sweep' || body.refreshing || refreshRequested) return;
  const age = body.fetched_at ? Date.now() - Date.parse(body.fetched_at) : Infinity;
  if (age > AUTO_REFRESH_AFTER_MS) requestRefresh();
}

async function requestRefresh() {
  refreshRequested = true;
  await fetchJson('/api/refresh/sweep', { method: 'POST' });
  setTimeout(() => { refreshRequested = false; }, 60 * 1000);
}

function errorText(raw) {
  try { return JSON.parse(raw).error; } catch (error) { return raw.slice(0, 200); }
}

function showMessage(text) {
  document.getElementById('sweepMessage').innerHTML = text ? `<div class="errors">${escapeHtml(text)}</div>` : '';
}

// ---- drawing

function draw(body) {
  const data = body.data;
  document.getElementById('sweepHead').innerHTML = headLine(body);
  if (!data) {
    document.getElementById('sweepBody').innerHTML = '<div class="placeholder">no sweep state to show</div>';
    return;
  }
  if (!data.run) {
    document.getElementById('sweepBody').innerHTML = totalSection(data) +
      '<div class="placeholder">no plan yet: run `python3 scripts/wind-down/sweep/cli.py plan …`</div>';
    return;
  }
  document.getElementById('sweepBody').innerHTML =
    totalSection(data) + runLine(data) + planSection(data) + byDenomTable(data) +
    `<div class="two"><div>${batchesPanel(data)}</div><div>${ladderPanel(data)}${refundedPanel(data)}</div></div>` +
    exclusionsPanel(data) + deferredPanel(data) + keylessPanel(data);
}

function headLine(body) {
  const age = body.fetched_at ? `live holders read ${formatRelative(body.fetched_at)}` : 'live holders not read yet';
  const refreshing = body.refreshing ? ' · refreshing…' : '';
  const error = body.last_error ? ` · <span class="t-bad">last refresh failed: ${escapeHtml(body.last_error)}</span>` : '';
  return `<div class="ops-bar"><span>${age}${refreshing}${error}</span>
    <button class="refresh" id="sweepRefresh" type="button" ${body.refreshing ? 'disabled' : ''}>Refresh live holders</button></div>`;
}

function runLine(data) {
  const run = data.run;
  const b = run.batches;
  const fees = run.operator_strd === null ? 'operator balance n/a'
    : `operator holds ${formatAmount(run.operator_strd, 6)} STRD, ${formatAmount(run.fee_estimate_ustrd, 6)} STRD needed for the pending batches`;
  const skipping = (run.skipped_denoms || []).length ? ` · skipping ${run.skipped_denoms.join(', ')}` : '';
  return `<div class="callout">Run ${run.run_id}${run.test ? ' (test)' : ''} · floor $${run.floor_usd}${skipping} · planned ${formatRelative(run.created_at)} at height ${run.height}
    · ${b.total} batches: ${b.confirmed} confirmed, ${b.submitted} submitted, ${b.failed} failed, ${b.lost} lost, ${b.pending} pending · ${fees}</div>`;
}

// One summary tile: the address count, with the USD (or "n/a" when the figure is null) and an optional note below.
function summaryTile(cls, label, bucket, note) {
  const value = bucket ? Number(bucket.addresses).toLocaleString() : 'n/a';
  const sub = bucket ? `addresses · $${money(bucket.usd)}` : 'needs a live holder read';
  return `<div class="tile ${cls}"><div class="k">${label}</div><div class="v">${value}</div><div class="s">${sub}${note ? ' · ' + note : ''}</div></div>`;
}

// A stacked bar with a legend; each part is [class, label, value, text]. 
function segmentBar(title, parts) {
  const total = parts.reduce((sum, part) => sum + part[2], 0);
  const share = (value) => (total ? 100 * value / total : 0);
  const legend = parts.map(([cls, label, value, text]) => `<span><i class="${cls}"></i>${label} ${text} (${share(value).toFixed(1)}%)</span>`).join('');
  const bar = parts.map(([cls, , value]) => `<span class="${cls}" style="width:${share(value).toFixed(2)}%"></span>`).join('');
  return `<div class="seg"><div class="seg-title">${title}</div><div class="legend">${legend}</div><div class="progress"><div class="bar">${bar}</div></div></div>`;
}

function addressText(bucket) { return `${Number(bucket.addresses).toLocaleString()} addr`; }
function usdText(bucket) { return `$${money(bucket.usd)}`; }

// TOTAL: every run in the ledger, chain-wide. Remaining is null without a live snapshot.
function totalSection(data) {
  const total = data.summary.total;
  const tiles = `<div class="tiles tiles-3">${summaryTile('sw-swept', 'Swept', total.swept)}${summaryTile('sw-remaining', 'Remaining', total.remaining)}
    ${summaryTile('sw-deferred', 'Deferred (STRD unbonding)', total.deferred)}</div>`;
  const none = { addresses: '0', usd: '0' };
  const bar = (title, key, text) => segmentBar(title, [
    ['sw-swept', 'swept', Number(total.swept[key]), text(total.swept)],
    ['sw-remaining', 'remaining', Number((total.remaining || none)[key]), total.remaining ? text(total.remaining) : 'n/a'],
    ['sw-deferred', 'deferred', Number(total.deferred[key]), text(total.deferred)]]);
  return `<div class="panel"><h2>Total <span class="sub">every run in the ledger, chain-wide; remaining is the live holders not yet swept, excluded, skipped or deferred</span></h2>
    ${tiles}${bar('By USD', 'usd', usdText)}${bar('By addresses', 'addresses', addressText)}</div>`;
}

// CURRENT PLAN: this plan.json alone, with the per-tier table folded under the bar.
function planSection(data) {
  const plan = data.summary.plan;
  if (!plan) return '';
  const tiles = `<div class="tiles tiles-5">${summaryTile('sw-swept', 'Swept', plan.swept)}${summaryTile('sw-remaining', 'Remaining', plan.remaining)}
    ${summaryTile('sw-deferred', 'Deferred', plan.deferred)}${summaryTile('sw-excluded', 'Excluded', plan.excluded)}
    ${summaryTile('sw-refund', 'Refunded', plan.refunded, 'the next plan re-sweeps')}</div>`;
  const bar = segmentBar('By addresses', [
    ['sw-swept', 'swept', Number(plan.swept.addresses), addressText(plan.swept)],
    ['sw-refund', 'refunded', Number(plan.refunded.addresses), addressText(plan.refunded)],
    ['sw-remaining', 'remaining', Number(plan.remaining.addresses), addressText(plan.remaining)]]);
  return `<div class="panel"><h2>Current plan <span class="sub">this plan.json only</span></h2>${tiles}${bar}${tierTable(data.progress)}</div>`;
}

function tierTable(p) {
  if (!p) return '';
  const count = (done, total) => `${Number(done).toLocaleString()} of ${Number(total).toLocaleString()}`;
  const rows = p.tiers.map((t) => `<tr><td>${escapeHtml(t.name)}</td><td class="num">${count(t.batches_confirmed, t.batches)}</td>
    <td class="num">${count(t.addresses_confirmed, t.addresses)}</td></tr>`).join('');
  return `<table class="tier-table"><thead><tr><th>Tier</th><th class="num">Batches confirmed</th><th class="num">Holders confirmed</th></tr></thead><tbody>${rows}</tbody></table>`;
}

// A thin swept-versus-remaining bar by USD for one token row; nothing when both are zero.
function miniBar(d) {
  const swept = Number(d.swept.usd);
  const remaining = Number(d.remaining.usd);
  const total = swept + remaining;
  if (!total) return '';
  return `<div class="bar mini"><span class="sw-swept" style="width:${(100 * swept / total).toFixed(2)}%"></span><span class="sw-remaining" style="width:${(100 * remaining / total).toFixed(2)}%"></span></div>`;
}

function byDenomTable(data) {
  if (!data.by_denom) return '';
  const rows = data.by_denom.map((d) => `<tr>
    <td>${escapeHtml(d.symbol)}</td><td class="muted">${escapeHtml(d.destination)} · ${escapeHtml(d.channel)}</td>
    <td class="num">${d.swept.addresses}</td><td class="num">${d.remaining.addresses}</td>
    <td class="num">${formatAmount(d.swept.amount, Number(d.decimals))}</td><td class="num">${formatAmount(d.remaining.amount, Number(d.decimals))}</td>
    <td class="num">$${money(d.swept.usd)}</td><td class="num">$${money(d.remaining.usd)}</td>
    <td class="num ${d.refunded_addresses !== '0' ? 't-warn' : 'muted'}">${d.refunded_addresses}</td>
    <td class="num">$${money(d.excluded_usd)}</td><td class="num">$${money(d.below_floor_usd)}</td><td class="num">$${money(d.deferred_usd)}</td>
    <td>${miniBar(d)}</td></tr>`).join('');
  return `<div class="panel"><h2>By token <span class="sub">live balances of planned and excluded holders; swept amounts from the ledger's transfer events</span></h2>
    <div class="table-scroll"><table><thead><tr><th>Token</th><th>Destination</th><th class="num">Swept addr</th><th class="num">Remaining addr</th><th class="num">Swept amount</th><th class="num">Remaining amount</th>
    <th class="num">Swept USD</th><th class="num">Remaining USD</th><th class="num">Refunded</th><th class="num">Excluded USD</th><th class="num">Below floor USD</th><th class="num">Deferred USD</th><th>Swept / remaining</th></tr></thead>
    <tbody>${rows}</tbody></table></div></div>`;
}

function batchesPanel(data) {
  let currentRun = null;
  const rows = data.batches.map((b) => {
    const header = b.run_id !== currentRun ? `<tr class="zone"><td colspan="7">Run ${escapeHtml(b.run_id)}</td></tr>` : '';
    currentRun = b.run_id;
    const skips = b.skipped.length ? ` <span class="badge bad" title="${escapeHtml(b.skipped.map((s) => `${s.address}: ${s.reason}`).join('\n'))}">${b.skipped.length} skipped</span>` : '';
    const refunds = b.refunded !== '0' ? ` <span class="badge warn">${b.refunded} refunded</span>` : '';
    return `${header}<tr><td>${escapeHtml(b.batch_id)}</td><td>${b.tier ? escapeHtml(b.tier) : '<span class="muted">earlier run</span>'}</td>
      <td class="num">${b.addresses ?? '<span class="muted">n/a</span>'}</td><td class="num">${b.usd === null ? '<span class="muted">n/a</span>' : '$' + money(b.usd)}</td>
      <td>${pill(STATE_CLASS[b.state] || 'idle', b.state)}${skips}${refunds}</td><td>${b.tx_hash ? addressCell(b.tx_hash, 6, 4) : '<span class="muted">—</span>'}</td>
      <td class="muted">${b.at ? formatRelative(b.at) : '—'}</td></tr>`;
  }).join('');
  return `<div class="panel"><h2>Runs and batches <span class="sub">from state/ledger.jsonl; a batch is one tx</span></h2>
    <table><thead><tr><th>Batch</th><th>Tier</th><th class="num">Addr</th><th class="num">USD</th><th>Status</th><th>Tx</th><th>When</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

function ladderPanel(data) {
  if (!data.ladder) return '';
  const rows = data.ladder.map((r, i) => `<tr><td class="${i === 0 ? 'cur' : ''}">$${r.floor}${i === 0 ? ' (current plan)' : ''}</td>
    <td class="num">${Number(r.holders).toLocaleString()}</td><td class="num">$${money(r.usd)}</td></tr>`).join('');
  return `<div class="panel"><h2>Next floor <span class="sub">live holders not yet swept, by floor; deferred holders counted</span></h2>
    <table class="ladder"><thead><tr><th>Floor</th><th class="num">Holders</th><th class="num">USD</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

function refundedPanel(data) {
  if (!data.refunded.length) return '';
  const rows = data.refunded.map((r) => `<tr><td>${addressCell(r.address)}</td><td>${escapeHtml(r.denom)}</td><td class="num">${r.amount}</td>
    <td>${escapeHtml(r.channel)}</td><td>${escapeHtml(r.batch_id)}</td></tr>`).join('');
  return `<div class="panel"><h2>Refunded <span class="sub">swept, then the transfer timed out and the balance is back; the next plan re-sweeps them</span></h2>
    <table><thead><tr><th>Address</th><th>Denom</th><th class="num">Amount (base units)</th><th>Channel</th><th>Batch</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

function exclusionsPanel(data) {
  const total = data.exclusions.reduce((sum, e) => sum + Number(e.live_usd || 0), 0);
  const rows = data.exclusions.map((e) => `<tr><td>${escapeHtml(e.section)}</td><td>${addressCell(e.address, 14, 8)}</td><td>${escapeHtml(e.label)}</td>
    <td class="muted">${escapeHtml(e.reason)}</td><td class="num">${e.live_usd === null ? 'n/a' : '$' + money(e.live_usd)}</td></tr>`).join('');
  return `<details class="panel"><summary><h2>Excluded <span class="sub">from sweep/exclusions.json · ${data.exclusions.length} addresses · $${money(String(total))} live</span></h2></summary>
    <table><thead><tr><th>Section</th><th>Address</th><th>Label</th><th>Reason</th><th class="num">Live USD</th></tr></thead><tbody>${rows}</tbody></table></details>`;
}

function deferredPanel(data) {
  const d = data.deferred;
  if (!d) return '';
  const rows = d.top.map((r) => `<tr><td>${addressCell(r.address, 14, 8)}</td><td class="num">$${money(r.usd)}</td>
    <td class="num">${money(r.staked_strd)}</td><td class="muted">${r.completes_at ? escapeHtml(r.completes_at.slice(0, 10)) : 'n/a'}</td></tr>`).join('');
  return `<details class="panel"><summary><h2>Deferred <span class="sub">${d.addresses} addresses · $${money(d.usd)} · top ${d.top.length} by plan USD</span></h2></summary>
    <div class="note">Holders with STRD still unbonding, any amount, left out of pass 1 so no address is processed twice; pass 2 sweeps them with the matured STRD.</div>
    <table><thead><tr><th>Address</th><th class="num">USD</th><th class="num">Staked STRD</th><th>Completes</th></tr></thead><tbody>${rows}</tbody></table></details>`;
}

function keylessPanel(data) {
  const k = data.keyless;
  if (!k) return '';
  return `<details class="panel"><summary><h2>Keyless tier <span class="sub">${k.addresses} addresses · $${money(k.usd)} · ${k.swept} swept · never signed; swept last so owners get the most time</span></h2></summary>
    <div class="note">Accounts with no pubkey and sequence 0 at plan time. They sit in the plan's last tier; one that signs a tx before its batch runs moves into the main tier at the next plan.</div></details>`;
}

function money(text) {
  return Number(text).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

})();
