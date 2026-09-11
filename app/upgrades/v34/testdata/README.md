# v34 mainnet export fixture

`mainnet_export.json.gz` is a trimmed `strided export` of post-v33 mainnet
state, containing only the `poa` module section. It is NOT committed by
default — the v34 mainnet-export test suite skips when it is absent.

Generate during release prep (requires a synced post-v33 node):

```bash
strided export --height <recent height> > full_export.json
jq '{app_state: {poa: .app_state.poa}}' full_export.json > trimmed.json
gzip -c trimmed.json > app/upgrades/v34/testdata/mainnet_export.json.gz
```

The suite runs with the REAL v34 constants (no test-key substitution) and
verifies them against actual mainnet state, including the POA-set ≡
payout-registry invariant — run it as the final release gate before tagging.
