# v34 mainnet export fixture

`mainnet_export.json.gz` is post-v33 mainnet state trimmed to the sections the
v34 mainnet-export suite consumes:

- `app_state.poa` — `{params, validators}` for the POA validator swap.
- `app_state.stakeibc` — `{host_zone_list: [injective-1, celestia, cosmoshub-4]}`
  for the Injective and Celestia delegation reconciliations and the Cosmos Hub
  LSM deposit close-out.
- `app_state.records` — `{deposit_record_list, lsm_token_deposit_list}`: the
  celestia deposit records only (every status, real ids) and the cosmoshub-4
  LSM token deposits only.
- `app_state.staketia` — `{host_zone}`: the staketia (multisig) host zone for
  the remaining delegated balance correction.
- `app_state.icacallbacks` — `{callback_data_list}`: every delegate callback on
  the celestia DELEGATION port's active channel (255 entries, one per current
  `DELEGATION_IN_PROGRESS` record) plus a representative sample of dead-channel
  entries (3 per dead channel, 321 entries across all 109 dead channels the
  port has ever used) and all 24 `rebalance`-id entries, so
  `removeDelegateCallbacks` is exercised end to end: real callback removals,
  real survivors, and real non-delegate/dead-channel noise it must leave alone.
- `app_state.icacallbacks_active_channel` — not a real export field: a
  single string naming the celestia DELEGATION port's active channel id
  (`channel-862` as of this snapshot), so the suite can register it as the
  open ICA channel the way mainnet has it, without querying IBC channel state
  itself.

The suite skips when the file is absent.

## Committed fixture provenance

**POA section** assembled 2026-09-11 at mainnet height **40202716**; **injective-1
host zone** assembled 2026-09-15 at mainnet height **40302363**; **celestia and
cosmoshub-4 host zones, records, staketia and icacallbacks sections** assembled
2026-09-18 at mainnet height **40364142**. All from the public REST API, which
serves the same module state a `strided export` emits for these sections (the
REST field names match the module genesis protos, and string-typed integers
decode through the app codec exactly as an export does):

```bash
API=https://stride-api.polkachu.com
curl -s $API/cosmos/poa/v1/params                              > poa_params.json
curl -s $API/cosmos/poa/v1/validators                          > poa_vals.json
curl -s $API/Stride-Labs/stride/stakeibc/host_zone/injective-1 > inj_hz.json

# The newer sections are pinned to one height so every piece is one consistent snapshot
H='x-cosmos-block-height: 40364142'
curl -s -H "$H" $API/Stride-Labs/stride/stakeibc/host_zone/celestia    > tia_hz.json
curl -s -H "$H" $API/Stride-Labs/stride/stakeibc/host_zone/cosmoshub-4 > hub_hz.json
curl -s -H "$H" "$API/Stride-Labs/stride/records/deposit_record?pagination.limit=1000"         > deposit_records.json
curl -s -H "$H" "$API/Stride-Labs/stride/stakeibc/lsm_deposits?chain_id=cosmoshub-4"           > hub_lsm.json
curl -s -H "$H" $API/Stride-Labs/stride/staketia/host_zone                                     > staketia_hz.json

jq -n --slurpfile p poa_params.json --slurpfile v poa_vals.json \
      --slurpfile inj inj_hz.json --slurpfile tia tia_hz.json --slurpfile hub hub_hz.json \
      --slurpfile dr deposit_records.json --slurpfile lsm hub_lsm.json --slurpfile stk staketia_hz.json \
  '{app_state: {poa: {params: $p[0].params, validators: $v[0].validators},
                stakeibc: {host_zone_list: [$inj[0].host_zone, $tia[0].host_zone, $hub[0].host_zone]},
                records: {deposit_record_list: [$dr[0].deposit_record[] | select(.host_zone_id == "celestia")],
                          lsm_token_deposit_list: [$lsm[0].deposits[] | select(.chain_id == "cosmoshub-4")]},
                staketia: {host_zone: $stk[0].host_zone}}}' > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
```

Some public endpoints reject the default curl user agent; add `-H 'User-Agent: curl/8.0'`
if a request returns an HTML error page. The committed fixture was assembled by
merging the height-40364142 pieces into the previously committed POA and
injective-1 sections, which are byte-identical to the earlier fixture.

### icacallbacks section (Celestia DELEGATION port)

The `callback_data` endpoint is not filterable by port and paginates in pages
of ~5000 across the whole chain (~22k entries chain-wide at this height); fetch
every page with `-H "$H"` and a `User-Agent`, then filter client-side:

