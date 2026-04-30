# v33 Upgrade Test Fixtures

## Status

The v33 upgrade is exercised by two integration tests:

1. **Synthetic-state test** — `upgrades_test.go::TestUpgradeTestSuite`. Seeds a
   controlled ICS validator set, ICS reward module accounts, and an active gov
   proposal directly via test helpers, then runs the handler. Fast, granular
   assertions, but the test app already has POA wired so it's closer to
   "fresh-binary against seeded ICS state" than a true v32→v33 migration.

2. **Mainnet-export test** — `mainnet_export_test.go::TestUpgradeFromMainnetExport`.
   Loads real post-v32 mainnet state (this fixture) into the relevant module
   keepers, then runs the handler. More representative of production state
   shape (8 real validators, real ICS account balances, real gov state)
   while still allowing granular Go-level assertions. The test is gated on
   the fixture being present and `t.Skip`s otherwise.

Both layer below the localstride mainnet-state-export testnet (which validates
block production end-to-end without granular assertions) and the k8s network
test (most prod-faithful, but no in-process assertions).

## How to generate `mainnet_export.json.gz`

The raw `strided export` output at current mainnet height is typically
150–300 MB — far too large to commit. The trim step below keeps only the
modules touched by the v33 handler (or asserted against by the test) and
discards everything else (`ibc`, `transfer`, `wasm`, all the unrelated
Stride domain modules, etc.). Trimmed output is normally < 10 MB compressed.

```bash
# 1. Sync a Stride mainnet node to a recent height (or use one you trust).
# 2. Stop the node.
# 3. Export.
strided export --height <H> > /tmp/mainnet_export_raw.json

# 4. Trim to v33-relevant modules and gzip.
./scripts/trim_export.sh /tmp/mainnet_export_raw.json testdata/mainnet_export.json.gz

# 5. Commit. The trimmed fixture is small enough for plain git
#    (no git-lfs needed). Bump only when the test needs newer state.
```

The trim script is `app/upgrades/v33/scripts/trim_export.sh` — see the header
comment for the exact module keep-list and rationale.
