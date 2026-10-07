// Pools tab: per zone, the transmuter pools the vault administers on Osmosis (one canonical plus one per in-scope
// route), what each holds, what it must be funded with, and the pre-funding checks that used to live in
// check_transmuter_pool.py. Every integer in the payload is a string; arithmetic here is BigInt. A zone's panel is
// collapsed unless something needs attention (a failing check or a missing route); a panel the user toggled by hand
// keeps that state across re-renders.
(() => {

const userOpen = new Map(); // chain id -> open, once the user toggled the panel by hand

registerTab('pools', renderPools);

function renderPools(data, root) {
  root.innerHTML = data.zones.map(zonePanel).join('');
  // `toggle` does not bubble, so listen in the capture phase; once is enough since the root element is reused.
  if (!root.dataset.wired) {
    root.addEventListener('toggle', onToggle, true);
    root.dataset.wired = 'true';
  }
}

function onToggle(event) {
  const target = event.target;
  if (target.dataset.zone) userOpen.set(target.dataset.zone, target.open);
}

// ---- zone panel

function zonePanel(zone) {
  if (zone.error) {
    return `<div class="panel"><h2>${escapeHtml(zone.chain_id)} ${pill('bad', 'error')} <span class="sub t-bad">${escapeHtml(zone.error)}</span></h2></div>`;
  }
  const open = userOpen.has(zone.chain_id) ? userOpen.get(zone.chain_id) : needsAttention(zone);
  return `<details class="panel" data-zone="${escapeHtml(zone.chain_id)}" ${open ? 'open' : ''}>
    <summary><h2>${escapeHtml(zone.chain_id)} <span class="muted">· st${escapeHtml(zone.symbol)}</span> ${statePill(zone)}
      <span class="sub">${headerLine(zone)}</span></h2></summary>
    ${missingRoutesNote(zone)}${poolsTable(zone)}</details>`;
}

// Open by default when the operator has something to fix: a check fails or a route has no pool yet.
function needsAttention(zone) {
  return zone.missing_routes.length > 0 || zone.pools.some((pool) => pool.checks.some((check) => check.ok === false));
}

// funded (every pool funded exactly and marked) > ready (pools in place, checks green) > not ready; n/a when unknown.
function statePill(zone) {
  if (zone.pools_funded === true) return pill('ok', 'funded');
  if (zone.pools_ready === true) return pill('ok', 'ready');
  if (zone.pools_ready === null) return pill('idle', 'n/a');
  return pill('bad', 'not ready');
}

function headerLine(zone) {
  const tokens = (value) => tokenAmount(value, zone);
  const onOsmosis = `${tokens(zone.native_on_osmosis)} on Osmosis (vault ${tokens(zone.vault_native)} + pools ${tokens(zone.pools_native)})`;
  const rate = zone.stride_rate === null ? 'rate n/a' : `rate ${escapeHtml(shortRate(zone.stride_rate))}`;
  return `needs ${tokens(zone.needed)} · ${onOsmosis} · ${coveragePill(zone.coverage)} · ${rate} · ${zone.pools.length} pool${zone.pools.length === 1 ? '' : 's'}`;
}

// Covered once the ratio reaches one; idle below (the Funds tab says where the rest still is); n/a when unknown.
function coveragePill(coverage) {
  if (coverage === null) return pill('idle', 'coverage n/a');
  const ratio = Number(coverage);
  return pill(ratio >= 1 ? 'ok' : 'idle', `${(ratio * 100).toFixed(1)}% covered`);
}

function missingRoutesNote(zone) {
  if (!zone.missing_routes.length) return '';
  const channels = zone.missing_routes.map((channel) => `<span class="mono">${escapeHtml(channel)}</span>`).join(', ');
  return `<div class="note t-warn">missing route pools: no code-996 pool administered by the vault holds st${escapeHtml(zone.symbol)} via Stride ${channels}</div>`;
}

// ---- pools table

function poolsTable(zone) {
  if (!zone.pools.length) return `<div class="note">no pool administered by the vault holds st${escapeHtml(zone.symbol)} yet</div>`;
  const header = `<tr><th>Pool</th><th>Id</th><th class="num">Native</th><th class="num">stToken</th>
    <th class="num" title="alloyed shares held by the vault / by anyone else">Shares vault / outside</th>
    <th class="num" title="what the pool must be funded with: route = escrow × the pool's rate, canonical = the remainder">Allocation</th>
    <th title="the native token marked corrupted (one-way pool)">Corrupted</th><th class="num" title="the pool's rate and its gap below Stride's frozen rate">Rate</th></tr>`;
  return `<div class="table-scroll"><table>${header}${zone.pools.map((pool) => poolRows(pool, zone)).join('')}</table></div>`;
}

// The pool's figures, then its checks on a full-width row beneath (fifteen marks do not fit a column).
function poolRows(pool, zone) {
  const cssClass = pool.checks.some((check) => check.ok === false) ? 'pl-fail' : '';
  return `<tr class="${cssClass}">
    <td>${kindCell(pool)}</td>
    <td>${idCell(pool)}</td>
    <td class="num">${amount(pool.native_balance, zone.decimals)}</td>
    <td class="num">${amount(pool.st_balance, zone.decimals)}</td>
    <td class="num">${amount(pool.vault_shares, zone.decimals)} / ${amount(pool.outside_shares, zone.decimals)}</td>
    <td class="num">${amount(pool.allocation, zone.decimals)} ${fundedMark(pool)}</td>
    <td>${corruptedCell(pool, zone)}</td>
    <td class="num">${rateCell(pool)}</td></tr>
    <tr class="${cssClass}"><td class="pl-checks" colspan="8">${pool.checks.map(checkMark).join('')}</td></tr>`;
}

// canonical, or the route's holder chain and the Stride channel its stTokens left over (the escrow the pool pays out).
function kindCell(pool) {
  if (pool.kind === 'canonical') return `<b>canonical</b>`;
  if (pool.kind === 'route' && pool.route) {
    const title = `Stride ${pool.route.stride_channel} ↔ ${pool.route.chain_id} ${pool.route.counterparty_channel} · escrow ${pool.escrow === null ? 'n/a' : pool.escrow}`;
    return `<b>route</b> <span class="muted" title="${escapeHtml(title)}">${escapeHtml(pool.route.chain_id)} · ${escapeHtml(pool.route.stride_channel)}</span>`;
  }
  return `${pill('bad', 'unrecognised')} <span class="muted mono" title="${escapeHtml(pool.st_trace)}">${escapeHtml(pool.st_trace)}</span>`;
}

function idCell(pool) {
  const id = pool.pool_id === null ? '<span class="muted">–</span>' : `<b>${escapeHtml(pool.pool_id)}</b>`;
  return `${id} ${addressCell(pool.contract, 10, 6)}`;
}

// The cumulative funding audit: vault shares equal the allocation once the pool has been joined with exactly it.
// Below it is the normal state between the test join and the rest; above it means too much went in.
function fundedMark(pool) {
  if (pool.funded_exactly === true) return pill('ok', 'funded exactly');
  if (pool.funded_exactly === null) return pill('idle', 'n/a');
  const shares = BigInt(pool.vault_shares || 0);
  if (shares === 0n) return pill('idle', 'unfunded');
  return shares < BigInt(pool.allocation) ? pill('warn', 'partly funded') : pill('bad', 'over-funded');
}

function corruptedCell(pool, zone) {
  const mark = pool.native_marked ? pill('ok', 'marked') : pill('idle', 'not marked');
  const others = pool.corrupted.filter((denom) => denom !== zone.osmosis_denom);
  if (!others.length) return mark;
  return `${mark} <span class="t-bad" title="${escapeHtml(others.join(', '))}">+${others.length} other denom${others.length === 1 ? '' : 's'}</span>`;
}

function rateCell(pool) {
  if (pool.rate === null) return '<span class="muted">n/a</span>';
  const gap = pool.rate_gap_pct === null ? '' : ` <span class="muted">(gap ${escapeHtml(pool.rate_gap_pct)}%)</span>`;
  return `${escapeHtml(shortRate(pool.rate))}${gap}`;
}

// ✓ / ✗ / n/a with the check's detail on hover; a failed check is what turns the row red.
function checkMark(check) {
  const cssClass = check.ok === true ? 'ok' : check.ok === false ? 'bad' : 'na';
  const mark = check.ok === true ? '✓' : check.ok === false ? '✗' : 'n/a';
  return `<span class="pl-check ${cssClass}" title="${escapeHtml(check.detail || check.name)}">${mark} ${escapeHtml(check.name)}</span>`;
}

// ---- formatting

function amount(value, decimals, fractionDigits = 2) {
  if (value === null || value === undefined) return '<span class="muted">n/a</span>';
  return formatAmount(value, Number(decimals), fractionDigits);
}

function tokenAmount(value, zone) {
  if (value === null || value === undefined) return 'n/a';
  return `${formatAmount(value, Number(zone.decimals), 0)} ${escapeHtml(zone.symbol)}`;
}

function shortRate(rate) {
  return Number(rate).toFixed(4);
}
})();
