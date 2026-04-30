# v33 Upgrade Test Fixtures

## Status

The v33 upgrade is currently exercised by **synthetic-state integration tests** in
`app/upgrades/v33/upgrades_test.go::TestUpgradeTestSuite`. Those tests seed a
controlled ICS validator set, govenator state, ICS reward module accounts, and
an active gov proposal — then run the upgrade handler and assert post-state.
They are the primary protection against regressions in the migration logic.

This directory exists for an additional **mainnet-export integration test** that
runs the v33 upgrade handler against real Stride mainnet state. That test is
not yet implemented because (a) it requires a fresh mainnet state export
(produced by an actual mainnet node), and (b) it requires a `SetupFromGenesis`
helper in `app/apptesting/test_helpers.go` that loads `StrideApp` from a
genesis JSON path and runs `InitChain` to populate state.

When ops is ready to validate v33 against a real export — usually shortly
before the binary ships — follow the workflow below.

## How to generate `mainnet_export.json.gz`

1. Sync a Stride mainnet node to a recent height (or use one you trust).
2. Stop the node.
3. Run `strided export --height <H> > mainnet_export.json`.
4. (Optional) Sanitize. Strip account keys / signatures if any are present.
   For this test, raw export is fine — we are not committing private data.
5. Compress: `gzip -9 mainnet_export.json`.
6. Move into this directory: `mv mainnet_export.json.gz app/upgrades/v33/testdata/`.
7. Commit. The file may be 50–200MB; consider git-lfs if size is a concern.
   Otherwise commit once and update only when the test specifically needs newer state.

The fixture is consumed exclusively by `TestUpgradeFromMainnetExport` (to be
implemented when the fixture lands).
