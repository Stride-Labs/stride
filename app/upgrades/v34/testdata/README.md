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

The suite skips when the file is absent.

## Committed fixture provenance

**POA section** assembled 2026-09-11 at mainnet height **40202716**; **injective-1
host zone** assembled 2026-09-15 at mainnet height **40302363**; **celestia and
cosmoshub-4 host zones, records and staketia sections** assembled 2026-09-18 at
mainnet height **40364142**. All from the public REST API, which serves the same
module state a `strided export` emits for these sections (the REST field names
match the module genesis protos, and string-typed integers decode through the
app codec exactly as an export does):

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
                 staketia: {host_zone: .app_state.staketia.host_zone}}}' \
  full_export.json > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
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
  (`GetUndelegatedBalance + TotalDelegations`) are unchanged to the utia. The
  in-progress records reconcile without icacallbacks entries, which the
  fixture does not carry: the handler only removes callbacks it finds.
- Staketia: `remaining_delegated_balance` moves by exactly
  `StaketiaRemainingDelegatedBalanceDelta` (a negative result would have been
  skipped), and stakeibc's celestia `TotalDelegations` is not mirrored.
- Cosmos Hub: the export carries the stranded stakewithus LSM deposit in
  `DETOKENIZATION_FAILED` with the constant's amount and validator, the
  close-out removes it, leaves the other cosmoshub-4 deposits untouched, and
  raises stakewithus and `TotalDelegations` by exactly 10,999,999 uatom.

Note the snapshots are point-in-time: they catch a wrong address, a delta that
no longer fits or a record that changed status, not drift that happens after
the snapshot. Still re-measure every table and delta immediately before the
proposal (the measurement commands are in `celestia.go`, `cosmoshub.go` and
`injective.go`). If the POA set, any host zone's validator set, the celestia
deposit records, the LSM deposit or the staketia balance changes on mainnet
between now and the release, regenerate the fixture together with the
constants so the gate tests against current state.
