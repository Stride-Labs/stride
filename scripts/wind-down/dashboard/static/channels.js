// Channels tab: tiles, Stride <-> host channels, host -> Osmosis legs, holder routes into Osmosis.
(() => {

const STUCK_AFTER_SECONDS = 30 * 60;
const SOON_EXPIRY_SECONDS = 24 * 3600;
const ZONE_TABLE_COLUMNS = 12;

let routesOpen = false; // the holder-routes panel is collapsed until someone opens it

registerTab('channels', renderChannels);

function renderChannels(data, root) {
  const now = Date.now();
  root.innerHTML = [
    channelTiles(data.tiles),
    `<div class="panel"><h2>Stride ↔ host zones <span class="sub">transfer + ICA channels we relay · source: channels.main.stridenet.co + Stride/host REST and RPC</span></h2>
      <div class="table-scroll"><table>${zoneRows(data.zones, now)}</table></div></div>`,
    `<div class="panel"><h2>Host → Osmosis legs <span class="sub">what MsgTransferFromIca rides on · relayed by others</span></h2>
      <div class="table-scroll"><table>${legRows(data.legs, now)}</table></div></div>`,
    `<details class="panel" id="routesPanel" ${routesOpen ? 'open' : ''}>
      <summary><h2>Holder routes into Osmosis <span class="sub">from the relayer scope table · Osmosis side only</span></h2></summary>
      <div class="table-scroll"><table>${routeRows(data.routes, now)}</table></div></details>`,
  ].join('');

  document.getElementById('routesPanel').ontoggle = (event) => {
    routesOpen = event.target.open;
  };
}

// ---- tiles

function channelTiles(tiles) {
  const plus = tiles.pending_capped ? '+' : '';
  const nothingPending = tiles.pending_packets + tiles.pending_acks === 0;
  const oldestSub = tiles.oldest_pending_where || (nothingPending ? 'nothing pending' : 'age unknown (not in tx index)');
  const soonest = tiles.soonest_expiry_seconds === null ? 'n/a' : formatDuration(tiles.soonest_expiry_seconds);
  return `<div class="tiles">
    ${tile('Channels open', `${tiles.channels_open} / ${tiles.channels_total}`,
      `${tiles.channels_closed} closed · ${tiles.channels_handshake_stuck} handshake stuck`)}
    ${tile('Pending packets', `${tiles.pending_packets}${plus}`, `across ${tiles.pending_packet_channels} channel direction(s)`)}
    ${tile('Pending acks', `${tiles.pending_acks}${plus}`, `across ${tiles.pending_ack_channels} channel direction(s)`)}
    ${tile('Oldest pending', formatDuration(tiles.oldest_pending_seconds), oldestSub)}
    ${tile('Clients', `${tiles.clients_live} live`,
      `of ${tiles.clients_total} · soonest expiry ${soonest}${tiles.soonest_expiry_where ? ' (' + tiles.soonest_expiry_where + ')' : ''}`)}
  </div>`;
}

function tile(label, value, sub) {
  return `<div class="tile"><div class="k">${escapeHtml(label)}</div><div class="v">${escapeHtml(value)}</div><div class="s">${escapeHtml(sub)}</div></div>`;
}

// ---- Stride <-> host zones

function zoneRows(zones, now) {
  const header = `<tr><th>Channel</th><th>Stride side</th><th>Host side</th><th>State</th><th>Last sent</th><th>Last received</th><th>Last ack</th>
    <th class="num">Pending →</th><th class="num">Pending ←</th><th class="num">Pending acks</th><th>Oldest pending</th><th>Status</th></tr>`;
  return header + zones.map((zone) => (zone.error ? errorZoneRow(zone) : zoneHeaderRow(zone) + zone.channels.map((row) => channelRow(row, now)).join(''))).join('');
}

function errorZoneRow(zone) {
  return `<tr class="zone"><td colspan="${ZONE_TABLE_COLUMNS - 1}">${escapeHtml(zone.chain_id)}<span class="sub t-bad">${escapeHtml(zone.error)}</span></td><td>${pill('bad', 'error')}</td></tr>`;
}

function zoneHeaderRow(zone) {
  const hostConnection = zone.host_connection_state === null ? 'n/a' : escapeHtml(zone.host_connection_state);
  const hostDown = zone.host_error ? ` ${pill('warn', 'host REST unreachable')}` : '';
  const sub = `${zone.symbol} · Stride client ${clientPill(zone.stride_client)} · host client ${clientPill(zone.host_client)}
    · connection ${escapeHtml(zone.stride_connection_state)} / ${hostConnection}${hostDown}`;
  const title = zone.host_error ? ` title="${escapeHtml(zone.host_error)}"` : '';
  return `<tr class="zone"${title}><td colspan="${ZONE_TABLE_COLUMNS - 1}">${escapeHtml(zone.chain_id)}<span class="sub">${sub}</span></td><td>${statusPill(zone.status)}</td></tr>`;
}

function channelRow(row, now) {
  const inbound = row.inbound;
  const acks = sumKnown([row.outbound.pending_acks, inbound && inbound.pending_acks]);
  return `<tr>
    <td>${row.name === 'transfer' ? 'transfer' : 'ICA · ' + escapeHtml(row.name)}</td>
    <td class="mono">${escapeHtml(row.stride_channel)}</td>
    <td class="mono">${row.host_channel ? escapeHtml(row.host_channel) : '<span class="muted">–</span>'}</td>
    <td>${statePills(row.stride_state, row.host_state)}</td>
    ${activityCell(row.last_sent, now)}${activityCell(row.last_received, now)}${activityCell(row.last_ack, now)}
    ${pendingCell(row.outbound, 'pending_packets')}
    ${inbound ? pendingCell(inbound, 'pending_packets') : '<td class="num muted">–</td>'}
    <td class="num ${acks ? 't-warn' : ''}">${acks === null ? 'n/a' : acks + (anyCapped(row) ? '+' : '')}</td>
    ${oldestCell(row, now)}
    <td>${statusPill(row.status)}</td></tr>`;
}

function statePills(strideState, hostState) {
  if (strideState === 'OPEN' && hostState === 'OPEN') return pill('ok', 'OPEN');
  const describe = (state) => state || '–';
  const unreachable = strideState === 'OPEN' && hostState === 'UNREACHABLE';
  return pill(unreachable ? 'warn' : 'bad', `${describe(strideState)} / ${describe(hostState)}`);
}

function activityCell(reference, now) {
  if (!reference) return '<td class="muted">n/a</td>';
  return `<td class="muted" title="height ${reference.height}">${formatRelative(reference.time, now)}</td>`;
}

function pendingCell(flow, field) {
  const count = flow[field];
  if (count === null) return '<td class="num muted">n/a</td>';
  return `<td class="num ${count ? 't-warn' : ''}">${count}${flow.capped && count ? '+' : ''}</td>`;
}

function oldestCell(row, now) {
  const flows = [row.outbound, row.inbound].filter((flow) => flow && flow.oldest_sequence !== null);
  if (!flows.length) return '<td class="muted">–</td>';

  // The flow whose oldest commitment was sent first; unknown send times sort last.
  const sentAt = (flow) => (flow.oldest_sent_at ? Date.parse(flow.oldest_sent_at) : Infinity);
  const flow = flows.reduce((first, next) => (sentAt(next) < sentAt(first) ? next : first));
  if (!flow.oldest_sent_at) return `<td class="muted">seq ${flow.oldest_sequence}</td>`;

  const ageSeconds = (now - Date.parse(flow.oldest_sent_at)) / 1000;
  return `<td class="${ageSeconds > STUCK_AFTER_SECONDS ? 't-warn' : 'muted'}">${formatDuration(ageSeconds)} · seq ${flow.oldest_sequence}</td>`;
}

function sumKnown(values) {
  const known = values.filter((value) => value !== null && value !== undefined);
  return known.length ? known.reduce((total, value) => total + value, 0) : null;
}

function anyCapped(row) {
  return row.outbound.capped || Boolean(row.inbound && row.inbound.capped);
}

function clientPill(client) {
  if (!client) return '<span class="muted">n/a</span>';
  if (client.status !== 'Active') return pill('bad', client.status);

  const soon = client.seconds_remaining !== null && client.seconds_remaining < SOON_EXPIRY_SECONDS;
  return pill(soon ? 'warn' : 'ok', `live · ${formatDuration(client.seconds_remaining)}`);
}

// ---- host -> Osmosis legs and holder routes

function legRows(legs, now) {
  const header = `<tr><th>Zone</th><th>Host side</th><th>Osmosis side</th><th>State</th><th>Host client</th><th>Osmosis client</th>
    <th>Last received (Osmosis)</th><th>Last ack (host)</th><th>Status</th></tr>`;
  return header + legs.map((leg) => {
    if (leg.error) return errorRow(leg.chain_id, leg.error, 9);
    return `<tr><td>${escapeHtml(leg.chain_id)}</td><td class="mono">${escapeHtml(leg.host_channel)}</td>
      <td class="mono">${leg.osmosis_channel ? escapeHtml(leg.osmosis_channel) : '<span class="muted">–</span>'}</td>
      <td>${statePills(leg.host_state, leg.osmosis_state)}</td><td>${clientPill(leg.host_client)}</td><td>${clientPill(leg.osmosis_client)}</td>
      ${activityCell(leg.last_received, now)}${activityCell(leg.last_ack, now)}<td>${statusPill(leg.status)}</td></tr>`;
  }).join('');
}

function routeRows(routes, now) {
  const header = `<tr><th>Chain</th><th>Osmosis channel</th><th>Chain side</th><th>Relayed by</th><th>State</th><th>Client</th>
    <th>Last received</th><th>Last ack</th><th>Status</th></tr>`;
  return header + routes.map((route) => {
    if (route.error) return errorRow(route.chain_id, route.error, 9);
    const relayedClass = route.relayed_by === 'free' ? 'ok' : route.relayed_by === 'ours' ? 'idle' : 'warn';
    return `<tr><td>${escapeHtml(route.chain)}</td><td class="mono">${escapeHtml(route.osmosis_channel)}</td>
      <td class="mono">${route.counterparty_channel ? escapeHtml(route.counterparty_channel) : '<span class="muted">–</span>'}</td>
      <td>${pill(relayedClass, route.relayed_by)}</td><td>${statePills(route.state, route.state)}</td><td>${clientPill(route.client)}</td>
      ${activityCell(route.last_received, now)}${activityCell(route.last_ack, now)}<td>${statusPill(route.status)}</td></tr>`;
  }).join('');
}

function errorRow(name, message, columns) {
  return `<tr><td>${escapeHtml(name)}</td><td colspan="${columns - 2}" class="t-bad">${escapeHtml(message)}</td><td>${pill('bad', 'error')}</td></tr>`;
}
})();
