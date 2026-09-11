# v34 mainnet export fixture

`mainnet_export.json.gz` is post-v33 mainnet state trimmed to the `poa`
module section (`{app_state: {poa: {params, validators}}}`) — the only state
the v34 mainnet-export suite consumes. The suite skips when the file is
absent.

## Committed fixture provenance

Assembled 2026-09-11 at mainnet height **40202716** from the public REST API
(`/cosmos/poa/v1/params` + `/cosmos/poa/v1/validators`), which serves the
same module state a `strided export` emits for the `poa` section:

```bash
curl -s https://stride-api.polkachu.com/cosmos/poa/v1/params    > poa_params.json
curl -s https://stride-api.polkachu.com/cosmos/poa/v1/validators > poa_vals.json
jq -n --slurpfile p poa_params.json --slurpfile v poa_vals.json \
  '{app_state: {poa: {params: $p[0].params, validators: $v[0].validators}}}' > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
```

Cross-validated at assembly time: the fixture's pubkeys and powers match the
live CometBFT signing set (`stride-rpc.polkachu.com/validators`, a separate
consensus-layer data path), and an independent provider
(`rest.cosmos.directory/stride`) agreed on every field for all 8 validators.
The only state a real `strided export` would add is POA fee-accounting
(`allocated_fees` etc.), which the suite never reads.

## Regenerating from a node export (byte-provenance alternative)

From a synced post-v33 node's data dir:

```bash
strided export --height <recent height> > full_export.json
jq '{app_state: {poa: .app_state.poa}}' full_export.json > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
```

The suite runs with the REAL v34 constants (no test-key substitution) and
verifies them against actual mainnet state, including the POA-set ≡
payout-registry invariant — it is the final release gate before tagging. If
the POA set changes on mainnet between now and the release, regenerate the
fixture so the gate tests against current state.
