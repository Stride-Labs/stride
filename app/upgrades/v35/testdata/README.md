# v35 mainnet export fixture

`mainnet_export.json.gz` is not committed by this PR. The release gate (PR 6) adds the
mainnet-export suite that replays the v35 handler against real state with the real
constants; it skips when the file is absent. This README fixes what that fixture must
contain so it can be assembled from the public REST API right before the proposal, the
way v34's was (see `app/upgrades/v34/testdata/README.md` for the assembly commands and
the trimming rules). This list is the first draft; the PR 6 plan's fixture task rewrites
it with the final section names and the assembly commands, and that version supersedes
this one.

Sections the v35 suite consumes, all under `app_state`:

- `stakeibc.host_zone_list`: every host zone (the eleven in-scope zones, comdex-1 and the
  three deprecated ones). Drives the stale-flag reset (validators' `delegation_changes_in_progress`,
  `connection_id`), the comdex flag, the haqq delta table (validators' `delegation`,
  `address`) and the haqq slash-flag reset (`slash_query_in_progress`).
- `stakeibc.trade_routes`: the `uusdc`/`adydx` route.
- `interchainquery.queries`: every pending query (`chain_id`, `callback_id`), for the two purges.
- `autopilot.params`: `stakeibc_active`.
- `icahost.params` (`/ibc/apps/interchain_accounts/host/v1/params`): the allow-list.
- `wasm.params` (`code_upload_access`) and `wasm.contracts[].contract_info` (`admin`) for
  every instantiated contract, plus the `wasm.codes` entries those contracts reference so
  the suite can re-store them (the four deploy-key-admin contracts are the assertion target).
- `icaoracle.oracles`: the three active oracles.
- `ratelimit.rate_limits`, `ratelimit.blacklisted_denoms`, `ratelimit.whitelisted_address_pairs`.
- `icacallbacks_active_channel` (not a real export field, as in v34): a map from in-scope
  chain id to its delegation ICA's open active channel id, so the suite can register the
  channels the way mainnet has them and assert the reset on the ones with no commitment.

Record the height and date each section was taken at in the "Committed fixture provenance"
section, as v34 does. The haqq table is expected to be regenerated at the same height.
