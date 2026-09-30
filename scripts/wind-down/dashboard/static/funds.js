// Funds flow tab: the stage table per zone, the flow diagram of the selected zone, its ICA transfers to the
// vault, and its accounts. Every integer in the payload is a string; arithmetic here is BigInt.
(() => {

const STAGES = [
  { key: 'staked', label: 'Staked', color: '--s-staked' },
  { key: 'unbonding', label: 'Unbonding', color: '--s-unbond' },
  { key: 'liquid', label: 'Liquid in ICAs', color: '--s-ica' },
  { key: 'in_flight', label: 'In flight (IBC)', color: '--s-flight' },
  { key: 'vault', label: 'Osmosis vault', color: '--s-vault' },
  { key: 'pools', label: 'In pools', color: '--s-pool' },
];
const STAGE_TABLE_COLUMNS = 12;
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

let selectedZone = null; // chain id of the zone whose diagram and accounts are shown; survives re-renders
let latest = null; // the last payload, so a click re-renders without a refetch

registerTab('funds', renderFunds);

function renderFunds(data, root) {
  latest = data;
  const selectable = data.zones.filter((zone) => !zone.error);
  if (!selectable.some((zone) => zone.chain_id === selectedZone)) {
    selectedZone = selectable.length ? selectable[0].chain_id : null;
  }

  root.innerHTML = stagePanel(data.zones) + `<div id="fundsSelection">${selectionPanels(data)}</div>`;
  root.querySelectorAll('tr[data-zone]').forEach((row) => {
    row.onclick = () => selectZone(root, row.dataset.zone);
  });
}

function selectZone(root, chainId) {
  selectedZone = chainId;
  root.querySelectorAll('tr[data-zone]').forEach((row) => row.classList.toggle('sel', row.dataset.zone === chainId));
  root.querySelector('#fundsSelection').innerHTML = selectionPanels(latest);
}

function selectionPanels(data) {
  const zone = data.zones.find((candidate) => candidate.chain_id === selectedZone);
  if (!zone) return '<div class="placeholder">no zone to show</div>';
  return diagramPanel(zone) + transfersPanel(zone) + accountsPanel(zone, data.operators);
}

// ---- stage table

function stagePanel(zones) {
  const legend = STAGES.map((stage) => `<span><i style="background:var(${stage.color})"></i>${stage.label}</span>`).join('');
  const header = `<tr><th>Zone</th><th>Progress</th><th class="num">Staked</th><th class="num">Unbonding</th><th title="earliest unbonding completion">Next maturity</th>
    <th class="num" title="delegation + withdrawal + fee ICA balances">Liquid</th><th class="num">In flight</th><th class="num">Vault</th><th class="num">In pools</th>
    <th class="num" title="owed to open user claims, not part of the bar">Redemption</th><th class="num" title="stToken supply × redemption rate">Needed</th><th>Coverage</th></tr>`;
  return `<div class="panel"><h2>Where the backing is, per zone <span class="sub">native units · click a row for its diagram and accounts</span></h2>
    <div class="legend">${legend}</div>
    <div class="table-scroll"><table>${header}${zones.map(stageRow).join('')}</table></div></div>`;
}

function stageRow(zone) {
  if (zone.error) {
    return `<tr><td>${escapeHtml(zone.chain_id)}</td><td colspan="${STAGE_TABLE_COLUMNS - 2}" class="t-bad">${escapeHtml(zone.error)}</td><td>${pill('bad', 'error')}</td></tr>`;
  }
  const stages = zone.stages;
  const cell = (value) => `<td class="num">${amount(value, zone.decimals, 0)}</td>`;
  return `<tr data-zone="${escapeHtml(zone.chain_id)}" class="${zone.chain_id === selectedZone ? 'sel' : ''}">
    <td><b>${escapeHtml(zone.chain_id)}</b> <span class="muted">· ${escapeHtml(zone.symbol)}</span></td>
    <td>${stageBar(stages)}</td>
    ${cell(stages.staked)}${cell(stages.unbonding)}
    <td class="muted">${maturityCell(zone)}</td>
    ${cell(stages.liquid)}${cell(stages.in_flight)}${cell(stages.vault)}${cell(stages.pools)}
    ${cell(zone.redemption_ica_balance)}${cell(zone.needed)}
    <td>${coveragePill(zone)}</td></tr>`;
}

function stageBar(stages) {
  const values = STAGES.map((stage) => BigInt(stages[stage.key] || 0));
  const total = values.reduce((sum, value) => sum + value, 0n);
  if (total === 0n) return '<div class="bar"></div>';

  const segments = STAGES.map((stage, index) => {
    const percent = Number((values[index] * 10000n) / total) / 100;
    return `<span style="width:${percent}%;background:var(${stage.color})" title="${stage.label}: ${percent}%"></span>`;
  });
  return `<div class="bar">${segments.join('')}</div>`;
}

function maturityCell(zone) {
  if (!zone.unbonding_earliest) return '–';
  const remaining = (Date.parse(zone.unbonding_earliest) - Date.now()) / 1000;
  return `${shortDate(zone.unbonding_earliest)} · ${remaining > 0 ? formatDuration(remaining) : 'matured'}`;
}

// Covered once the ratio reaches one; bad when nothing is left upstream and it has not; idle while funds still move.
function coveragePill(zone) {
  if (zone.coverage === null) return pill('idle', 'n/a');
  const ratio = Number(zone.coverage);
  const text = `${(ratio * 100).toFixed(1)}%`;
  if (ratio >= 1) return pill('ok', text);

  const stages = zone.stages;
  const upstream = BigInt(stages.staked) + BigInt(stages.unbonding) + BigInt(stages.liquid) + BigInt(stages.in_flight || 0);
  return pill(upstream === 0n ? 'bad' : 'idle', upstream === 0n ? `${text} · short` : text);
}

// ---- diagram

function diagramPanel(zone) {
  const sub = `rate ${escapeHtml(shortRate(zone.redemption_rate))} · st${escapeHtml(zone.symbol)} supply ${amount(zone.st_supply, zone.decimals, 0)}`
    + ` · needs ${amount(zone.needed, zone.decimals, 0)} ${escapeHtml(zone.symbol)} on Osmosis`;
  return `<div class="panel"><h2>${escapeHtml(zone.chain_id)} · ${escapeHtml(zone.symbol)} <span class="sub">${sub}</span></h2>${diagram(zone)}</div>`;
}

function diagram(zone) {
  const height = zone.staketia ? 440 : 380;
  const parts = [
    `<defs><marker id="arr" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0,0 L10,5 L0,10 z" fill="var(--muted)"/></marker></defs>`,
    `<rect x="6" y="6" width="640" height="296" rx="10" fill="var(--lane-host)"/>${text(18, 24, 'lbl', zone.chain_id.toUpperCase())}`,
    `<rect x="846" y="6" width="268" height="${height - 12}" rx="10" fill="var(--lane-osmo)"/>${text(858, 24, 'lbl', 'OSMOSIS')}`,
    hostLane(zone),
    osmosisLane(zone),
    zone.staketia ? staketiaLane(zone) : '',
  ];
  return `<svg viewBox="0 0 1120 ${height}" width="100%" role="img" aria-label="Flow of ${escapeHtml(zone.symbol)} from staked to the Osmosis pools">${parts.join('')}</svg>`;
}

function hostLane(zone) {
  const stages = zone.stages;
  const symbol = zone.symbol;
  const tokens = (value) => tokenAmount(value, zone);
  const maturity = zone.unbonding_earliest
    ? `next matures ${shortDate(zone.unbonding_earliest)} · last ${shortDate(zone.unbonding_latest)}`
    : 'nothing unbonding';
  const icas = [
    ['DELEGATION', 'Delegation ICA (liquid)'], ['WITHDRAWAL', 'Withdrawal ICA'], ['FEE', 'Fee ICA'], ['REDEMPTION', 'Redemption ICA'],
  ];
  const records = zone.open_redemption_records === null ? 'n/a' : zone.open_redemption_records;
  const inFlight = zone.transfers === null
    ? 'in flight: n/a (tx index)'
    : `in flight: ${tokens(stages.in_flight)} (${zone.transfers.filter((transfer) => transfer.status === 'in flight').length} packet(s))`;
  const leg = zone.host_channel ? `${zone.host_channel} → ${zone.osmosis_channel || '?'} (osmosis)` : 'bank MsgSend on Osmosis';
  return [
    node(20, 50, 180, 62, `Staked · ${zone.validators} validators`, tokens(stages.staked), 'delegation ICA' + (zone.staketia ? ' + multisig' : '')),
    edge('M200,81 L236,81'), text(204, 44, 'tx', 'MsgUndelegateFromValidators'),
    node(238, 50, 180, 62, `Unbonding · ${zone.unbonding_entries} entries`, tokens(stages.unbonding), maturity),
    edge('M418,81 L446,81'),
    `<rect x="448" y="34" width="190" height="258" rx="9" fill="none" stroke="var(--line)" stroke-dasharray="4 3"/>`,
    ...icas.map(([ica, label], index) => node(456, 50 + 58 * index, 174, 50, label, tokens(zone.ica_liquid[ica]))),
    edge('M630,249 L700,249'), text(660, 268, 'lbl', `user claims (${records} open records)`),
    edge('M638,81 L878,81'), text(662, 69, 'tx', 'MsgTransferFromIca'), text(662, 99, 'lbl', leg), text(662, 115, 'lbl', inFlight),
  ].join('');
}

function osmosisLane(zone) {
  const tokens = (value) => tokenAmount(value, zone);
  const pools = zone.pools;
  const canonical = pools ? pools.find((pool) => pool.kind === 'canonical') : null;
  const routes = pools ? pools.filter((pool) => pool.kind === 'route') : [];
  const swapped = (list) => `st${zone.symbol} swapped in: ${tokens(list.length ? sumOf(list.map((pool) => pool.st_amount)) : null)}`;
  const routeNative = routes.length ? sumOf(routes.map((pool) => pool.native)) : (pools ? '0' : null);
  return [
    node(880, 50, 200, 62, 'Osmosis vault', tokens(zone.stages.vault), shortAddress(zone.vault_address || '')),
    edge('M980,112 L980,170'), text(988, 146, 'tx', 'join_pool'),
    node(880, 172, 200, 62, `st${zone.symbol} canonical pool`, canonical ? tokens(canonical.native) : (pools ? '–' : 'n/a'),
      canonical ? swapped([canonical]) : (pools ? 'not found' : 'Osmosis unreachable')),
    node(880, 244, 200, 62, `Route pools (${routes.length})`, tokens(routeNative), routes.length ? swapped(routes) : 'none'),
    text(880, 330, 'lbl', 'Coverage'),
    text(880, 350, 'amt', `${tokens(zone.covered)} / ${tokens(zone.needed)}`),
    text(880, 366, 'lbl', coverageNote(zone)),
  ].join('');
}

function coverageNote(zone) {
  const stages = zone.stages;
  const values = STAGES.map((stage) => stages[stage.key]);
  if (zone.coverage === null || values.some((value) => value === null)) return 'coverage n/a';

  const tracked = sumOf(values);
  const difference = BigInt(tracked) - BigInt(zone.needed);
  const sign = difference < 0n ? '−' : '+';
  const magnitude = (difference < 0n ? -difference : difference).toString();
  return `tracked ${amount(tracked, zone.decimals, 0)} → ${sign}${amount(magnitude, zone.decimals, 0)} ${difference < 0n ? 'short' : 'surplus'}`;
}

function staketiaLane(zone) {
  const staketia = zone.staketia;
  const tokens = (value) => tokenAmount(value, zone);
  const parked = sumOf([staketia.unbonding, staketia.liquid]);
  const records = staketia.unbonding_records === null ? 'n/a' : staketia.unbonding_records;
  return [
    `<rect x="6" y="312" width="640" height="122" rx="10" fill="var(--lane-stride)"/>`,
    text(18, 330, 'lbl', 'STAKETIA (multisig on Celestia → claim address on Stride)'),
    node(20, 350, 180, 62, 'Multisig staked', tokens(staketia.staked), shortAddress(staketia.multisig_address)),
    edge('M200,381 L236,381'),
    node(238, 350, 180, 62, 'Multisig unbonding + liquid', tokens(parked), `${records} records · liquid ${tokens(staketia.liquid)}`),
    edge('M418,381 L454,381'),
    node(456, 350, 174, 62, 'Claim address (Stride)', tokens(staketia.claim_balance), `${shortAddress(staketia.claim_address)} · ibc/${zone.symbol}`),
    edge('M543,350 L543,294'), text(552, 330, 'tx', 'MsgTransferStaketiaClaimBalance'),
  ].join('');
}

function node(x, y, width, height, label, value, sub) {
  return `<rect class="node" x="${x}" y="${y}" width="${width}" height="${height}" rx="7"/>`
    + text(x + 10, y + 18, 'lbl', label) + text(x + 10, y + 38, 'amt', value) + (sub ? text(x + 10, y + 54, 'lbl', sub) : '');
}

function edge(path) {
  return `<path class="edge" d="${path}"/>`;
}

function text(x, y, cssClass, content) {
  return `<text x="${x}" y="${y}" class="${cssClass}">${escapeHtml(content)}</text>`;
}

// ---- transfers and accounts

function transfersPanel(zone) {
  const title = `<h2>ICA transfers to the Osmosis vault <span class="sub">${zone.host_channel ? 'from the host tx index · in flight while the host still holds the packet commitment' : 'osmosis-1 moves by bank send; nothing is in flight'}</span></h2>`;
  if (zone.transfers === null) return `<div class="panel">${title}<div class="note t-warn">n/a: the host's tx index could not be searched</div></div>`;
  if (!zone.transfers.length) return `<div class="panel">${title}<div class="note">none yet</div></div>`;

  const header = '<tr><th>Time</th><th>ICA</th><th class="num">Amount</th><th>Denom</th><th class="num">Sequence</th><th>Status</th></tr>';
  const rows = zone.transfers.map((transfer) => `<tr>
    <td class="muted" title="height ${escapeHtml(transfer.height)}">${escapeHtml(shortDateTime(transfer.time))} · ${formatRelative(transfer.time)}</td>
    <td>${escapeHtml(transfer.ica)}</td>
    <td class="num">${transfer.denom === zone.host_denom ? amount(transfer.amount, zone.decimals) : escapeHtml(transfer.amount)}</td>
    <td class="mono">${shortDenom(transfer.denom)}</td>
    <td class="num mono">${escapeHtml(transfer.sequence)}</td>
    <td>${pill(transfer.status === 'in flight' ? 'warn' : 'ok', transfer.status)}</td></tr>`);
  return `<div class="panel">${title}<div class="table-scroll"><table>${header}${rows.join('')}</table></div></div>`;
}

function accountsPanel(zone, operators) {
  const header = '<tr><th>Account</th><th>Chain</th><th>Address</th><th class="num">Liquid</th><th class="num">Staked</th><th class="num">Unbonding</th><th>Note</th></tr>';
  const rows = zone.accounts.concat(operators).map(accountRow);
  return `<div class="panel"><h2>Accounts <span class="sub">every address the diagram reads · ${escapeHtml(zone.chain_id)} + operators · click an address to copy</span></h2>
    <div class="table-scroll"><table>${header}${rows.join('')}</table></div></div>`;
}

function accountRow(account) {
  const cell = (value) => `<td class="num">${value === null ? '<span class="muted">–</span>' : amount(value, account.decimals) + ' ' + escapeHtml(account.symbol)}</td>`;
  const others = account.other_balances.map((balance) => `also ${escapeHtml(balance.amount)} ${shortDenom(balance.denom)}`);
  const note = [account.note, ...others].filter(Boolean).join(' · ');
  return `<tr><td>${escapeHtml(account.name)}</td><td>${escapeHtml(account.chain)}</td><td>${addressCell(account.address, 14, 8)}</td>
    ${cell(account.liquid)}${cell(account.staked)}${cell(account.unbonding)}<td class="muted">${note}</td></tr>`;
}

// ---- formatting

function amount(value, decimals, fractionDigits = 2) {
  if (value === null || value === undefined) return '<span class="muted">n/a</span>';
  return formatAmount(value, Number(decimals), fractionDigits);
}

function tokenAmount(value, zone) {
  if (value === null || value === undefined) return 'n/a';
  return `${formatAmount(value, Number(zone.decimals), 0)} ${zone.symbol}`;
}

function sumOf(values) {
  return values.reduce((total, value) => total + BigInt(value), 0n).toString();
}

function shortRate(rate) {
  return Number(rate).toFixed(4);
}

function shortDate(iso) {
  const date = new Date(iso);
  return `${MONTHS[date.getUTCMonth()]} ${date.getUTCDate()}`;
}

function shortDateTime(iso) {
  const date = new Date(iso);
  const pad = (number) => String(number).padStart(2, '0');
  return `${shortDate(iso)} ${pad(date.getUTCHours())}:${pad(date.getUTCMinutes())} UTC`;
}

function shortAddress(address) {
  if (!address) return '';
  return `${address.slice(0, 10)}…${address.slice(-6)}`;
}

function shortDenom(denom) {
  const shown = denom.length > 20 ? `${denom.slice(0, 12)}…${denom.slice(-4)}` : denom;
  return `<span class="mono" title="${escapeHtml(denom)}">${escapeHtml(shown)}</span>`;
}
})();