```bash
key=""
> all_callbacks.json.tmp
while :; do
  url="$API/Stride-Labs/stride/icacallbacks/callback_data?pagination.limit=5000"
  [ -n "$key" ] && url="$url&pagination.key=$key"
  page=$(curl -s -H "User-Agent: curl/8.0" -H "$H" "$url")
  echo "$page" | jq -c '.callback_data[]' >> all_callbacks.json.tmp
  key=$(echo "$page" | jq -r '.pagination.next_key // empty')
  [ -z "$key" ] && break
done
jq -s '[.[] | select(.port_id == "icacontroller-celestia.DELEGATION")]' all_callbacks.json.tmp > tia_callbacks.json
```

The active channel id came from confirming the channel is open at the same height:

```bash
curl -s -H "$H" "$API/ibc/core/channel/v1/channels/channel-862/ports/icacontroller-celestia.DELEGATION"
# -> state STATE_OPEN, counterparty channel-700 on icahost; connection-125 matches the
#    celestia host zone's connection_id in the stakeibc section above
```

`tia_callbacks.json` has 19,845 entries (~47MB decompressed) — too large to
commit whole. The committed fixture keeps every entry on `channel-862` (255,
one per current `DELEGATION_IN_PROGRESS` record — enough to exercise every
branch of `removeDelegateCallbacks` deterministically) plus a representative
dead-channel sample (the first 3 `delegate`-id entries per distinct dead
channel id, 321 entries across 109 channels) plus every `rebalance`-id entry
(24, on `channel-859`/`channel-481`, to prove non-delegate callbacks are never
touched):

```bash
jq '[.[] | select(.channel_id == "channel-862")]
    + (group_by(.channel_id) | map(select(.[0].channel_id != "channel-862" and .[0].callback_id == "delegate")) | map(.[0:3]) | add)
    + [.[] | select(.callback_id == "rebalance")]' tia_callbacks.json > tia_callbacks_sample.json

jq --slurpfile cb tia_callbacks_sample.json '.app_state.icacallbacks = {callback_data_list: $cb[0]}
    | .app_state.icacallbacks_active_channel = "channel-862"' trimmed.json > trimmed_with_callbacks.json
gzip -c trimmed_with_callbacks.json > app/upgrades/v34/testdata/mainnet_export.json.gz
```

POA cross-validation at assembly time: the fixture's pubkeys and powers match
the live CometBFT signing set (`stride-rpc.polkachu.com/validators`, a separate
consensus-layer data path), and an independent provider
(`rest.cosmos.directory/stride`) agreed on every field for all 8 validators.
The only state a real `strided export` would add is POA fee-accounting
(`allocated_fees` etc.), which the suite never reads.

## Regenerating from a node export (byte-provenance alternative)

From a synced post-v33 node's data dir:

```bash
strided export --height <recent height> > full_export.json
jq '{app_state: {poa: .app_state.poa,
                 stakeibc: {host_zone_list: [.app_state.stakeibc.host_zone_list[]
                              | select(.chain_id == "injective-1" or .chain_id == "celestia" or .chain_id == "cosmoshub-4")]},
                 records: {deposit_record_list: [.app_state.records.deposit_record_list[] | select(.host_zone_id == "celestia")],
                           lsm_token_deposit_list: [.app_state.records.lsm_token_deposit_list[] | select(.chain_id == "cosmoshub-4")]},
                 staketia: {host_zone: .app_state.staketia.host_zone},
                 icacallbacks: {callback_data_list: [.app_state.icacallbacks.callback_data_list[]
                              | select(.port_id == "icacontroller-celestia.DELEGATION")]}}}' \
  full_export.json > trimmed_full_callbacks.json
# then apply the same channel-862 + per-channel-sample + rebalance filter as above to
# trimmed_full_callbacks.json's icacallbacks.callback_data_list, and add
# app_state.icacallbacks_active_channel by querying the node's IBC channel state directly
# (`strided q ibc channel end icacontroller-celestia.DELEGATION <channel>` for the OPEN one)
gzip -c trimmed_full_callbacks.json > app/upgrades/v34/testdata/mainnet_export.json.gz
```

## What the gate checks

The suite runs with the REAL v34 constants (no test-key substitution) and
verifies them against actual mainnet state — it is the final release gate
before tagging. Every accounting fix in the handler deliberately skips (rather
than errors) when its constants no longer describe chain state, so the suite
asserts the applied *effects*; that is what turns a stale constant into a red
build instead of a silently deferred reconciliation.

- POA: the confirmed pubkeys and payout addresses, and the POA-set ≡
  payout-registry invariant.
- Injective: the delta table applies in full against the real host zone —
  every table address is a host zone validator and no delta drives a
  delegation negative — so the handler queues exactly the whole-table sum as
  the pending undelegation.
