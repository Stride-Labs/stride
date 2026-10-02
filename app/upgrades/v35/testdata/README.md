# v35 mainnet export fixture

`mainnet_export.json.gz` is mainnet state trimmed to the sections the v35 mainnet-export
suite consumes. The suite (`../mainnet_export_test.go`) replays the whole v35 handler
against it with the real constants and asserts every effect of spec §5; it skips when the
file is absent so CI stays green before release prep.

## Sections the suite consumes (all under `app_state`)

- `stakeibc.host_zone_list`: every host zone (the eleven in-scope zones, comdex-1 and the
  three deprecated ones). Drives the stale-flag reset (validators'
  `delegation_changes_in_progress`, `connection_id`), the comdex flag, the haqq delta table
  (validators' `delegation`, `address`) and the haqq slash-flag reset (`slash_query_in_progress`).
- `stakeibc.trade_routes`: the `uusdc`/`adydx` route.
- `interchainquery.queries`: every pending query (`chain_id`, `callback_id`), for the two purges.
- `autopilot.params`: `stakeibc_active`.
- `interchainaccounts.host_genesis_state.params`: the ICA host allow-list, as the ICA genesis
  proto names it.
- `wasm.params` (`code_upload_access`) and `wasm.contracts[]` (`contract_address`,
  `contract_info.admin`) for the contracts whose admin is the deploy key: the suite
  instantiates one stand-in contract per entry with that admin and asserts on those.
- `icaoracle.oracles`: the three active oracles.
- `ratelimit.rate_limits`, `ratelimit.blacklisted_denoms`, `ratelimit.whitelisted_address_pairs`.
- `records.epoch_unbonding_record_list`: every epoch unbonding record, asserted untouched.
- `delegation_channels` (not a real export field): chain id -> `{connection_id, channel_id,
  packet_commitments}` for every in-scope zone's delegation ICA, so the suite can register
  each open active channel the way mainnet has it and assert the stale-flag reset on exactly
  the zones with zero commitments. `channel_id` is empty when the zone has no open delegation
  channel.

Stakedym is deprecated and left halted by the handler (no stakedym step), so the fixture
carries no stakedym state. On mainnet its BeginBlocker re-adds `stadym` to the rate-limit
blacklist every block, so the suite allows `stadym` (and nothing else) to be blacklisted after
the upgrade.

## Assembly

Pick a recent height `H` (`curl -s -H 'User-Agent: Mozilla/5.0' https://stride-rpc.polkachu.com/status |
jq -r .result.sync_info.latest_block_height`, minus a few blocks so the node has the state) and
pin every request to it. The public REST needs a browser-like `User-Agent`, answers `429`
under load (retry with backoff) and does not implement `/ibc/core/channel/v1/channels`
pagination reliably, so channels are listed per connection.

```bash
API=https://stride-api.polkachu.com
H=<H>
get() { curl -s -H 'User-Agent: Mozilla/5.0' -H "x-cosmos-block-height: $H" "$1"; }
mkdir -p /tmp/v35-fixture && cd /tmp/v35-fixture

get "$API/Stride-Labs/stride/stakeibc/host_zone"                                     > host_zones.json
get "$API/Stride-Labs/stride/stakeibc/trade_routes"                                  > trade_routes.json
get "$API/Stride-Labs/stride/interchainquery/pending_queries"                        > queries.json
get "$API/Stride-Labs/stride/autopilot/params"                                       > autopilot.json
get "$API/ibc/apps/interchain_accounts/host/v1/params"                               > icahost.json
get "$API/cosmwasm/wasm/v1/codes/params"                                             > wasm_params.json
get "$API/Stride-Labs/stride/icaoracle/oracles"                                      > oracles.json
get "$API/ibc/apps/rate-limiting/v1/ratelimits"                                      > rate_limits.json
get "$API/ibc/apps/rate-limiting/v1/ratelimit/blacklisted_denoms"                    > blacklist.json
get "$API/ibc/apps/rate-limiting/v1/ratelimit/whitelisted_addresses"                 > whitelist.json
get "$API/Stride-Labs/stride/records/epoch_unbonding_record?pagination.limit=2000"   > epoch_unbonding.json

# The contracts whose admin is the deploy key (spec §3; four on mainnet: the Hyperlane IGP, IGP
# hook, aggregate hook and multisig ISM). Every code's contracts are listed, then each one's info.
get "$API/cosmwasm/wasm/v1/code?pagination.limit=200" > codes.json
: > all_contracts.txt
for code in $(jq -r '.code_infos[].code_id' codes.json); do
  get "$API/cosmwasm/wasm/v1/code/$code/contracts?pagination.limit=200" | jq -r '.contracts[]' >> all_contracts.txt
  sleep 0.5
done
sort -u all_contracts.txt -o all_contracts.txt
: > contracts.tmp
while read -r addr; do
  get "$API/cosmwasm/wasm/v1/contract/$addr" | jq -c '{contract_address: .address, contract_info: .contract_info}' >> contracts.tmp
  sleep 0.5
done < all_contracts.txt
jq -s '[.[] | select(.contract_info.admin == "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh")]' contracts.tmp > deploy_key_contracts.json
jq length deploy_key_contracts.json   # expected 4 (spec §3); if not, note the new count below

# Delegation ICA channels per in-scope zone: the open channel on the delegation port and how many
# packet commitments it holds. Listed per connection.
: > dc.tmp
for chain in celestia cosmoshub-4 dydx-mainnet-1 haqq_11235-1 injective-1 juno-1 laozi-mainnet osmosis-1 phoenix-1 sommelier-3 ssc-1; do
  port="icacontroller-$chain.DELEGATION"
  conn=$(jq -r --arg c "$chain" '.host_zone[] | select(.chain_id == $c) | .connection_id' host_zones.json)
  chan=$(get "$API/ibc/core/channel/v1/connections/$conn/channels?pagination.limit=1000" \
         | jq -r --arg p "$port" '[.channels[] | select(.port_id == $p and .state == "STATE_OPEN")] | last | .channel_id // empty')
  if [ -z "$chan" ]; then
    jq -nc --arg c "$chain" --arg conn "$conn" '{($c): {connection_id: $conn, channel_id: "", packet_commitments: 0}}' >> dc.tmp
    continue
  fi
  n=$(get "$API/ibc/core/channel/v1/channels/$chan/ports/$port/packet_commitments" | jq '.commitments | length')
  jq -nc --arg c "$chain" --arg conn "$conn" --arg chan "$chan" --argjson n "$n" \
     '{($c): {connection_id: $conn, channel_id: $chan, packet_commitments: $n}}' >> dc.tmp
  sleep 0.5
done
jq -s add dc.tmp > delegation_channels.json

jq -n --slurpfile hz host_zones.json --slurpfile tr trade_routes.json --slurpfile q queries.json \
      --slurpfile ap autopilot.json --slurpfile ih icahost.json --slurpfile wp wasm_params.json \
      --slurpfile wc deploy_key_contracts.json --slurpfile or oracles.json \
      --slurpfile rl rate_limits.json --slurpfile bl blacklist.json --slurpfile wl whitelist.json \
      --slurpfile eu epoch_unbonding.json --slurpfile dc delegation_channels.json \
  '{app_state: {
      stakeibc: {host_zone_list: $hz[0].host_zone, trade_routes: $tr[0].trade_routes},
      interchainquery: {queries: $q[0].pending_queries},
      autopilot: {params: $ap[0].params},
      interchainaccounts: {host_genesis_state: {params: $ih[0].params}},
      wasm: {params: $wp[0].params, contracts: $wc[0]},
      icaoracle: {oracles: $or[0].oracles},
      ratelimit: {rate_limits: $rl[0].rate_limits, blacklisted_denoms: $bl[0].denoms, whitelisted_address_pairs: $wl[0].address_pairs},
      records: {epoch_unbonding_record_list: $eu[0].epoch_unbonding_record},
      delegation_channels: $dc[0]
   }}' > trimmed.json
gzip -9 -c trimmed.json > <repo>/app/upgrades/v35/testdata/mainnet_export.json.gz
```

If a REST field name above differs from the module's genesis JSON name (the suite decodes each
section with the app codec), fix the `jq` key, not the Go: the names used are the proto field
names of each module's `GenesisState`.

