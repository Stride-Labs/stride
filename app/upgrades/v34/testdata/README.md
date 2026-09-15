# v34 mainnet export fixture

`mainnet_export.json.gz` is post-v33 mainnet state trimmed to the two sections
the v34 mainnet-export suite consumes:

- `app_state.poa` — `{params, validators}` for the POA validator swap.
- `app_state.stakeibc` — `{host_zone_list: [injective-1]}` for the Injective
  delegation reconciliation.

The suite skips when the file is absent.

## Committed fixture provenance

**POA section** assembled 2026-09-11 at mainnet height **40202716**; **stakeibc
section** assembled 2026-09-15 at mainnet height **40302363**. Both from the
public REST API, which serves the same module state a `strided export` emits
for these sections:

```bash
API=https://stride-api.polkachu.com
curl -s $API/cosmos/poa/v1/params                              > poa_params.json
curl -s $API/cosmos/poa/v1/validators                          > poa_vals.json
curl -s $API/Stride-Labs/stride/stakeibc/host_zone/injective-1 > inj_hz.json
jq -n --slurpfile p poa_params.json --slurpfile v poa_vals.json --slurpfile hz inj_hz.json \
  '{app_state: {poa: {params: $p[0].params, validators: $v[0].validators},
                stakeibc: {host_zone_list: [$hz[0].host_zone]}}}' > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
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
                 stakeibc: {host_zone_list: [.app_state.stakeibc.host_zone_list[] | select(.chain_id == "injective-1")]}}}' \
  full_export.json > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
```

## What the gate checks

The suite runs with the REAL v34 constants (no test-key substitution) and
verifies them against actual mainnet state — it is the final release gate
before tagging:

- POA: the confirmed pubkeys and payout addresses, and the POA-set ≡
  payout-registry invariant.
- Injective: the delta table applies in full against the real host zone —
  every table address is a host zone validator and no delta drives a
  delegation negative — so the handler queues exactly the whole-table sum as
  the pending undelegation. The handler deliberately skips (rather than
  errors) on an inconsistent table, so this test is what turns a stale table
  into a red build instead of a silently deferred reconciliation.

Note the host-zone snapshot is point-in-time: it catches a wrong address or a
delta that no longer fits, not drift that happens after the snapshot. Still
re-measure the table against `injective-1` immediately before the proposal.
If the POA set or the Injective validator set changes on mainnet between now
and the release, regenerate the fixture so the gate tests against current
state.
