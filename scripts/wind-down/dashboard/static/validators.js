// Validators tab: all zones stacked (or one, via the chips), diff filters, one row per validator (recorded vs host).
(() => {

const MULTISIG = 'multisig';
const SEVERITY_CLASS = { neutral: '', amber: 't-warn', red: 't-bad' };
const SEVERITY_PILL = { amber: 'warn', red: 'bad' };

const ALL_ZONES = 'all';

// View state survives the periodic re-render; chips and filters change it client-side, no refetch.
let selectedChainId = ALL_ZONES;
const filters = { delegation: true, rate: true }; // a row shows when it matches any checked filter
let lastData = null;
let lastRoot = null;

registerTab('validators', renderValidators);

function renderValidators(data, root) {
  lastData = data;
  lastRoot = root;

  const entries = data.zones;
  if (selectedChainId !== ALL_ZONES && !entries.some((entry) => entry.chain_id === selectedChainId)) selectedChainId = ALL_ZONES;
  const shown = selectedChainId === ALL_ZONES ? entries : entries.filter((entry) => entry.chain_id === selectedChainId);

  root.innerHTML = [
    `<div class="chips">${allChip()}${entries.map(chip).join('')}</div>`,
    filterBar(),
    ...shown.map((entry) => entry.error ? `<div class="errors">${escapeHtml(entry.chain_id)}: ${escapeHtml(entry.error)}</div>` : entryBody(entry)),
  ].join('');

  root.querySelectorAll('.chip').forEach((element) => {
    element.onclick = () => {
      selectedChainId = element.dataset.chain;
      renderValidators(lastData, lastRoot);
    };
  });
  root.querySelectorAll('input[data-filter]').forEach((element) => {
    element.onchange = () => {
      filters[element.dataset.filter] = element.checked;
      renderValidators(lastData, lastRoot);
    };
  });
}

function entryBody(entry) {
  return entry.kind === MULTISIG ? multisigBody(entry) : zoneBody(entry);
}

// ---- chips and filters

function allChip() {
  const on = selectedChainId === ALL_ZONES ? 'on' : '';
  return `<span class="chip ${on}" data-chain="${ALL_ZONES}">all zones</span>`;
}

function chip(entry) {
  const on = entry.chain_id === selectedChainId ? 'on' : '';
  return `<span class="chip ${on}" data-chain="${escapeHtml(entry.chain_id)}">${escapeHtml(entry.chain_id)} ${chipBadge(entry)}</span>`;
}

function filterBar() {
  const box = (key, label) => `<label><input type="checkbox" data-filter="${key}" ${filters[key] ? 'checked' : ''}> ${label}</label>`;
  return `<div class="filters"><span class="muted">show validators with</span>
    ${box('delegation', 'delegation diff ≠ 0')} ${box('rate', 'rate diff ≠ 0')}
    <span class="muted">(either; untick both for every validator)</span></div>`;
}

function filterActive() {
  return filters.delegation || filters.rate;
}

function rowMatches(row) {
  if (!filterActive()) return true;
  const delegationDiff = BigInt(row.diff) !== 0n;
  const rateDiff = row.rate_difference !== null && Number(row.rate_difference) !== 0;
  return (filters.delegation && delegationDiff) || (filters.rate && rateDiff);
}

function chipBadge(entry) {
  if (entry.error) return pill('bad', 'error');
  if (entry.kind === MULTISIG) return entry.severity === 'neutral' ? pill('ok', 'ok') : pill(SEVERITY_PILL[entry.severity], 'over');

  const over = entry.totals.over_count;
  if (over === 0) return pill('ok', 'ok');
  return pill(entry.totals.red_count > 0 ? 'bad' : 'warn', `${over} over`);
}

// ---- zone

function zoneBody(zone) {
  const rows = zone.validators.filter(rowMatches);
  const shownNote = rows.length === zone.validators.length ? `${zone.validators.length} validators` : `${rows.length} of ${zone.validators.length} validators match`;
  const table = rows.length === 0
    ? '<div class="note">no validators match the filters</div>'
    : `<div class="table-scroll"><table>${zoneHeader()}${rows.map((row) => zoneRow(row, zone)).join('')}</table></div>`;
  return `${selectedChainId === ALL_ZONES ? '' : zoneTiles(zone)}
    <div class="panel"><h2>${escapeHtml(zone.chain_id)} <span class="sub">${selectedChainId === ALL_ZONES ? zoneSummary(zone) + ' · ' : ''}delegation ICA ${addressCell(zone.delegation_address)} · ${shownNote}, largest |diff| first</span></h2>
      ${table}</div>`;
}

// The stacked all-zones view trades the tiles for one summary line per zone.
function zoneSummary(zone) {
  const totals = zone.totals;
  const amount = (value) => `${formatAmount(value, zone.decimals, 6)} ${zone.symbol}`;
  const diffClass = BigInt(totals.diff) > 0n ? 't-warn' : '';
  return `recorded ${amount(totals.recorded)} · actual ${amount(totals.actual)} · diff <span class="${diffClass}">${amount(totals.diff)}</span> (${totals.over_count} over, ${totals.red_count} red) · ${totals.in_progress_count} in progress`;
}

function zoneTiles(zone) {
  const totals = zone.totals;
  const amount = (value) => `${formatAmount(value, zone.decimals, 6)} ${zone.symbol}`;
  const diffClass = BigInt(totals.diff) > 0n ? 't-warn' : '';
  return `<div class="tiles">
      ${tile('Validators', totals.validator_count, `${totals.delegated_count} hold a delegation on the host`)}
      ${tile('Recorded', amount(totals.recorded), 'sum of Stride host_zone.validators')}
      ${tile('Actual', amount(totals.actual), 'delegation ICA on the host chain')}
      ${tile('Diff', amount(totals.diff), `recorded − actual · ${totals.over_count} over (${totals.red_count} red)`, diffClass)}
      ${tile('In progress', totals.in_progress_count, 'slash queries or delegation changes')}
    </div>`;
}

function zoneHeader() {
  return `<tr><th>Validator</th><th>Operator</th><th class="num">Weight</th><th class="num">Recorded</th><th class="num">Actual</th>
    <th class="num">Diff</th><th class="num">Diff %</th><th class="num">Rate diff</th><th class="num">Unbonding</th><th class="num" title="delegation_changes_in_progress">Δ in progress</th><th title="slash_query_in_progress">Slash query</th><th>Bond status</th></tr>`;
}

function zoneRow(row, zone) {
  const amount = (value) => formatAmount(value, zone.decimals, 6);
  const severityClass = SEVERITY_CLASS[row.severity];
  const notRegistered = row.registered ? '' : ` ${pill('warn', 'not registered')}`;
  return `<tr>
    <td>${escapeHtml(row.moniker)}${notRegistered}</td>
    <td>${addressCell(row.address)}</td>
    <td class="num">${escapeHtml(row.weight_percent)}%</td>
    <td class="num">${amount(row.recorded)}</td>
    <td class="num">${amount(row.actual)}</td>
    <td class="num ${severityClass}">${amount(row.diff)}</td>
    <td class="num ${severityClass}">${row.diff_percent === null ? '<span class="muted">–</span>' : escapeHtml(row.diff_percent) + '%'}</td>
    <td class="num ${rateDifferenceClass(row.rate_difference)}">${row.rate_difference === null ? '<span class="muted">n/a</span>' : escapeHtml(row.rate_difference)}</td>
    <td class="num ${unbondingClass(row.unbonding_entries, zone.max_unbonding_entries)}">${unbondingText(row.unbonding_entries, zone.max_unbonding_entries)}</td>
    <td class="num ${row.delegation_changes_in_progress ? 't-warn' : 'muted'}">${row.delegation_changes_in_progress === null ? '–' : row.delegation_changes_in_progress}</td>
    <td>${slashQueryCell(row.slash_query_in_progress)}</td>
    <td>${bondCell(row.bond_status, row.jailed)}</td></tr>`;
}

function rateDifferenceClass(difference) {
  return difference === null || Number(difference) === 0 ? 'muted' : 't-warn';
}

function unbondingText(entries, max) {
  return entries === null ? 'n/a' : `${entries} / ${max}`;
}

function unbondingClass(entries, max) {
  if (entries === null || entries === 0) return 'muted';
  return entries >= max ? 't-bad' : '';
}

function slashQueryCell(inProgress) {
  if (inProgress === null) return '<span class="muted">–</span>';
  return inProgress ? pill('warn', 'in progress') : '<span class="muted">no</span>';
}

function bondCell(status, jailed) {
  if (status === null) return '<span class="muted">n/a</span>';
  const jailedPill = jailed ? ` ${pill('bad', 'jailed')}` : '';
  return `${pill(status === 'bonded' ? 'ok' : 'idle', status)}${jailedPill}`;
}

// ---- staketia multisig

function multisigBody(entry) {
  const amount = (value) => `${formatAmount(value, entry.decimals, 6)} ${entry.symbol}`;
  const diffClass = SEVERITY_CLASS[entry.severity];
  const tiles = selectedChainId === ALL_ZONES ? '' : `<div class="tiles">
      ${tile('Validators', entry.validators.length, 'with a delegation from the multisig')}
      ${tile('Actual', amount(entry.actual_total), 'multisig delegations on Celestia')}
      ${tile('Remaining delegated', amount(entry.remaining_delegated_balance), 'staketia host_zone on Stride')}
      ${tile('Diff', amount(entry.diff), 'remaining − actual', diffClass)}
    </div>`;
  const summary = selectedChainId === ALL_ZONES
    ? `actual ${amount(entry.actual_total)} · remaining delegated ${amount(entry.remaining_delegated_balance)} · diff <span class="${diffClass}">${amount(entry.diff)}</span> · `
    : '';
  // The multisig has no per-validator recorded amounts, so the diff filters have nothing to match; the table only shows unfiltered.
  const table = filterActive()
    ? '<div class="note">no per-validator diffs for the multisig; untick the filters to list its delegations</div>'
    : `<div class="table-scroll"><table>
        <tr><th>Validator</th><th>Operator</th><th class="num">Actual</th><th>Bond status</th></tr>
        ${entry.validators.map((row) => `<tr><td>${escapeHtml(row.moniker)}</td><td>${addressCell(row.address)}</td>
          <td class="num">${formatAmount(row.actual, entry.decimals, 6)}</td><td>${bondCell(row.bond_status, row.jailed)}</td></tr>`).join('')}
      </table></div>`;
  return `${tiles}
    <div class="panel"><h2>staketia multisig <span class="sub">${summary}5-of-7 multisig ${addressCell(entry.delegation_address)} · on Celestia · no recorded per-validator amounts</span></h2>
      ${table}</div>`;
}

function tile(label, value, sub, valueClass = '') {
  return `<div class="tile"><div class="k">${escapeHtml(label)}</div><div class="v ${valueClass}">${escapeHtml(value)}</div><div class="s">${escapeHtml(sub)}</div></div>`;
}
})();