## Committed fixture provenance

Assembled 2026-09-29 (evening US time, block time about 2026-09-30 03:08 UTC) at mainnet
height **40837200** from `stride-api.polkachu.com`, every section pinned to that height with
`x-cosmos-block-height`. Deploy-key contracts found: 4 (the Hyperlane IGP, IGP hook, aggregate
hook and multisig ISM, code ids 30, 31 and 44, out of 21 contracts on 44 codes). Zones with an
open delegation channel and zero packet commitments at that height: cosmoshub-4, dydx-mainnet-1,
injective-1, osmosis-1, phoenix-1, sommelier-3, ssc-1. juno-1 has an open channel with 16
packets in flight; celestia, haqq_11235-1 and laozi-mainnet have no open delegation channel.

The haqq delta table and its tracked-delegation pins (`HaqqDelegationDeltas` and
`HaqqExpectedTrackedDelegations` in `../haqq.go`) were generated earlier than this fixture. The
fixture's haqq host zone was compared with the pins at height 40837200: all sixteen pinned
validators are tracked at exactly the pinned values, so the table applies in full against this
fixture and the suite asserts that. The table and pins were not regenerated. The table itself
should still be regenerated (`scripts/wind-down/gen_delta_table.py`) at the proposal height, and
the fixture re-assembled at the same height.

## Staleness gate

`python3 app/upgrades/v35/testdata/verify_constants.py` checks that the haqq pins equal the live
tracked delegations, recomputes the haqq delta table from the live chain, checks both channel
maps against the hosts' and Stride's IBC state, and checks that both operator addresses have
signed at least once on their chain. Run it right before cutting the release; it exits non-zero
on any drift.

## Regenerating from a node export

From a synced node: `strided export --height <H> > full_export.json`, then a `jq` that keeps
`stakeibc.{host_zone_list,trade_routes}`, `interchainquery.queries`, `autopilot.params`,
`interchainaccounts.host_genesis_state.params`, `wasm.params` and the deploy-key entries of
`wasm.contracts`, `icaoracle.oracles`, `ratelimit.{rate_limits,blacklisted_denoms,
whitelisted_address_pairs}` and `records.epoch_unbonding_record_list`; `delegation_channels`
is still built from the IBC queries above, since an export does not say which channel is
active.