- Celestia: the delta table applies in full (every validator moved by its
  delta, `TotalDelegations` up by the table sum), exactly the table sum is
  removed from the real celestia deposit records (which must cover it),
  transfer-status records are untouched, and the redemption rate components
  (`GetUndelegatedBalance + TotalDelegations`) are unchanged to the utia.
  `removeDelegateCallbacks` is exercised against the real callback set: every
  delegate callback whose `DepositRecordId` was a `DELEGATION_IN_PROGRESS`
  record that got deleted this run is gone (on any channel), every other
  delegate and non-delegate callback is untouched, and each validator's
  `DelegationChangesInProgress` moved down by exactly the number of splits it
  was named in among the removed *active-channel* callbacks (never below
  zero) — dead-channel removals are orphans and never decrement anything.
  Real pre-existing dangling callbacks outside this snapshot (mainnet noise
  from the ~90 channel deaths) are left alone, as expected.
- Staketia: `remaining_delegated_balance` moves by exactly
  `StaketiaRemainingDelegatedBalanceDelta` (a negative result would have been
  skipped), and stakeibc's celestia `TotalDelegations` is not mirrored.
- Cosmos Hub: the export carries the stranded stakewithus LSM deposit in
  `DETOKENIZATION_FAILED` with the constant's amount and validator, the
  close-out removes it, leaves the other cosmoshub-4 deposits untouched,
  raises stakewithus and `TotalDelegations` by exactly 10,999,999 uatom, and
  the redemption rate components (`GetTotalTokenizedDelegations +
  TotalDelegations`) are unchanged to the uatom.

Note the snapshots are point-in-time: they catch a wrong address, a delta that
no longer fits or a record that changed status, not drift that happens after
the snapshot. Still re-measure every table and delta immediately before the
proposal (the measurement commands are in `celestia.go`, `cosmoshub.go` and
`injective.go`). If the POA set, any host zone's validator set, the celestia
deposit records, the LSM deposit, the staketia balance, or the celestia
DELEGATION port's active channel changes on mainnet between now and the
release, regenerate the fixture together with the constants (including
`CelestiaDelegationChannelId`) so the gate tests against current state.

## The pinned delegation channel guard

`ReconcileCelestia` applies only if the celestia DELEGATION port's active
channel is still `CelestiaDelegationChannelId` (measured with the delta table)
and that channel has zero packet commitments. A delegate that executes on the
host without its ack being booked is the only thing that changes the delta
table, and such a packet leaves its commitment on Stride forever, so an
unchanged, empty channel proves the table is still exact. A restore would move
the active channel and hide that evidence, which is why the id is pinned.

Measure the constants only while that channel has zero packet commitments
(`verify_constants.py` checks this before and after the delta fetch): a packet
that executed unbooked is already counted as phantom, and an ack booked after
the measurement lowers the delta while erasing the commitment that would have
tripped the guard. Note the committed fixture was assembled with 255 delegate
callbacks still in flight on channel-862; the export suite mocks the channel
without their commitments, i.e. it rehearses the post-clear state the handler
will actually see, not the fixture-height state (which the guard would skip).
Regenerate the fixture in the quiet state at release prep.

The guard does not cover a Celestia slash on a table validator between the
last script run and the upgrade block; that over-books the slashed validator by
the slash amount until the routine slash query corrects it. Rare, and rate-safe.

Operationally: from constant measurement to the upgrade block, do not restore
the celestia DELEGATION channel, and make sure it has no unacknowledged
packets at the block (a closed channel with its timeouts relayed qualifies).
If the guard trips, the Celestia step skips with a log and every other v34
step still applies; re-measure and reconcile in the next upgrade.

## Release prep: run `python3 app/upgrades/v34/testdata/verify_constants.py`

The mainnet-export gate above only checks the constants against a fixture
snapshot from whenever it was assembled; it cannot catch drift that happens
between that snapshot and the actual release. The handler itself cannot
close that gap either: tracked delegations legitimately move as daily
delegations are acknowledged, so baking an on-chain fingerprint into
`ReconcileCelestia`/`CloseCosmosHubLsmDeposit` would false-skip a perfectly
valid reconciliation the moment a single ack lands between now and the
upgrade. `verify_constants.py` is the staleness gate that fills this gap: it
re-derives `CelestiaDelegationDeltas`, `StaketiaRemainingDelegatedBalanceDelta`
and `CosmosHubStrandedLsmDeposit` straight from `celestia.go`/`cosmoshub.go`,
recomputes every one of them against the *live* chain (per-validator Celestia
deltas, the staketia delta formula, and the Hub LSM record and delegation
gap, plus the Hub tokenized-rate invariance the close-out guard depends on),
cross-checks them against the committed fixture's record sums, and prints
PASS/FAIL per check with a non-zero exit on any failure. Run it right before
cutting the release, alongside re-measuring the tables by hand:

```bash
python3 app/upgrades/v34/testdata/verify_constants.py
```
