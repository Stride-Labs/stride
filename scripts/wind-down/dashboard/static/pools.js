// Pools tab: per zone, the transmuter pools the vault administers on Osmosis (one canonical plus one per in-scope
// route), what each holds, what it must be funded with, and the pre-funding checks that used to live in
// check_transmuter_pool.py. Until a pool exists its planned row (the denoms, whether the stToken denom is seeded on
// Osmosis as the test wallet's balance, the factors) sits above the live rows. Every integer in the payload is a string;
// arithmetic here is BigInt. A zone's panel is collapsed unless something needs attention (a failing check, a missing
// route, a pool still to create); a panel the user toggled by hand keeps that state across re-renders. A planned row's
// "created" link is `#pools/<contract>`, which scrolls to that pool's live row.
(() => {

const FEE_DECIMALS = 6; // the poolmanager fee denoms (alloyed USDC, uosmo) are six-decimal; the payload has no decimals for it

const userOpen = new Map(); // chain id -> open, once the user toggled the panel by hand
let drawn = false; // the first draw honours a `#pools/<contract>` hash by scrolling to that live row

registerTab('pools', renderPools);

function renderPools(data, root) {
  root.innerHTML = data.zones.map(zonePanel).join('');
  // `toggle` does not bubble, so listen in the capture phase; once is enough since the root element is reused.
  if (!root.dataset.wired) {
    root.addEventListener('toggle', onToggle, true);
    window.addEventListener('hashchange', scrollToHashedPool);
    root.dataset.wired = 'true';
  }
  if (drawn) return;
  drawn = true;
  scrollToHashedPool();
}

function onToggle(event) {
  const target = event.target;
  if (target.dataset.zone) userOpen.set(target.dataset.zone, target.open);
}

// `#pools/<contract>` opens the zone holding that live pool and scrolls to its row (the planned table's "created" link).
function scrollToHashedPool() {
  const [tab, contract] = location.hash.slice(1).split('/');
  if (tab !== 'pools' || !contract) return;
  const row = document.getElementById(`pool-${contract}`);
  if (!row) return;
  row.closest('details').open = true;
  row.scrollIntoView();
}

// ---- zone panel

function zonePanel(zone) {
  if (zone.error) {
    return `<div class="panel"><h2>${escapeHtml(zone.chain_id)} ${pill('bad', 'error')} <span class="sub t-bad">${escapeHtml(zone.error)}</span></h2></div>`;
  }
  const open = userOpen.has(zone.chain_id) ? userOpen.get(zone.chain_id) : needsAttention(zone);
  return `<details class="panel" data-zone="${escapeHtml(zone.chain_id)}" ${open ? 'open' : ''}>
    <summary><h2>${escapeHtml(zone.chain_id)} <span class="muted">· st${escapeHtml(zone.symbol)}</span> ${statePill(zone)}
      <span class="sub">${headerLine(zone)}</span>${feeLine(zone)}</h2></summary>
    ${missingRoutesNote(zone)}${plannedTable(zone)}${poolsTable(zone)}</details>`;
}

// Open by default when the operator has something to fix: a check fails, a route has no pool yet, or a planned pool
// is still to be created.
function needsAttention(zone) {
  return (
    zone.missing_routes.length > 0 ||
    zone.pools.some((pool) => pool.checks.some((check) => check.ok === false)) ||
    plannedPools(zone).some((pool) => !pool.live_contract)
  );
}

// An older payload has no `planned`; the tab then shows the live rows alone.
function plannedPools(zone) {
  return zone.planned || [];
}

// `creation fee 20 allUSDC × 3 to create · vault holds 0`: what the create-pool txs will cost the vault, in red while
// it cannot pay for the pools still to create. Only while the zone has planned pools.
function feeLine(zone) {
  const planned = plannedPools(zone);
  if (!planned.length) return '';
  const toCreate = planned.filter((pool) => !pool.live_contract).length;
  const fee = zone.creation_fee
    ? `${formatAmount(zone.creation_fee.amount, FEE_DECIMALS, 0)} ${escapeHtml(feeSymbol(zone.creation_fee.denom))}`
    : 'n/a';
  const holds = zone.vault_fee_balance === null || zone.vault_fee_balance === undefined ? 'n/a' : formatAmount(zone.vault_fee_balance, FEE_DECIMALS, 0);
  const text = `creation fee ${fee} × ${toCreate} to create · vault holds ${holds}`;
  return `<span class="sub">${zone.creation_fee_short ? `<span class="t-bad">${text}</span>` : text}</span>`;
}

// The readable tail of a factory denom (`factory/<creator>/alloyed/allUSDC` reads allUSDC); a bare denom is itself.
function feeSymbol(denom) {
  return denom.split('/').pop();
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
  const rate = `rate ${escapeHtml(shortRate(zone.stride_rate))}`;
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

// ---- planned table: the pools the zone will get, one row each until the pool exists

// Only the pools still to create: a created one is on the live rows below, not here.
function plannedTable(zone) {
  const planned = plannedPools(zone).filter((pool) => !pool.live_contract);
  if (!planned.length) return '';
  const header = `<tr><th>Planned</th><th>Denom on holder chain</th><th>Denom on Osmosis</th>
    <th title="the test wallet holds the denom on Osmosis (so it exists there, and can fund the test join)">Test wallet</th>
    <th title="the normalization factors the pool is created with: stToken, then native (the rate at that scale)">Factors</th>
    <th>Status</th></tr>`;
  return `<div class="table-scroll pl-planned"><table>${header}${planned.map((pool) => plannedRow(pool, zone)).join('')}</table></div>`;
}

// A planned pool shows everything its creation needs; a route that could not be resolved shows why in place of its
// denoms.
function plannedRow(pool, zone) {
  const denoms = pool.error
    ? `<td colspan="2" class="t-bad pl-error">${escapeHtml(pool.error)}</td>`
    : `<td>${holderDenomCell(pool, zone)}</td><td>${addressCell(pool.denom_on_osmosis, 12, 6)}</td>`;
  const create = `<a href="#multisig/pool-creation/${escapeHtml(zone.chain_id)}" title="the create-pool tx on the Multisig tab">→ create</a>`;
  return `<tr><td>${plannedKindCell(pool)}</td>${denoms}
    <td>${seededMark(pool, zone)}</td>
    <td class="mono">${factorsCell(pool)}</td>
    <td>${pool.blocked ? `<span class="badge bad" title="${escapeHtml(pool.blocked)}">blocked</span>` : pill('idle', 'planned')} ${create}</td></tr>`;
}

// canonical, or the route's holder chain and the Stride channel its stTokens leave over; the alloyed subdenom on hover.
function plannedKindCell(pool) {
  const subdenom = `alloyed ${pool.alloyed_subdenom}`;
  if (pool.kind === 'canonical') return `<b title="${escapeHtml(subdenom)}">canonical</b>`;
  const title = `Stride ${pool.stride_channel} ↔ ${pool.holder_chain_id} ${pool.counterparty_channel || 'n/a'} · ${subdenom}`;
  const holder = pool.holder_name || pool.holder_chain_id || 'unknown chain';
  return `<b>route</b> <span class="muted" title="${escapeHtml(title)}">${escapeHtml(holder)} · ${escapeHtml(pool.stride_channel)}</span>`;
}

// The canonical pool's stToken never leaves Stride before Osmosis, so its holder-chain denom is the stToken itself.
function holderDenomCell(pool, zone) {
  if (pool.kind === 'canonical') return `<span class="mono muted" title="the stToken itself, on Stride">${escapeHtml(zone.st_denom)}</span>`;
  return addressCell(pool.denom_on_holder, 12, 6);
}

function seededMark(pool, zone) {
  if (pool.seeded === true) return `<span class="pl-check ok" title="the test wallet's balance">✓ ${amount(pool.test_wallet_balance, zone.decimals)}</span>`;
  if (pool.seeded === false) return `<span class="pl-check bad" title="the test wallet does not hold this denom yet: send it some">✗</span>`;
  return `<span class="pl-check na" title="Osmosis's supply of the denom could not be read">n/a</span>`;
}

function factorsCell(pool) {
  return `<span class="muted">st</span> ${escapeHtml(factorLabel(pool.st_factor))} · <span class="muted">native</span> ${escapeHtml(pool.native_factor)}`;
}

// A one followed by zeros reads as its power of ten (1e18, 1e6); anything else is printed whole.
function factorLabel(factor) {
  const match = /^1(0*)$/.exec(String(factor));
  return match ? `1e${match[1].length}` : String(factor);
}


// The live pool's id (its contract when the listing had no id), linking to its row below.
function livePoolLink(contract, zone) {
  const live = zone.pools.find((pool) => pool.contract === contract);
  const label = live && live.pool_id !== null ? live.pool_id : contract;
  return `<a href="#pools/${escapeHtml(contract)}" title="${escapeHtml(contract)}">${escapeHtml(label)}</a>`;
}

// ---- pools table

function poolsTable(zone) {
  if (!zone.pools.length) return `<div class="note">no pool administered by the vault holds st${escapeHtml(zone.symbol)} yet</div>`;
  const header = `<tr><th>Pool</th><th>Id</th><th class="num">Native</th><th class="num">stToken</th>
    <th class="num" title="alloyed shares held by the vault / by the test wallet">Shares vault / test wallet</th>
    <th class="num" title="what the pool must be funded with: route = escrow × the pool's rate, canonical = the remainder">Allocation</th>
    <th title="the native token marked corrupted (one-way pool)">Corrupted</th><th class="num" title="the pool's rate and its gap below Stride's frozen rate">Rate</th></tr>`;
  return `<div class="table-scroll"><table>${header}${zone.pools.map((pool) => poolRows(pool, zone)).join('')}</table></div>`;
}

// The pool's figures, then its checks on a full-width row beneath (fifteen marks do not fit a column). The row's id is
// what a planned row's "created" link scrolls to.
// A pool whose every check passes folds them into one green check beside its kind (the list on hover); a pool with
// a failing or unknown check keeps the full list under its row.
function poolRows(pool, zone) {
  const allOk = pool.checks.every((check) => check.ok === true);
  const cssClass = pool.checks.some((check) => check.ok === false) ? 'pl-fail' : '';
  const checksRow = allOk ? '' : `<tr class="${cssClass}"><td class="pl-checks" colspan="8">${pool.checks.map(checkMark).join('')}</td></tr>`;
  return `<tr class="${cssClass}" id="pool-${escapeHtml(pool.contract)}">
    <td>${allOk ? allChecksMark(pool) : ''}${kindCell(pool)}</td>
    <td>${idCell(pool)}</td>
    <td class="num">${amount(pool.native_balance, zone.decimals)}</td>
    <td class="num">${amount(pool.st_balance, zone.decimals)}</td>
    <td class="num">${amount(pool.vault_shares, zone.decimals)} / ${amount(pool.test_wallet_shares, zone.decimals)}</td>
    <td class="num">${amount(pool.allocation, zone.decimals)} ${fundedMark(pool)}</td>
    <td>${corruptedCell(pool, zone)}</td>
    <td class="num">${rateCell(pool)}</td></tr>${checksRow}`;
}

function allChecksMark(pool) {
  const title = `${pool.checks.length} checks pass:\n${pool.checks.map((check) => `✓ ${check.name}`).join('\n')}`;
  return `<span class="pl-check ok pl-all" title="${escapeHtml(title)}">✓</span> `;
}

// canonical, or the route's holder chain and the Stride channel its stTokens left over (the escrow the pool pays out).
function kindCell(pool) {
  if (pool.kind === 'canonical') return `<b>canonical</b>`;
  if (pool.kind === 'route' && pool.route) {
    const title = `Stride ${pool.route.stride_channel} ↔ ${pool.route.chain_id} ${pool.route.counterparty_channel} · escrow ${pool.escrow === null ? 'n/a' : pool.escrow}`;
    return `<b>route</b> <span class="muted" title="${escapeHtml(title)}">${escapeHtml(pool.route.chain_id)} · ${escapeHtml(pool.route.stride_channel)}</span>`;
  }
  if (pool.kind === 'route') return `<b>route</b> <span class="muted">lookup n/a</span>`;
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
