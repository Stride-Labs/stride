# Wind-Down PR 6: Release Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the five wind-down PRs into a releasable v35: fill and prove the two operator address constants, replay the finished upgrade handler against a real mainnet export with the real constants, ship the §10 coverage-check script, and write the changelog and the pre-proposal checklist.

**Architecture:** No new protocol logic. This PR only (a) sets two `var`s that PR 4 left empty and hardens their tests, (b) adds a mainnet-export suite in `app/upgrades/v35` in the shape of v34's (skips when the fixture is absent, seeds every section the handler touches, asserts every §5 effect on real state), with the fixture assembled from the public REST API and a `verify_constants.py` staleness gate beside it, (c) adds `scripts/wind-down/coverage_check.py`, the off-chain assertion the design relies on, and (d) records the release in `CHANGELOG.md` and maps the spec's §9 pre-proposal checklist to concrete commands.

**Tech Stack:** Go 1.25 / cosmos-sdk v0.54.3 / ibc-go v11.2.0 / wasmd v0.70.2 (test app, `apptesting.AppTestHelper`, testify suites); Python 3 stdlib (`urllib`, `json`, `gzip`, `decimal`, `unittest`) for the scripts; `curl` + `jq` for the fixture.

> **Branching:** PR 1 branches off `wind-down-design-consolidation`. Each later PR branches
> off the previous PR's branch (PR 2 off PR 1, PR 3 off PR 2, and so on) and the PRs are
> implemented and merged strictly in order: 1, 2, 3, 4, 5, 6. Branch names:
> `wind-down-pr1-remove-handlers`, `wind-down-pr2-freeze-by-code`,
> `wind-down-pr3-upgrade-handler`, `wind-down-pr4-admin-txs`, `wind-down-pr5-sweep-tx`,
> `wind-down-pr6-release-gate`.
> The Go module path stays `github.com/Stride-Labs/stride/v34` in every PR; the bump to
> `/v35` is a manual step after all six land and is out of scope for every plan.

This plan is PR 6: branch `wind-down-pr6-release-gate` off `wind-down-pr5-sweep-tx`
(`git checkout wind-down-pr5-sweep-tx && git checkout -b wind-down-pr6-release-gate`). It
assumes PRs 1 to 5 are merged in order: the v35 package and every handler helper of the PR 3
plan exist with the names used below, `x/stakeibc/types/wind_down.go` and
`wind_down_test.go` exist as PR 4 wrote them, and the sweep tx and its batch builder exist as
PR 5 wrote them.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md`, sections §4, §9, §10, §11, §12 item 6, §13. Every value below is copied from there or from the PR 3, 4 and 5 plans.
- The two address constants live in `x/stakeibc/types/wind_down.go` as `var SweepOperatorAddress` and `var OsmosisVaultAddress`. Their values are **whatever spec §4 holds at implementation time**. On 2026-09-29 §4 holds sweep operator `stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh` and Osmosis vault `osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af`. The vault is the protocol-admin multisig re-encoded with the `osmo` prefix, on purpose (same signer set on both chains); which key signs the sweep is an ops choice recorded in §4, not something a test enforces. The test asserts only the bech32 prefix and the 20-byte length of each; the double-check against §4 and the signed-spend proof (Task 1 Step 2) are what guard the values.
- §9 rule: every constant is proven to be an address we control before anything is sent there: a test transfer to it from any wallet, then a signed spend from it.
- Handler helper names and order, from the PR 3 plan (Task 10 Step 1): `DisableAutopilotStakeibc`, `RemoveStakeibcFromICAHostAllowList`, `SetWasmUploadAccessToGov` (the only step that may error), `MoveDeployKeyContractAdminsToGov`, `DeprecateComdex`, `DeleteDydxTradeRoute`, `DeactivateICAOracles`, `RemoveAllRateLimits`, `ResetStaleDelegationChangesInProgress`, `PurgeHaqqSlashQueries`, `PurgeWithdrawalBalanceQueries`, `ReconcileHaqqDelegations`. Constants: `v35.UpgradeName = "v35"`, `v35.HaqqChainId`, `v35.ComdexChainId`, `v35.DydxTradeRouteRewardDenom = "uusdc"`, `v35.DydxTradeRouteHostDenom = "adydx"`, `v35.WasmDeployKey = "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh"`, `v35.HaqqDelegationDeltas []v35.DelegationDelta{Name, Address, Delta}`, `v35.GovModuleAddress() sdk.AccAddress`.
- Export suite conventions (v34's `app/upgrades/v34/mainnet_export_test.go`): fixture at `testdata/mainnet_export.json.gz`, `strideExport{AppState map[string]json.RawMessage}`, `s.App.AppCodec().UnmarshalJSON(raw, &genesis)` per section, one `populate*FromExport` helper per section, `s.ConfirmUpgradeSucceeded(v35.UpgradeName)` as the act, skip when the fixture is absent.
- §10 coverage check, per stToken: native tokens held on Osmosis for that denom ≥ Stride bank supply of the stToken × `HostZone.RedemptionRate`; per route pool exactly that channel's escrow balance × the rate; the canonical pool the remainder.
- Mainnet REST `https://stride-api.polkachu.com` and RPC `https://stride-rpc.polkachu.com` need a `User-Agent` header with curl (`-H 'User-Agent: curl/8.0'`); pin every fixture request to one height with `-H 'x-cosmos-block-height: <H>'`.
- Test commands: `go test ./app/upgrades/v35/... -run 'TestMainnetExportTestSuite' -v`, `go test ./x/stakeibc/types/... -run 'TestOperatorAddresses' -v`, `cd scripts/wind-down && python3 -m unittest test_coverage_check -v` (the test imports `coverage_check` as a module, so it runs from its own directory). The full suite has one pre-existing failure, `utils` `TestCreateModuleAccount`, which fails on `main` too and is not this PR's.
- Python: module-qualified imports (`import x.y` then `x.y.f()`), classes by name, every parameter and return typed, `X | None`, `Decimal` for token arithmetic, dataclasses for multi-field results, guard clauses, blank-line paragraphs with why-comments, no `getattr`/`hasattr`, named parameters on multi-argument calls.
- Commit messages end with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`. Nothing is pushed by this plan.

---

## File structure

| Path | Responsibility |
|---|---|
| `x/stakeibc/types/wind_down.go` (modify) | The two address vars get their values. Nothing else changes. |
| `x/stakeibc/types/wind_down_test.go` (modify) | `TestOperatorAddressesParse` becomes hard assertions (filled, right prefix, 20 bytes). |
| `app/upgrades/v35/testdata/mainnet_export.json.gz` (create) | Trimmed mainnet state, assembled per the README, committed during release prep only. |
| `app/upgrades/v35/testdata/README.md` (rewrite) | The exact assembly recipe and provenance (PR 3 left a section list; this PR fills in the commands and the height). |
| `app/upgrades/v35/testdata/verify_constants.py` (create) | Staleness gate: haqq delta table and the two channel maps against live chain state. |
| `app/upgrades/v35/mainnet_export_test.go` (create) | The export suite: seeds every section, runs the handler with the real constants, asserts every §5 effect. |
| `scripts/wind-down/coverage_check.py` (create) | The §10 check over an export and the vault's Osmosis balances. |
| `scripts/wind-down/test_coverage_check.py` (create) | Unit test on a synthetic export with an injected liquidity fetcher. |
| `CHANGELOG.md` (modify) | The v35 entry covering all six PRs. |

---

## Task 1: Fill and prove the two address constants

**Files:**
- Modify: `x/stakeibc/types/wind_down.go` (the `var` block, lines ~8-13 as PR 4 wrote it)
- Modify: `x/stakeibc/types/wind_down_test.go` (`TestOperatorAddressesParse`)

**Interfaces:**
- Consumes: `types.SweepOperatorAddress`, `types.OsmosisVaultAddress` (PR 4), `types.OsmosisBech32Prefix`.
- Produces: the filled constants every later task and the mainnet-export suite read.
- Review: yes (these two strings are where every ICA transfer lands and who may move user balances)

- [ ] **Step 1: Take the two addresses from spec §4**

Read spec §4's table and take the `Address` column of the "Sweep operator" and "Osmosis vault" rows. On 2026-09-29 they read:

| Role | §4 value today |
|---|---|
| Sweep operator | `stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh` |
| Osmosis vault | `osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af` |

The vault is the protocol-admin multisig with the `osmo` prefix (`strided keys parse` on both strings prints the same 20 bytes); that is the intended setup, the same signer set custodies the pools and signs the admin txs. Use exactly what §4 holds on the day; if §4 has changed since, the new strings win and nothing else in this task changes.

- [ ] **Step 2: Double-check both constants against §4 and prove each with a signed spend before this PR merges (ops, §9)**

This is the guard on the two strings: a constant is only committed once (a) it has been compared character by character with §4 by a second person, and (b) the address has received a test transfer and signed a spend. Run, and paste the four tx hashes into the commit message of Step 5:

```bash
# Sweep operator, on Stride (any funded wallet first, then the operator key itself)
strided tx bank send <any-wallet> stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh 1000000ustrd --chain-id stride-1 --node https://stride-rpc.polkachu.com:443 --gas auto --gas-adjustment 1.3 --fees 5000ustrd -y
strided tx bank send stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh <any-wallet>   500000ustrd --chain-id stride-1 --node https://stride-rpc.polkachu.com:443 --gas auto --gas-adjustment 1.3 --fees 5000ustrd -y --from sweep-operator

# Osmosis vault, on Osmosis (a multisig spends via `osmosisd tx multisign`; the spend proves the signer set)
osmosisd tx bank send <any-wallet> osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af 1000000uosmo --chain-id osmosis-1 --node https://osmosis-rpc.polkachu.com:443 --gas auto --gas-adjustment 1.3 --fees 2500uosmo -y
osmosisd tx bank send osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af <any-wallet>   500000uosmo --chain-id osmosis-1 --node https://osmosis-rpc.polkachu.com:443 --gas auto --gas-adjustment 1.3 --fees 2500uosmo -y --from vault
```

Expected: all four land (`code: 0`), and `strided q bank balances stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh` / `osmosisd q bank balances osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af` show the remainder. The operator keeps STRD for sweep fees only (§4). `verify_constants.py` (Task 2) later re-checks the automated half of this: both accounts must show `sequence > 0` on chain.

- [ ] **Step 3: Rewrite the address test as hard assertions**

Replace `TestOperatorAddressesParse` in `x/stakeibc/types/wind_down_test.go` (it currently logs and passes while the vars are empty) with:

```go
// The release gate fills the two operator addresses; from here on they must be set and must
// carry the right prefix and length. Which keys they are is decided in spec §4 and proven by
// the signed spends recorded in the commit that filled them, not asserted here.
func TestOperatorAddresses(t *testing.T) {
	require.NotEmpty(t, types.OsmosisVaultAddress, "OsmosisVaultAddress must be filled by the release gate")
	vaultBytes, err := sdk.GetFromBech32(types.OsmosisVaultAddress, types.OsmosisBech32Prefix)
	require.NoError(t, err, "osmosis vault must be an osmo bech32 address")
	require.Len(t, vaultBytes, 20, "osmosis vault must be a 20-byte account address")

	require.NotEmpty(t, types.SweepOperatorAddress, "SweepOperatorAddress must be filled by the release gate")
	operatorBytes, err := sdk.GetFromBech32(types.SweepOperatorAddress, "stride")
	require.NoError(t, err, "sweep operator must be a stride bech32 address")
	require.Len(t, operatorBytes, 20, "sweep operator must be a 20-byte account address")
}
```

Delete the old `TestOperatorAddressesParse`. No new imports are needed (`sdk`, `require` and `types` are already imported by PR 4's test file).

Run: `go test ./x/stakeibc/types/... -run 'TestOperatorAddresses' -v`
Expected: `--- FAIL: TestOperatorAddresses` with `OsmosisVaultAddress must be filled by the release gate` (the vars are still empty).

- [ ] **Step 4: Fill the constants**

In `x/stakeibc/types/wind_down.go` replace the `var` block with the §4 values (today's are shown; use what §4 holds on the day, per Step 1):

```go
// Wind-down operator addresses (spec §4), proven by a test transfer to each and a signed spend
// from each before this commit (tx hashes in the commit message). Vars rather than consts so
// tests can substitute them; every use fails closed if one is ever emptied again.
var (
	SweepOperatorAddress = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh" // the only signer of MsgSweepTokensOffStride
	OsmosisVaultAddress  = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"  // receiver of every MsgTransferFromIca; pool admin and moderator
)
```

Run: `go test ./x/stakeibc/types/... -run 'TestOperatorAddresses|TestHostToOsmosisTransferChannel|TestSweepUnwindChannels' -v`
Expected: `--- PASS` for all three.

Then confirm the fail-closed paths from PRs 4 and 5 still pass with the vars filled (their tests set and restore the vars themselves):

Run: `go test ./x/stakeibc/... -run 'TestKeeperTestSuite/(TestTransferFromIca|TestSweepTokensOffStride)|TestMsgSweepTokensOffStride|TestMsgTransferFromIca' -v`
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add x/stakeibc/types/wind_down.go x/stakeibc/types/wind_down_test.go
git commit -m "feat(stakeibc): fill the wind-down operator addresses

Sweep operator <address>, proven by <tx hash in> / <tx hash out>.
Osmosis vault <address> (the protocol-admin multisig with the osmo prefix), proven by <tx hash in> / <tx hash out>.
Both compared against spec §4 by <second person>.

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Parallel-safe tasks

Tasks 2, 3, 4 and 5 depend only on Task 1 (Task 3 depends on Task 2's fixture *format*, which this plan fixes below, not on the fixture file; the suite skips until the file exists and is run against it in Task 6). They touch disjoint files.

---

## Task 2: The mainnet export fixture, its README and the staleness gate

**Files:**
- Create: `app/upgrades/v35/testdata/mainnet_export.json.gz`
- Rewrite: `app/upgrades/v35/testdata/README.md` (PR 3 wrote the section list; keep its content and add the recipe and provenance below)
- Create: `app/upgrades/v35/testdata/verify_constants.py`

**Interfaces:**
- Consumes: `v35.HaqqDelegationDeltas` (parsed from `app/upgrades/v35/haqq.go` by regex), `types.HostToOsmosisTransferChannel` and `types.SweepUnwindChannels` (parsed from `x/stakeibc/types/wind_down.go` by regex).
- Produces: the fixture shape Task 3 reads: `app_state.stakeibc.{host_zone_list,trade_routes}`, `app_state.interchainquery.queries`, `app_state.autopilot.params`, `app_state.interchainaccounts.host_genesis_state.params`, `app_state.wasm.{params,contracts}`, `app_state.icaoracle.oracles`, `app_state.ratelimit.{rate_limits,blacklisted_denoms,whitelisted_address_pairs}`, `app_state.records.epoch_unbonding_record_list`, and the synthetic `app_state.delegation_channels` (chain id → `{connection_id, channel_id, packet_commitments}`).
- Depends on: Task 1
- Review: no (data assembly and a read-only script; the suite in Task 3 is what checks it)

- [ ] **Step 1: Assemble the fixture at one height**

Pick a recent height `H` from `curl -s -H 'User-Agent: curl/8.0' https://stride-rpc.polkachu.com/status | jq -r .result.sync_info.latest_block_height` and run:

```bash
API=https://stride-api.polkachu.com
UA='User-Agent: curl/8.0'
H="x-cosmos-block-height: <H>"
mkdir -p /tmp/v35-fixture && cd /tmp/v35-fixture

curl -s -H "$UA" -H "$H" "$API/Stride-Labs/stride/stakeibc/host_zone"                                     > host_zones.json
curl -s -H "$UA" -H "$H" "$API/Stride-Labs/stride/stakeibc/trade_routes"                                  > trade_routes.json
curl -s -H "$UA" -H "$H" "$API/Stride-Labs/stride/interchainquery/pending_queries"                        > queries.json
curl -s -H "$UA" -H "$H" "$API/Stride-Labs/stride/autopilot/params"                                       > autopilot.json
curl -s -H "$UA" -H "$H" "$API/ibc/apps/interchain_accounts/host/v1/params"                               > icahost.json
curl -s -H "$UA" -H "$H" "$API/cosmwasm/wasm/v1/codes/params"                                             > wasm_params.json
curl -s -H "$UA" -H "$H" "$API/Stride-Labs/stride/icaoracle/oracles"                                      > oracles.json
curl -s -H "$UA" -H "$H" "$API/ibc/apps/rate-limiting/v1/ratelimits"                                      > rate_limits.json
curl -s -H "$UA" -H "$H" "$API/ibc/apps/rate-limiting/v1/ratelimit/blacklisted_denoms"                    > blacklist.json
curl -s -H "$UA" -H "$H" "$API/ibc/apps/rate-limiting/v1/ratelimit/whitelisted_addresses"                 > whitelist.json
curl -s -H "$UA" -H "$H" "$API/Stride-Labs/stride/records/epoch_unbonding_record?pagination.limit=2000"   > epoch_unbonding.json

# The four contracts whose admin is the deploy key (spec §3); re-list them right before the
# proposal, the set is small enough to check by hand
for c in $(curl -s -H "$UA" -H "$H" "$API/cosmwasm/wasm/v1/codes?pagination.limit=200" | jq -r '.code_infos[].code_id'); do
  curl -s -H "$UA" -H "$H" "$API/cosmwasm/wasm/v1/code/$c/contracts?pagination.limit=200" | jq -r '.contracts[]'
done | sort -u > all_contracts.txt
> contracts.json.tmp
while read -r addr; do
  curl -s -H "$UA" -H "$H" "$API/cosmwasm/wasm/v1/contract/$addr" | jq -c '{contract_address: .address, contract_info: .contract_info}' >> contracts.json.tmp
  sleep 1
done < all_contracts.txt
jq -s '[.[] | select(.contract_info.admin == "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh")]' contracts.json.tmp > deploy_key_contracts.json
jq length deploy_key_contracts.json   # expected 4 (spec §3); if not, note the new count in the README

# Delegation ICA channels per in-scope zone: the open active channel and whether it has packet commitments
> delegation_channels.json.tmp
for chain in celestia cosmoshub-4 dydx-mainnet-1 haqq_11235-1 injective-1 juno-1 laozi-mainnet osmosis-1 phoenix-1 sommelier-3 ssc-1; do
  port="icacontroller-$chain.DELEGATION"
  conn=$(jq -r --arg c "$chain" '.host_zone[] | select(.chain_id == $c) | .connection_id' host_zones.json)
  # Channels are listed per connection (the global list paginates past what one call returns)
  chan=$(curl -s -H "$UA" -H "$H" "$API/ibc/core/connection/v1/connections/$conn/channels?pagination.limit=1000" \
         | jq -r --arg p "$port" '[.channels[] | select(.port_id == $p and .state == "STATE_OPEN")] | last | .channel_id // empty')
  if [ -z "$chan" ]; then
    jq -n --arg c "$chain" --arg conn "$conn" '{($c): {connection_id: $conn, channel_id: "", packet_commitments: 0}}' >> delegation_channels.json.tmp
    continue
  fi
  n=$(curl -s -H "$UA" -H "$H" "$API/ibc/core/channel/v1/channels/$chan/ports/$port/packet_commitments" | jq '.commitments | length')
  jq -n --arg c "$chain" --arg conn "$conn" --arg chan "$chan" --argjson n "$n" \
     '{($c): {connection_id: $conn, channel_id: $chan, packet_commitments: $n}}' >> delegation_channels.json.tmp
done
jq -s 'add' delegation_channels.json.tmp > delegation_channels.json

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
gzip -c trimmed.json > <repo>/app/upgrades/v35/testdata/mainnet_export.json.gz
ls -la <repo>/app/upgrades/v35/testdata/mainnet_export.json.gz   # expected well under 1 MB
```

Expected: every `curl` returns JSON (an HTML page means the `User-Agent` header was dropped; a `429` means slow down, the loops already sleep). If a REST field name above differs from the module's genesis JSON name (the suite decodes each section with the app codec), fix the `jq` key, not the Go: the names used are the proto field names of each module's `GenesisState`.

- [ ] **Step 2: Write the README**

Rewrite `app/upgrades/v35/testdata/README.md` (keep PR 3's section list as the "Sections the suite consumes" part) so it has these parts, in the v34 README's shape:

```markdown
# v35 mainnet export fixture

`mainnet_export.json.gz` is mainnet state trimmed to the sections the v35 mainnet-export
suite consumes. The suite (`../mainnet_export_test.go`) replays the whole v35 handler
against it with the real constants and asserts every effect of spec §5; it skips when the
file is absent so CI stays green before release prep.

## Sections the suite consumes (all under `app_state`)

This list supersedes the one PR 3 wrote in this file: two of PR 3's entries are renamed to
the genesis proto names the suite decodes with (`icahost.params` is now
`interchainaccounts.host_genesis_state.params`, `icacallbacks_active_channel` is now
`delegation_channels`, which also carries the packet-commitment count), and `wasm.codes` is
dropped because the suite instantiates a stand-in code per contract instead of re-storing
Hyperlane's.

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
- `delegation_channels` (not a real export field): chain id → `{connection_id, channel_id,
  packet_commitments}` for every in-scope zone's delegation ICA, so the suite can register
  each open active channel the way mainnet has it and assert the stale-flag reset on exactly
  the zones with zero commitments.

## Assembly

<the Step 1 commands, verbatim>

## Committed fixture provenance

Assembled <date> at mainnet height **<H>** from `stride-api.polkachu.com`, every section
pinned to that height with `x-cosmos-block-height`. Deploy-key contracts found: <n>. Zones
with an open delegation channel and zero packet commitments at that height: <list>. The haqq
delta table in `../haqq.go` was regenerated at the same height
(`scripts/wind-down/gen_delta_table.py`, see the PR 3 plan Task 9).

## Staleness gate

`python3 app/upgrades/v35/testdata/verify_constants.py` recomputes the haqq delta table from
the live chain and checks both channel maps against the hosts' and Stride's IBC state. Run
it right before cutting the release; it exits non-zero on any drift.

## Regenerating from a node export

From a synced node: `strided export --height <H> > full_export.json`, then a `jq` that keeps
`stakeibc.{host_zone_list,trade_routes}`, `interchainquery.queries`, `autopilot.params`,
`interchainaccounts.host_genesis_state.params`, `wasm.params` and the deploy-key entries of
`wasm.contracts`, `icaoracle.oracles`, `ratelimit.{rate_limits,blacklisted_denoms,
whitelisted_address_pairs}` and `records.epoch_unbonding_record_list`; `delegation_channels`
is still built from the IBC queries above, since an export does not say which channel is
active.
```

- [ ] **Step 3: Write `verify_constants.py`**

```python
#!/usr/bin/env python3
"""Staleness gate for the v35 wind-down constants.

Three things in the binary are measurements of chain state that can drift between the PR and
the proposal: the haqq delegation delta table (app/upgrades/v35/haqq.go), the host-side
channel map to Osmosis (x/stakeibc/types/wind_down.go, HostToOsmosisTransferChannel) and the
sweep unwind whitelist (SweepUnwindChannels). This script recomputes each from live REST
state and fails loudly on any difference, so a stale constant is a red pre-proposal step and
not a silently skipped reconciliation or a transfer to the wrong chain. It also checks that
the two operator address constants have signed at least one transaction on their chain
(account sequence > 0), the automated half of the §9 signed-spend proof.

Stdlib-only. Reads the Go files relative to its own location.

Usage: python3 app/upgrades/v35/testdata/verify_constants.py
"""

import json
import pathlib
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from decimal import Decimal

USER_AGENT = "curl/8.0"
TIMEOUT_SECONDS = 20
STRIDE_API = "https://stride-api.polkachu.com"
OSMOSIS_API = "https://osmosis-api.polkachu.com"
CHAIN_REGISTRY = "https://raw.githubusercontent.com/cosmos/chain-registry/master"
OSMOSIS_CHAIN_ID = "osmosis-1"

SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
HAQQ_GO = SCRIPT_DIR.parent / "haqq.go"
WIND_DOWN_GO = SCRIPT_DIR.parent.parent.parent.parent / "x" / "stakeibc" / "types" / "wind_down.go"

HAQQ_CHAIN_ID = "haqq_11235-1"
HAQQ_APIS = ["https://haqq-rest.publicnode.com", "https://rest.cosmos.directory/haqq"]

# Stride chain id -> chain-registry directory name, for the hosts' REST endpoints
REGISTRY_NAMES = {
    "celestia": "celestia",
    "cosmoshub-4": "cosmoshub",
    "dydx-mainnet-1": "dydx",
    "haqq_11235-1": "haqq",
    "injective-1": "injective",
    "juno-1": "juno",
    "laozi-mainnet": "bandchain",
    "phoenix-1": "terra2",
    "sommelier-3": "sommelier",
    "ssc-1": "saga",
}

# The counterparty chain each whitelisted Stride channel must lead to (spec §7)
SWEEP_CHANNEL_CHAIN_IDS = {
    "channel-0": "cosmoshub-4",
    "channel-162": "celestia",
    "channel-5": "osmosis-1",
    "channel-24": "juno-1",
    "channel-150": "sommelier-3",
    "channel-213": "ssc-1",
    "channel-160": "dydx-mainnet-1",
}

failures: list[str] = []


@dataclass(frozen=True)
class DeltaEntry:
    name: str
    address: str
    delta: int


def report(name: str, ok: bool, detail: str = "") -> None:
    status = "PASS" if ok else "FAIL"
    suffix = f" -- {detail}" if detail else ""
    print(f"[{status}] {name}{suffix}")
    if not ok:
        failures.append(name)


def fetch_json(url: str) -> dict:
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(request, timeout=TIMEOUT_SECONDS) as response:
        return json.load(response)


def fetch_from_any(bases: list[str], path: str) -> dict:
    """Public LCDs for third-party chains come and go; the first one that answers wins."""
    errors = []
    for base in bases:
        try:
            return fetch_json(f"{base}{path}")
        except (urllib.error.URLError, TimeoutError, OSError, json.JSONDecodeError) as exc:
            errors.append(f"{base}: {exc}")
    sys.exit(f"no endpoint answered {path}: {errors}")


def registry_rest_endpoints(registry_name: str) -> list[str]:
    chain = fetch_json(f"{CHAIN_REGISTRY}/{registry_name}/chain.json")
    return [entry["address"].rstrip("/") for entry in chain.get("apis", {}).get("rest", [])]


# ----------------------------------------------------------------------------------------------
# Parsing the Go constants
# ----------------------------------------------------------------------------------------------

DELTA_LINE = re.compile(r'\{Name:\s*"([^"]+)",\s*Address:\s*"([^"]+)",\s*Delta:\s*mustInt\("(-?\d+)"\)\}')
MAP_ENTRY = re.compile(r'"([^"]+)":\s*"([^"]*)",')


def parse_haqq_table() -> list[DeltaEntry]:
    source = HAQQ_GO.read_text()
    entries = [DeltaEntry(name=n, address=a, delta=int(d)) for n, a, d in DELTA_LINE.findall(source)]
    if not entries:
        sys.exit(f"no delta entries parsed from {HAQQ_GO}")
    return entries


def parse_go_map(var_name: str) -> dict[str, str]:
    source = WIND_DOWN_GO.read_text()
    match = re.search(rf"var {var_name} = map\[string\]string\{{(.*?)\n\}}", source, re.S)
    if match is None:
        sys.exit(f"{var_name} not found in {WIND_DOWN_GO}")
    return dict(MAP_ENTRY.findall(match.group(1)))


def parse_go_string_var(var_name: str) -> str:
    source = WIND_DOWN_GO.read_text()
    match = re.search(rf'{var_name}\s*=\s*"([^"]*)"', source)
    if match is None:
        sys.exit(f"{var_name} not found in {WIND_DOWN_GO}")
    return match.group(1)


# ----------------------------------------------------------------------------------------------
# Checks
# ----------------------------------------------------------------------------------------------

def live_haqq_deltas() -> dict[str, int]:
    """On-chain delegation minus tracked delegation per validator, the same sign as the table."""
    host_zone = fetch_json(f"{STRIDE_API}/Stride-Labs/stride/stakeibc/host_zone/{HAQQ_CHAIN_ID}")["host_zone"]
    tracked = {v["address"]: int(v["delegation"]) for v in host_zone["validators"]}

    on_chain: dict[str, int] = {}
    next_key = ""
    while True:
        path = f"/cosmos/staking/v1beta1/delegations/{host_zone['delegation_ica_address']}?pagination.limit=200"
        if next_key:
            path += f"&pagination.key={urllib.parse.quote(next_key)}"
        page = fetch_from_any(HAQQ_APIS, path)
        for entry in page["delegation_responses"]:
            on_chain[entry["delegation"]["validator_address"]] = int(entry["balance"]["amount"])
        next_key = page.get("pagination", {}).get("next_key") or ""
        if not next_key:
            break

    return {address: on_chain.get(address, 0) - tracked_amount for address, tracked_amount in tracked.items()}


def check_haqq_table() -> None:
    table = {entry.address: entry for entry in parse_haqq_table()}
    live = live_haqq_deltas()

    # Every non-zero live delta must be in the table with the same value, and the table must
    # not carry an entry the chain no longer shows: either way the handler would skip the whole
    # table (validateDelegationDeltas) or true up to the wrong number
    for address, delta in live.items():
        entry = table.get(address)
        if delta == 0 and entry is None:
            continue
        expected = entry.delta if entry is not None else 0
        name = entry.name if entry is not None else address
        report(f"haqq delta {name}", delta == expected, f"live {delta} vs table {expected}")
    for address, entry in table.items():
        report(f"haqq validator {entry.name} tracked", address in live, "validator missing from the Stride host zone")

    net = sum(entry.delta for entry in table.values())
    print(f"       table net delta: {net} aISLM ({Decimal(net) / Decimal(10**18):.6f} ISLM)")


def channel_counterparty_chain_id(rest: str, channel_id: str) -> tuple[str, str]:
    """Returns (channel state, counterparty chain id) for a transfer channel on the given chain."""
    channel = fetch_json(f"{rest}/ibc/core/channel/v1/channels/{channel_id}/ports/transfer")["channel"]
    client_state = fetch_json(f"{rest}/ibc/core/channel/v1/channels/{channel_id}/ports/transfer/client_state")
    return channel["state"], client_state["identified_client_state"]["client_state"]["chain_id"]


def check_host_to_osmosis_map() -> None:
    channel_map = parse_go_map("HostToOsmosisTransferChannel")
    for chain_id, channel_id in sorted(channel_map.items()):
        if chain_id == OSMOSIS_CHAIN_ID:
            report(f"{chain_id} maps to the bank-send form", channel_id == "")
            continue
        endpoints = registry_rest_endpoints(REGISTRY_NAMES[chain_id])
        for rest in endpoints:
            try:
                state, counterparty = channel_counterparty_chain_id(rest=rest, channel_id=channel_id)
            except (urllib.error.URLError, TimeoutError, OSError, KeyError, json.JSONDecodeError):
                continue
            report(f"{chain_id} {channel_id} -> osmosis-1 and open",
                   state == "STATE_OPEN" and counterparty == OSMOSIS_CHAIN_ID,
                   f"{state}, counterparty {counterparty} via {rest}")
            break
        else:
            report(f"{chain_id} {channel_id} reachable", False, "no registry REST endpoint answered")


def check_sweep_unwind_channels() -> None:
    whitelist = parse_go_map("SweepUnwindChannels")
    report("sweep whitelist has exactly the seven §7 channels", set(whitelist) == set(SWEEP_CHANNEL_CHAIN_IDS))
    for channel_id, expected_chain in sorted(SWEEP_CHANNEL_CHAIN_IDS.items()):
        state, counterparty = channel_counterparty_chain_id(rest=STRIDE_API, channel_id=channel_id)
        report(f"stride {channel_id} -> {expected_chain} and open",
               state == "STATE_OPEN" and counterparty == expected_chain, f"{state}, counterparty {counterparty}")


def account_sequence(rest: str, address: str) -> int:
    """The signing sequence of an account; a multisig is a BaseAccount too, so one shape covers both."""
    account = fetch_json(f"{rest}/cosmos/auth/v1beta1/accounts/{address}")["account"]
    base = account.get("base_account", account)
    return int(base["sequence"])


def check_operator_addresses_have_signed() -> None:
    """Both constants must have signed at least once on their chain: the automated half of the §9
    proof that each address is one we control (the other half is the tx hashes in the commit)."""
    sweep_operator = parse_go_string_var("SweepOperatorAddress")
    vault = parse_go_string_var("OsmosisVaultAddress")
    report("SweepOperatorAddress is filled", bool(sweep_operator))
    report("OsmosisVaultAddress is filled", bool(vault))
    if not sweep_operator or not vault:
        return
    for name, rest, address in (("sweep operator", STRIDE_API, sweep_operator), ("osmosis vault", OSMOSIS_API, vault)):
        try:
            sequence = account_sequence(rest=rest, address=address)
        except (urllib.error.URLError, TimeoutError, OSError, KeyError, json.JSONDecodeError) as exc:
            report(f"{name} {address} has signed", False, f"account query failed: {exc}")
            continue
        report(f"{name} {address} has signed", sequence > 0, f"sequence {sequence}")


def main() -> int:
    print("== operator addresses ==")
    check_operator_addresses_have_signed()
    print("== haqq delegation delta table ==")
    check_haqq_table()
    print("== host-side channels to Osmosis ==")
    check_host_to_osmosis_map()
    print("== sweep unwind whitelist ==")
    check_sweep_unwind_channels()

    if failures:
        print(f"\n{len(failures)} check(s) FAILED: {failures}")
        return 1
    print("\nall checks passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
```

Run: `python3 app/upgrades/v35/testdata/verify_constants.py`
Expected: every line `[PASS]`, `all checks passed`, exit 0. A `[FAIL] ... has signed` line means an address constant has never signed on its chain: Task 1 Step 2 was skipped or the wrong string was committed. A `[FAIL] haqq delta ...` line means the table must be regenerated (PR 3 plan, Task 9 Step 7) at the fixture's height and the fixture re-assembled; a channel `[FAIL]` means a host has closed or replaced its channel to Osmosis and the map must be corrected in `wind_down.go` (with PR 4's `TestHostToOsmosisTransferChannel` rerun).

- [ ] **Step 4: Commit**

```bash
git add app/upgrades/v35/testdata/mainnet_export.json.gz app/upgrades/v35/testdata/README.md app/upgrades/v35/testdata/verify_constants.py
git commit -m "test(v35): mainnet export fixture at height <H>, assembly README and the constants staleness gate

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Task 3: The mainnet-export suite

**Files:**
- Create: `app/upgrades/v35/mainnet_export_test.go`

**Interfaces:**
- Consumes: the fixture format of Task 2; the v35 helpers and constants named in Global Constraints; the test-app keepers `s.App.StakeibcKeeper`, `s.App.InterchainqueryKeeper`, `s.App.AutopilotKeeper`, `s.App.ICAHostKeeper`, `s.App.WasmKeeper`, `s.App.ICAOracleKeeper`, `s.App.RatelimitKeeper`, `s.App.RecordsKeeper`, `s.App.IBCKeeper.ChannelKeeper`, `s.App.ICAControllerKeeper`; `apptesting` helpers `MockICAChannel(connectionId, channelId, owner, address)`, `ConfirmUpgradeSucceeded(name)`; the PR 3 test helper `s.storeAndInstantiateHackatom(admin)` is in `UpgradeTestSuite`, so this suite carries its own single-code store step (below).
- Depends on: Task 1 (the fixture file itself is only needed in Task 6; the suite skips without it)
- Review: yes (this is the release gate: it is the only place the real constants meet real state)

- [ ] **Step 1: Write the suite skeleton and the section loaders**

```go
package v35_test

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
	"testing"

	icagenesistypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/genesis/types"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	ratelimittypes "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/types"
	"github.com/stretchr/testify/suite"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/testdata"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	autopilottypes "github.com/Stride-Labs/stride/v34/x/autopilot/types"
	icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// mainnetExportPath is relative to this package: read from the testdata checkout, never
// shipped in the binary.
const mainnetExportPath = "testdata/mainnet_export.json.gz"

// MainnetExportTestSuite replays the v35 handler against real mainnet state with the REAL
// constants (no test-key substitution) and asserts every effect of spec §5 on it. It is the
// release gate: a constant that no longer matches state (the haqq delta table, the contract
// list, the channel picture) turns into a red build here instead of a silently skipped step.
// The fixture is committed during release prep only; the suite skips while it is absent.
type MainnetExportTestSuite struct {
	apptesting.AppTestHelper
}

func (s *MainnetExportTestSuite) SetupTest() {
	s.Setup()
}

func TestMainnetExportTestSuite(t *testing.T) {
	if _, err := os.Stat(mainnetExportPath); errors.Is(err, os.ErrNotExist) {
		t.Skipf("skipping: mainnet export fixture not present at %s — see testdata/README.md to generate it", mainnetExportPath)
	}
	suite.Run(t, new(MainnetExportTestSuite))
}

// strideExport is a thin view over the trimmed export JSON shape.
type strideExport struct {
	AppState map[string]json.RawMessage `json:"app_state"`
}

// delegationChannel is the fixture's synthetic per-zone delegation ICA picture (README).
type delegationChannel struct {
	ConnectionId      string `json:"connection_id"`
	ChannelId         string `json:"channel_id"`
	PacketCommitments int    `json:"packet_commitments"`
}

func (s *MainnetExportTestSuite) loadTrimmedExport() strideExport {
	f, err := os.Open(mainnetExportPath)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = f.Close() })

	gz, err := gzip.NewReader(f)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = gz.Close() })

	var export strideExport
	s.Require().NoError(json.NewDecoder(gz).Decode(&export))
	s.Require().NotEmpty(export.AppState, "trimmed export has no app_state — check testdata/README.md")
	return export
}

func (s *MainnetExportTestSuite) section(export strideExport, name string) json.RawMessage {
	raw, ok := export.AppState[name]
	s.Require().True(ok, "trimmed export missing %s section — regenerate per testdata/README.md", name)
	return raw
}

// populateStakeibcFromExport seeds every host zone and trade route and returns the zones by
// chain id for later comparison.
func (s *MainnetExportTestSuite) populateStakeibcFromExport(export strideExport) map[string]stakeibctypes.HostZone {
	var genesis stakeibctypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "stakeibc"), &genesis))
	s.Require().GreaterOrEqual(len(genesis.HostZoneList), 12, "export should carry the eleven in-scope zones and comdex-1 at least")

	hostZones := map[string]stakeibctypes.HostZone{}
	for _, hostZone := range genesis.HostZoneList {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
		hostZones[hostZone.ChainId] = hostZone
	}
	for _, tradeRoute := range genesis.TradeRoutes {
		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)
	}
	s.Require().Contains(hostZones, v35.HaqqChainId)
	s.Require().Contains(hostZones, v35.ComdexChainId)
	s.Require().Len(genesis.TradeRoutes, 1, "mainnet has exactly the dYdX trade route")
	return hostZones
}

func (s *MainnetExportTestSuite) populateQueriesFromExport(export strideExport) []icqtypes.Query {
	var genesis icqtypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "interchainquery"), &genesis))
	for _, query := range genesis.Queries {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
	}
	return genesis.Queries
}

func (s *MainnetExportTestSuite) populateAutopilotFromExport(export strideExport) {
	var genesis autopilottypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "autopilot"), &genesis))
	s.Require().True(genesis.Params.StakeibcActive, "autopilot stakeibc is active on mainnet before the upgrade")
	s.App.AutopilotKeeper.SetParams(s.Ctx, genesis.Params)
}

func (s *MainnetExportTestSuite) populateICAHostFromExport(export strideExport) []string {
	var genesis icagenesistypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "interchainaccounts"), &genesis))
	params := genesis.HostGenesisState.Params
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}))
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}))
	s.Require().Contains(params.AllowMessages, sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}))
	s.App.ICAHostKeeper.SetParams(s.Ctx, params)
	return params.AllowMessages
}

// populateWasmFromExport seeds the real upload-access params and one stand-in contract per
// deploy-key contract in the fixture. The fixture carries the contracts' infos but not their
// (Hyperlane) code, and wasmd's InitGenesis refuses a contract without a code history, so the
// suite stores hackatom once (through the gov permission keeper, which bypasses upload
// access) and instantiates it once per fixture entry with the deploy key as admin: the
// handler only reads and rewrites ContractInfo.Admin, so the code behind the address does
// not matter. What the fixture proves is the count and the admin of the real contracts;
// what the suite proves is that every contract with that admin ends with gov as admin.
func (s *MainnetExportTestSuite) populateWasmFromExport(export strideExport) []sdk.AccAddress {
	var genesis wasmtypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "wasm"), &genesis))
	s.Require().NotEmpty(genesis.Contracts, "export should carry the deploy-key contracts")
	s.Require().NoError(s.App.WasmKeeper.SetParams(s.Ctx, genesis.Params))

	govKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	creator := apptesting.CreateRandomAccounts(1)[0]
	codeId, _, err := govKeeper.Create(s.Ctx, creator, testdata.HackatomContractWasm(), nil)
	s.Require().NoError(err, "store hackatom")
	initMsg, err := json.Marshal(map[string]string{"verifier": creator.String(), "beneficiary": creator.String()})
	s.Require().NoError(err)

	deployKey := sdk.MustAccAddressFromBech32(v35.WasmDeployKey)
	addresses := []sdk.AccAddress{}
	for i, contract := range genesis.Contracts {
		s.Require().Equal(v35.WasmDeployKey, contract.ContractInfo.Admin, "fixture contract %d (%s) is not deploy-key administered", i, contract.ContractAddress)
		address, _, err := govKeeper.Instantiate(s.Ctx, codeId, creator, deployKey, initMsg, "hackatom", sdk.NewCoins())
		s.Require().NoError(err, "instantiate stand-in for %s", contract.ContractAddress)
		addresses = append(addresses, address)
	}

	params := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, params.CodeUploadAccess.Permission)
	s.Require().Len(params.CodeUploadAccess.Addresses, 2, "spec §3: upload restricted to two addresses before the upgrade")
	return addresses
}

func (s *MainnetExportTestSuite) populateOraclesFromExport(export strideExport) []icaoracletypes.Oracle {
	var genesis icaoracletypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "icaoracle"), &genesis))
	for _, oracle := range genesis.Oracles {
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)
	}
	return genesis.Oracles
}

func (s *MainnetExportTestSuite) populateRateLimitsFromExport(export strideExport) ratelimittypes.GenesisState {
	var genesis ratelimittypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "ratelimit"), &genesis))
	s.Require().NotEmpty(genesis.RateLimits, "spec §3: stToken rate limits exist before the upgrade")
	for _, rateLimit := range genesis.RateLimits {
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit)
	}
	for _, denom := range genesis.BlacklistedDenoms {
		s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, denom)
	}
	for _, pair := range genesis.WhitelistedAddressPairs {
		s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, pair)
	}
	return genesis
}

func (s *MainnetExportTestSuite) populateRecordsFromExport(export strideExport) recordstypes.GenesisState {
	var genesis recordstypes.GenesisState
	s.Require().NoError(s.App.AppCodec().UnmarshalJSON(s.section(export, "records"), &genesis))
	for _, record := range genesis.EpochUnbondingRecordList {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, record)
	}
	return genesis
}

// populateDelegationChannelsFromExport registers each zone's open delegation channel the way
// mainnet has it and plants one packet commitment per recorded in-flight packet, so the
// stale-flag reset sees the real picture. Returns the chain ids whose flags must be reset.
func (s *MainnetExportTestSuite) populateDelegationChannelsFromExport(export strideExport, hostZones map[string]stakeibctypes.HostZone) (resettable []string) {
	var channels map[string]delegationChannel
	s.Require().NoError(json.Unmarshal(s.section(export, "delegation_channels"), &channels))

	for chainId, channel := range channels {
		if channel.ChannelId == "" {
			continue
		}
		owner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_DELEGATION)
		s.MockICAChannel(channel.ConnectionId, channel.ChannelId, owner, hostZones[chainId].DelegationIcaAddress)
		portId, _ := icatypes.NewControllerPortID(owner)
		for sequence := 1; sequence <= channel.PacketCommitments; sequence++ {
			s.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(s.Ctx, portId, channel.ChannelId, uint64(sequence), []byte{1})
		}
		if channel.PacketCommitments == 0 {
			resettable = append(resettable, chainId)
		}
	}
	s.Require().NotEmpty(resettable, "at least one zone should have an open channel with nothing in flight")
	return resettable
}
```

The hackatom bytes come from wasmd's testdata package (`testdata.HackatomContractWasm()`), the same call the PR 3 plan's `storeAndInstantiateHackatom` makes; no file path is involved. `os` stays imported for the fixture file.

- [ ] **Step 2: Write the test body**

```go
func (s *MainnetExportTestSuite) TestUpgradeFromMainnetExport() {
	// ----- arrange: seed every section the handler touches from real mainnet state -----
	export := s.loadTrimmedExport()
	hostZones := s.populateStakeibcFromExport(export)
	queriesBefore := s.populateQueriesFromExport(export)
	s.populateAutopilotFromExport(export)
	allowBefore := s.populateICAHostFromExport(export)
	deployKeyContracts := s.populateWasmFromExport(export)
	oraclesBefore := s.populateOraclesFromExport(export)
	s.populateRateLimitsFromExport(export)
	recordsBefore := s.populateRecordsFromExport(export)
	resettableZones := s.populateDelegationChannelsFromExport(export, hostZones)

	haqqBefore := hostZones[v35.HaqqChainId]
	flagsBefore := map[string][]int64{}
	for chainId, hostZone := range hostZones {
		flagsBefore[chainId] = delegationChangeFlags(hostZone)
	}

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)

	// ----- assert: entry points -----
	s.Require().False(s.App.AutopilotKeeper.GetParams(s.Ctx).StakeibcActive, "autopilot stakeibc off")
	allowAfter := s.App.ICAHostKeeper.GetParams(s.Ctx).AllowMessages
	s.Require().NotContains(allowAfter, sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}))
	s.Require().NotContains(allowAfter, sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}))
	s.Require().Contains(allowAfter, sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}))
	s.Require().Len(allowAfter, len(allowBefore)-2, "exactly the two stakeibc messages leave the allow-list")

	// ----- assert: wasm -----
	uploadAccess := s.App.WasmKeeper.GetParams(s.Ctx).CodeUploadAccess
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, uploadAccess.Permission)
	s.Require().Equal([]string{v35.GovModuleAddress().String()}, uploadAccess.Addresses, "upload access is gov only")
	for _, address := range deployKeyContracts {
		info := s.App.WasmKeeper.GetContractInfo(s.Ctx, address)
		s.Require().NotNil(info)
		s.Require().Equal(v35.GovModuleAddress().String(), info.Admin, "contract %s admin moved to gov", address)
	}

	// ----- assert: stakeibc state flips -----
	comdex, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(found)
	s.Require().True(comdex.Deprecated, "comdex-1 deprecated")
	s.Require().False(comdex.Halted, "comdex-1 Halted untouched")
	s.Require().Empty(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), "dYdX trade route deleted")

	// ----- assert: oracles and rate limits -----
	for _, before := range oraclesBefore {
		after, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, before.ChainId)
		s.Require().True(found)
		s.Require().False(after.Active, "oracle %s inactive", before.ChainId)
	}
	s.Require().Empty(s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx))
	s.Require().Empty(s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx))
	s.Require().Empty(s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx))

	// ----- assert: stale flags reset exactly where nothing is in flight -----
	for chainId, before := range flagsBefore {
		after, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().True(found)
		if contains(resettableZones, chainId) {
			for _, flag := range delegationChangeFlags(after) {
				s.Require().Zero(flag, "%s: every DelegationChangesInProgress reset", chainId)
			}
			continue
		}
		s.Require().Equal(before, delegationChangeFlags(after), "%s: flags untouched (channel missing or packets in flight)", chainId)
	}

	// ----- assert: ICQ purges -----
	remaining := map[string]icqtypes.Query{}
	for _, query := range s.App.InterchainqueryKeeper.AllQueries(s.Ctx) {
		remaining[query.Id] = query
	}
	for _, query := range queriesBefore {
		_, stillThere := remaining[query.Id]
		isHaqqSlashPath := query.ChainId == v35.HaqqChainId && (query.CallbackId == stakeibckeeper.ICQCallbackID_Delegation ||
			query.CallbackId == stakeibckeeper.ICQCallbackID_Validator || query.CallbackId == stakeibckeeper.ICQCallbackID_Calibrate)
		isWithdrawalBalance := query.CallbackId == stakeibckeeper.ICQCallbackID_WithdrawalHostBalance
		s.Require().Equal(!(isHaqqSlashPath || isWithdrawalBalance), stillThere,
			"query %s (%s %s) purge decision", query.Id, query.ChainId, query.CallbackId)
	}
	haqqAfter, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, validator := range haqqAfter.Validators {
		s.Require().False(validator.SlashQueryInProgress, "haqq %s slash flag cleared", validator.Name)
	}

	// ----- assert: haqq deltas applied in full with the real table -----
	expectedNet := sdkmath.ZeroInt()
	for _, entry := range v35.HaqqDelegationDeltas {
		before, _, found := stakeibckeeper.GetValidatorFromAddress(haqqBefore.Validators, entry.Address)
		s.Require().True(found, "table validator %s is on the mainnet haqq zone", entry.Name)
		after, _, _ := stakeibckeeper.GetValidatorFromAddress(haqqAfter.Validators, entry.Address)
		s.Require().Equal(before.Delegation.Add(entry.Delta), after.Delegation, "haqq %s delta applied", entry.Name)
		expectedNet = expectedNet.Add(entry.Delta)
	}
	s.Require().True(expectedNet.IsNegative(), "the haqq table nets to a decrease (spec §5)")
	s.Require().Equal(haqqBefore.TotalDelegations.Add(expectedNet), haqqAfter.TotalDelegations, "haqq TotalDelegations moved by the net delta")
	sum := sdkmath.ZeroInt()
	for _, validator := range haqqAfter.Validators {
		sum = sum.Add(validator.Delegation)
	}
	s.Require().Equal(sum, haqqAfter.TotalDelegations, "haqq TotalDelegations == sum of validators")

	// ----- assert: nothing the handler must not touch -----
	// Compared by String(): Equal on structs holding sdkmath.Int is a DeepEqual over big.Int
	// internals and can differ for equal values (v34's suite compares the same way)
	recordsAfter := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	s.Require().Len(recordsAfter, len(recordsBefore.EpochUnbondingRecordList), "unbonding record count untouched")
	for i, before := range recordsBefore.EpochUnbondingRecordList {
		s.Require().Equal(before.String(), recordsAfter[i].String(), "unbonding record %d untouched", before.EpochNumber)
	}
	for chainId, before := range hostZones {
		after, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().Equal(before.RedemptionRate, after.RedemptionRate, "%s redemption rate untouched", chainId)
		s.Require().Equal(before.Halted, after.Halted, "%s Halted untouched", chainId)
	}
}

// DelegationChangesInProgress is an int64 on the proto (validator.pb.go).
func delegationChangeFlags(hostZone stakeibctypes.HostZone) []int64 {
	flags := make([]int64, 0, len(hostZone.Validators))
	for _, validator := range hostZone.Validators {
		flags = append(flags, validator.DelegationChangesInProgress)
	}
	return flags
}

func contains(list []string, item string) bool {
	for _, entry := range list {
		if entry == item {
			return true
		}
	}
	return false
}
```

- [ ] **Step 3: Run without the fixture, then with it**

Run: `go test ./app/upgrades/v35/... -run 'TestMainnetExportTestSuite' -v`
Expected without the fixture: `--- SKIP: TestMainnetExportTestSuite` with the README hint. With Task 2's fixture in place: `--- PASS: TestMainnetExportTestSuite/TestUpgradeFromMainnetExport`.

If `govKeeper.Create` fails (wasmvm cannot load in the test app), that is a finding to report, not something to route around: the PR 3 synthetic suite's `TestMoveDeployKeyContractAdminsToGov` fails the same way and must be fixed first.

- [ ] **Step 4: Commit**

```bash
git add app/upgrades/v35/mainnet_export_test.go
git commit -m "test(v35): mainnet-export suite replays the full handler with the real constants

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Task 4: The coverage-check script

**Files:**
- Create: `scripts/wind-down/coverage_check.py`
- Create: `scripts/wind-down/test_coverage_check.py`

**Interfaces:**
- Consumes: a trimmed Stride export JSON (`app_state.bank.supply`, `app_state.bank.balances`, `app_state.stakeibc.host_zone_list`, `app_state.ibc.channel_genesis.channels` for the transfer-port channel ids; `.json` or `.json.gz`), a pools file, the Osmosis REST API (vault balances, pool contract addresses, transmuter `get_total_pool_liquidity` smart queries; the same endpoints `check_transmuter_pool.py` uses).
- Produces: a table on stdout and exit code 1 on any shortfall. `evaluate(...)` is the pure function the unit test drives with an injected liquidity fetcher.
- Depends on: Task 1
- Review: yes (this script is the assertion the design admits it lacks on chain, §10)

Pools file shape (`pools.json`, written by ops as pools are created; one entry per in-scope stToken):

```json
{
  "stuatom": {
    "chain_id": "cosmoshub-4",
    "native_denom_on_osmosis": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
    "canonical_pool_id": "3595",
    "route_pools": [{"channel_id": "channel-0", "pool_id": "3601"}, {"channel_id": "channel-6", "pool_id": "3602"}]
  }
}
```

`native_denom_on_osmosis` is spelled out rather than derived so the script has one fewer map to keep in sync with `check_transmuter_pool.py`; that script's `native_denom_on_osmosis` is the cross-check.

- [ ] **Step 1: Write the failing unit test**

```python
# scripts/wind-down/test_coverage_check.py
"""Unit test for coverage_check.evaluate on a synthetic export with an injected fetcher."""

import unittest
from decimal import Decimal

import coverage_check

STRIDE_ESCROW_CH0 = coverage_check.escrow_address(channel_id="channel-0")
STRIDE_ESCROW_CH5 = coverage_check.escrow_address(channel_id="channel-5")
HOLDER = "stride1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"
VAULT = "osmo1vault"
NATIVE = "ibc/NATIVE"


def synthetic_export() -> dict:
    return {
        "app_state": {
            "bank": {
                "supply": [{"denom": "stuatom", "amount": "1000000"}],
                "balances": [
                    {"address": STRIDE_ESCROW_CH0, "coins": [{"denom": "stuatom", "amount": "100000"}]},
                    {"address": STRIDE_ESCROW_CH5, "coins": [{"denom": "stuatom", "amount": "300000"}]},
                    {"address": HOLDER, "coins": [{"denom": "stuatom", "amount": "600000"}]},
                ],
            },
            "stakeibc": {
                "host_zone_list": [
                    {"chain_id": "cosmoshub-4", "host_denom": "uatom", "redemption_rate": "1.500000000000000000"}
                ]
            },
            "ibc": {
                "channel_genesis": {
                    "channels": [
                        {"port_id": "transfer", "channel_id": "channel-0"},
                        {"port_id": "icacontroller-GAIA.DELEGATION", "channel_id": "channel-1"},
                        {"port_id": "transfer", "channel_id": "channel-5"},
                    ]
                }
            },
        }
    }


def pools() -> dict:
    return {
        "stuatom": {
            "chain_id": "cosmoshub-4",
            "native_denom_on_osmosis": NATIVE,
            "canonical_pool_id": "1",
            "route_pools": [{"channel_id": "channel-0", "pool_id": "2"}],
        }
    }


class CoverageCheckTest(unittest.TestCase):
    def test_covered_when_every_pool_holds_its_share(self) -> None:
        # supply 1,000,000 × 1.5 = 1,500,000 native needed; route escrow 100,000 × 1.5 = 150,000
        # in the route pool; the canonical pool gets everything else; the vault holds a surplus
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_000}}
        results = coverage_check.evaluate(
            export=synthetic_export(),
            pools=pools(),
            vault_balances={NATIVE: 10_000},
            fetch_liquidity=lambda pool_id: liquidity[pool_id],
        )
        self.assertEqual(len(results), 1)
        result = results[0]
        self.assertEqual(result.st_denom, "stuatom")
        self.assertEqual(result.required_native, 1_500_000)
        self.assertEqual(result.native_on_osmosis, 1_510_000)
        self.assertEqual(result.route_shortfalls, {})
        self.assertTrue(result.covered)

    def test_shortfall_in_total_and_in_a_route_pool(self) -> None:
        liquidity = {"1": {NATIVE: 1_000_000}, "2": {NATIVE: 149_999}}
        result = coverage_check.evaluate(
            export=synthetic_export(),
            pools=pools(),
            vault_balances={},
            fetch_liquidity=lambda pool_id: liquidity[pool_id],
        )[0]
        self.assertFalse(result.covered)
        self.assertEqual(result.native_on_osmosis, 1_149_999)
        self.assertEqual(result.route_shortfalls, {"channel-0": 1})
        self.assertEqual(result.canonical_expected, 1_500_000 - 150_000)

    def test_rate_rounds_down_to_base_units(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"][0]["redemption_rate"] = "1.333333333333333333"
        result = coverage_check.evaluate(
            export=export,
            pools=pools(),
            vault_balances={NATIVE: 2_000_000},
            fetch_liquidity=lambda pool_id: {NATIVE: 0},
        )[0]
        self.assertEqual(result.required_native, int(Decimal("1000000") * Decimal("1.333333333333333333")))
        self.assertEqual(result.route_expected["channel-0"], 133_333)

    def test_missing_pool_entry_is_an_error(self) -> None:
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                export=synthetic_export(),
                pools={},
                vault_balances={},
                fetch_liquidity=lambda pool_id: {},
            )

    def test_sttoken_missing_from_supply_is_an_error(self) -> None:
        # A typo in the pools file must not turn into "required 0, covered"
        mistyped = {"stuatom-typo": pools()["stuatom"]}
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                export=synthetic_export(),
                pools=mistyped,
                vault_balances={},
                fetch_liquidity=lambda pool_id: {},
            )


if __name__ == "__main__":
    unittest.main()
```

Run: `cd scripts/wind-down && python3 -m unittest test_coverage_check -v`
Expected: `ModuleNotFoundError: No module named 'coverage_check'`.

- [ ] **Step 2: Write the script**

```python
#!/usr/bin/env python3
"""The wind-down coverage check (spec §10), run before each stToken's pools are funded and
again before the halt.

Per in-scope stToken: the native tokens held on Osmosis for that denom (the vault's balance
plus every pool's native liquidity) must cover Stride's bank supply of the stToken times the
frozen HostZone.RedemptionRate; each route pool must hold exactly its channel's escrow
balance times the rate; the canonical pool the remainder. Bank supply is the right reference
because stTokens that left Stride over IBC are escrowed, not burned.

Inputs: a trimmed Stride export (bank supply and balances, stakeibc host zones), a pools file
(see the PR 6 plan for the shape), the Osmosis vault address. Reads Osmosis over REST.
Read-only; prints a table and exits 1 on any shortfall.

Usage:
  python3 scripts/wind-down/coverage_check.py --export export.json.gz --pools pools.json \
      --vault osmo1... [--osmosis-rest https://osmosis-api.polkachu.com]
"""

import argparse
import base64
import gzip
import hashlib
import json
import pathlib
import sys
import urllib.parse
import urllib.request
from dataclasses import dataclass, field
from decimal import Decimal
from typing import Callable

OSMOSIS_REST_DEFAULT = "https://osmosis-api.polkachu.com"
USER_AGENT = "curl/8.0"
TIMEOUT_SECONDS = 30
TRANSFER_PORT = "transfer"
ESCROW_ADDRESS_VERSION = "ics20-1"
STRIDE_BECH32_PREFIX = "stride"

LiquidityFetcher = Callable[[str], dict[str, int]]


class CoverageInputError(Exception):
    """A pools entry or export section is missing or malformed; nothing was checked."""


@dataclass(frozen=True)
class RoutePool:
    channel_id: str
    pool_id: str


@dataclass(frozen=True)
class PoolSpec:
    chain_id: str
    native_denom_on_osmosis: str
    canonical_pool_id: str | None
    route_pools: list[RoutePool]


@dataclass
class CoverageResult:
    st_denom: str
    chain_id: str
    rate: Decimal
    supply: int
    required_native: int
    native_on_osmosis: int
    route_expected: dict[str, int] = field(default_factory=dict)
    route_shortfalls: dict[str, int] = field(default_factory=dict)
    canonical_expected: int = 0
    canonical_actual: int = 0

    @property
    def covered(self) -> bool:
        total_ok = self.native_on_osmosis >= self.required_native
        canonical_ok = self.canonical_actual >= self.canonical_expected
        return total_ok and canonical_ok and not self.route_shortfalls


# ----------------------------------------------------------------------------------------------
# Public API
# ----------------------------------------------------------------------------------------------

def main() -> int:
    args = parse_args()
    export = load_export(path=args.export)
    pools = load_pools(path=args.pools)
    vault_balances = fetch_vault_balances(osmosis_rest=args.osmosis_rest, vault=args.vault)
    fetcher = make_liquidity_fetcher(osmosis_rest=args.osmosis_rest)

    results = evaluate(export=export, pools=pools, vault_balances=vault_balances, fetch_liquidity=fetcher)
    print_table(results)

    shortfalls = [result for result in results if not result.covered]
    if shortfalls:
        print(f"\n{len(shortfalls)} stToken(s) NOT covered: {[r.st_denom for r in shortfalls]}")
        return 1
    print("\nevery stToken is covered")
    return 0


def evaluate(
    export: dict,
    pools: dict,
    vault_balances: dict[str, int],
    fetch_liquidity: LiquidityFetcher,
) -> list[CoverageResult]:
    """Pure: the §10 arithmetic for every stToken in the pools file."""
    supply_by_denom = {entry["denom"]: int(entry["amount"]) for entry in export["app_state"]["bank"]["supply"]}
    rates = {zone["chain_id"]: Decimal(zone["redemption_rate"]) for zone in export["app_state"]["stakeibc"]["host_zone_list"]}
    escrow_balances = escrow_balances_by_channel(export=export)
    if not pools:
        raise CoverageInputError("pools file has no entries")

    results = []
    for st_denom, raw_spec in sorted(pools.items()):
        spec = parse_pool_spec(raw=raw_spec)
        if spec.chain_id not in rates:
            raise CoverageInputError(f"{st_denom}: host zone {spec.chain_id} not in the export")
        rate = rates[spec.chain_id]
        if st_denom not in supply_by_denom:
            # A mistyped stToken denom would otherwise make the requirement zero and pass
            raise CoverageInputError(f"{st_denom}: not in the export's bank supply")
        supply = supply_by_denom[st_denom]
        results.append(evaluate_one(
            st_denom=st_denom, spec=spec, rate=rate, supply=supply,
            escrow_balances=escrow_balances, vault_balances=vault_balances, fetch_liquidity=fetch_liquidity,
        ))
    return results


# ----------------------------------------------------------------------------------------------
# Per-token arithmetic
# ----------------------------------------------------------------------------------------------

def evaluate_one(
    st_denom: str,
    spec: PoolSpec,
    rate: Decimal,
    supply: int,
    escrow_balances: dict[str, dict[str, int]],
    vault_balances: dict[str, int],
    fetch_liquidity: LiquidityFetcher,
) -> CoverageResult:
    native = spec.native_denom_on_osmosis
    result = CoverageResult(
        st_denom=st_denom, chain_id=spec.chain_id, rate=rate, supply=supply,
        required_native=native_for(st_amount=supply, rate=rate),
        native_on_osmosis=vault_balances.get(native, 0),
    )

    # Each route pool must hold exactly its channel's escrow share; a pool below that cannot
    # pay every holder on that chain, a pool above it is surplus the canonical pool should hold
    route_escrow_total = 0
    for route in spec.route_pools:
        escrow = escrow_balances.get(route.channel_id, {}).get(st_denom, 0)
        route_escrow_total += escrow
        expected = native_for(st_amount=escrow, rate=rate)
        actual = fetch_liquidity(route.pool_id).get(native, 0)
        result.route_expected[route.channel_id] = expected
        result.native_on_osmosis += actual
        if actual < expected:
            result.route_shortfalls[route.channel_id] = expected - actual

    # The canonical pool covers every other holder: everything not in a route escrow
    result.canonical_expected = native_for(st_amount=supply - route_escrow_total, rate=rate)
    if spec.canonical_pool_id is not None:
        result.canonical_actual = fetch_liquidity(spec.canonical_pool_id).get(native, 0)
        result.native_on_osmosis += result.canonical_actual
    return result


def native_for(st_amount: int, rate: Decimal) -> int:
    """stTokens × rate, floored to base units (the pools round exact-in output down too)."""
    return int(Decimal(st_amount) * rate)


# ----------------------------------------------------------------------------------------------
# Export reading
# ----------------------------------------------------------------------------------------------

def load_export(path: pathlib.Path) -> dict:
    opener = gzip.open if path.suffix == ".gz" else open
    with opener(path, "rt", encoding="utf-8") as handle:
        return json.load(handle)


def load_pools(path: pathlib.Path) -> dict:
    return json.loads(path.read_text())


def parse_pool_spec(raw: dict) -> PoolSpec:
    try:
        return PoolSpec(
            chain_id=raw["chain_id"],
            native_denom_on_osmosis=raw["native_denom_on_osmosis"],
            canonical_pool_id=raw.get("canonical_pool_id"),
            route_pools=[RoutePool(channel_id=r["channel_id"], pool_id=r["pool_id"]) for r in raw.get("route_pools", [])],
        )
    except KeyError as missing:
        raise CoverageInputError(f"pools entry missing {missing}") from missing


def escrow_address(channel_id: str) -> str:
    """ibc-go's GetEscrowAddress: ADR-028 hash of "ics20-1" NUL "transfer/<channel>", 20 bytes."""
    pre_image = ESCROW_ADDRESS_VERSION.encode() + b"\x00" + f"{TRANSFER_PORT}/{channel_id}".encode()
    return bech32_encode(prefix=STRIDE_BECH32_PREFIX, data=hashlib.sha256(pre_image).digest()[:20])


def escrow_balances_by_channel(export: dict) -> dict[str, dict[str, int]]:
    """channel id -> denom -> amount, for every transfer-port channel in the export's IBC state."""
    balances_by_address = {
        entry["address"]: {coin["denom"]: int(coin["amount"]) for coin in entry["coins"]}
        for entry in export["app_state"]["bank"]["balances"]
    }

    # Escrow accounts are not labelled in an export, but every channel is: derive the escrow
    # address of each transfer-port channel, so no channel can be missed by a scan heuristic
    channels = export["app_state"]["ibc"]["channel_genesis"]["channels"]
    transfer_channel_ids = [channel["channel_id"] for channel in channels if channel["port_id"] == TRANSFER_PORT]
    if not transfer_channel_ids:
        raise CoverageInputError("export has no transfer channels under app_state.ibc.channel_genesis.channels")

    return {
        channel_id: balances_by_address[escrow_address(channel_id=channel_id)]
        for channel_id in transfer_channel_ids
        if escrow_address(channel_id=channel_id) in balances_by_address
    }


# ----------------------------------------------------------------------------------------------
# Osmosis REST
# ----------------------------------------------------------------------------------------------

def get_json(url: str) -> dict:
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(request, timeout=TIMEOUT_SECONDS) as response:
        return json.load(response)


def fetch_vault_balances(osmosis_rest: str, vault: str) -> dict[str, int]:
    page = get_json(f"{osmosis_rest}/cosmos/bank/v1beta1/balances/{vault}?pagination.limit=1000")
    return {coin["denom"]: int(coin["amount"]) for coin in page["balances"]}


def make_liquidity_fetcher(osmosis_rest: str) -> LiquidityFetcher:
    def fetch(pool_id: str) -> dict[str, int]:
        pool = get_json(f"{osmosis_rest}/osmosis/poolmanager/v1beta1/pools/{pool_id}")["pool"]
        contract = pool["contract_address"]
        query = base64.b64encode(json.dumps({"get_total_pool_liquidity": {}}).encode()).decode()
        response = get_json(f"{osmosis_rest}/cosmwasm/wasm/v1/contract/{contract}/smart/{query}")
        return {coin["denom"]: int(coin["amount"]) for coin in response["data"]["total_pool_liquidity"]}
    return fetch


# ----------------------------------------------------------------------------------------------
# Output and CLI
# ----------------------------------------------------------------------------------------------

def print_table(results: list[CoverageResult]) -> None:
    header = f"{'stToken':<14}{'zone':<16}{'supply':>20}{'rate':>22}{'required':>22}{'on osmosis':>22}  status"
    print(header)
    print("-" * len(header))
    for result in results:
        status = "OK" if result.covered else "SHORT"
        print(f"{result.st_denom:<14}{result.chain_id:<16}{result.supply:>20}{str(result.rate):>22}"
              f"{result.required_native:>22}{result.native_on_osmosis:>22}  {status}")
        for channel_id, shortfall in sorted(result.route_shortfalls.items()):
            print(f"    route {channel_id}: short by {shortfall} (expected {result.route_expected[channel_id]})")
        if result.canonical_actual < result.canonical_expected:
            print(f"    canonical: {result.canonical_actual} < expected {result.canonical_expected}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--export", type=pathlib.Path, required=True, help="trimmed Stride export (.json or .json.gz)")
    parser.add_argument("--pools", type=pathlib.Path, required=True, help="pools file, one entry per stToken")
    parser.add_argument("--vault", required=True, help="Osmosis vault address (spec §4)")
    parser.add_argument("--osmosis-rest", default=OSMOSIS_REST_DEFAULT)
    return parser.parse_args()


# ----------------------------------------------------------------------------------------------
# bech32 (stdlib has none; BIP-173 reference, enough for encoding 20-byte addresses)
# ----------------------------------------------------------------------------------------------

BECH32_CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"


def bech32_encode(prefix: str, data: bytes) -> str:
    five_bit = convert_bits(data=data, from_bits=8, to_bits=5)
    checksum = bech32_checksum(prefix=prefix, data=five_bit)
    return prefix + "1" + "".join(BECH32_CHARSET[d] for d in five_bit + checksum)


def convert_bits(data: bytes, from_bits: int, to_bits: int) -> list[int]:
    accumulator = 0
    bits = 0
    result = []
    max_value = (1 << to_bits) - 1
    for value in data:
        accumulator = (accumulator << from_bits) | value
        bits += from_bits
        while bits >= to_bits:
            bits -= to_bits
            result.append((accumulator >> bits) & max_value)
    if bits:
        result.append((accumulator << (to_bits - bits)) & max_value)
    return result


def bech32_polymod(values: list[int]) -> int:
    generator = [0x3B6A57B2, 0x26508E6D, 0x1EA119FA, 0x3D4233DD, 0x2A1462B3]
    checksum = 1
    for value in values:
        top = checksum >> 25
        checksum = ((checksum & 0x1FFFFFF) << 5) ^ value
        for i in range(5):
            checksum ^= generator[i] if ((top >> i) & 1) else 0
    return checksum


def bech32_checksum(prefix: str, data: list[int]) -> list[int]:
    expanded = [ord(c) >> 5 for c in prefix] + [0] + [ord(c) & 31 for c in prefix]
    polymod = bech32_polymod(expanded + data + [0] * 6) ^ 1
    return [(polymod >> 5 * (5 - i)) & 31 for i in range(6)]


if __name__ == "__main__":
    sys.exit(main())
```

- [ ] **Step 3: Run the unit test, then a smoke run of the bech32 helper**

Run: `cd scripts/wind-down && python3 -m unittest test_coverage_check -v`
Expected: 5 tests `ok`.

Run:

```bash
cd scripts/wind-down && python3 -c 'import coverage_check; print(coverage_check.escrow_address(channel_id="channel-5"))'
curl -s -H 'User-Agent: curl/8.0' 'https://stride-api.polkachu.com/ibc/apps/transfer/v1/channels/channel-5/ports/transfer/escrow_address' | jq -r .escrow_address
```

Expected: both lines print the same `stride1...` address (the chain's own derivation of the channel-5 escrow account). A mismatch means the bech32 or the pre-image is wrong, and the unit test would not catch it because it only checks self-consistency.

- [ ] **Step 4: Commit**

```bash
git add scripts/wind-down/coverage_check.py scripts/wind-down/test_coverage_check.py
git commit -m "scripts(wind-down): coverage check (spec §10) over an export and the vault's Osmosis balances

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Task 5: Changelog

**Files:**
- Modify: `CHANGELOG.md` (the `## Unreleased` block, above the v34 lines)

**Interfaces:**
- Depends on: Task 1
- Review: no

- [ ] **Step 1: Add the v35 entries at the top of `## Unreleased`'s `### On-Chain changes` list**

Renumber the existing v34 items after them. Insert (PR numbers filled in as each PR opens):

```markdown
1. v35: remove every user- and admin-facing message the wind-down no longer needs (stakeibc liquid stake, LSM liquid stake, redeem, register host zone, trade routes, rebate, trade controller, rebalance, clear balance, resume; staketia and stakedym redeem and resume; icaoracle add/instantiate; icqoracle register/remove; auction; airdrop; claim); message types and registrations stay so historical txs still decode ([#PR1](https://github.com/Stride-Labs/stride/pull/PR1))
2. v35: freeze every redemption rate by deleting the epoch calls that move stake or the rate (rate update, reinvest, delegate, rebalance, reward transfer, withdrawal-address set, deposit and unbonding record creation, reward-collector auction) and the slash callback's rate rewrite; admin-gate `UpdateValidatorSharesExchRate` and `CalibrateDelegation` and lift the calibration cap ([#PR2](https://github.com/Stride-Labs/stride/pull/PR2))
3. v35 upgrade handler: autopilot stakeibc off, ICA host allow-list minus liquid stake and redeem, wasm upload access and the deploy-key contract admins to gov, comdex-1 deprecated, dYdX trade route deleted, ICA oracles off, rate limiter emptied, stale `DelegationChangesInProgress` flags reset on zones with nothing in flight, haqq slash-path and every withdrawal-balance ICQ purged, haqq_11235-1 delegations reconciled (net −1,758 ISLM) ([#PR3](https://github.com/Stride-Labs/stride/pull/PR3))
4. v35: wind-down admin txs `MsgUndelegateFromValidators`, `MsgTransferFromIca` (to the hard-coded Osmosis vault over a hard-coded per-host channel) and `MsgTransferStaketiaClaimBalance` ([#PR4](https://github.com/Stride-Labs/stride/pull/PR4))
5. v35: `MsgSweepTokensOffStride`, the sweep-operator-gated batched transfer of stTokens and STRD to holders' Osmosis addresses and of whitelisted vouchers back to their source chains ([#PR5](https://github.com/Stride-Labs/stride/pull/PR5))
6. v35: release gate — operator address constants, mainnet-export handler suite, coverage-check script ([#PR6](https://github.com/Stride-Labs/stride/pull/PR6))
```

- [ ] **Step 2: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs: changelog for v35 (protocol wind-down)

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Task 6: Merge gate, dry run and the pre-proposal checklist

**Files:** none new; this task runs the artifacts of Tasks 1 to 5 together.

**Interfaces:**
- Depends on: Tasks 1, 2, 3, 4, 5 (post-wave; run alone after the merges)
- Review: yes (the release decision is made from this task's output)

- [ ] **Step 1: Build, lint, full suite**

```bash
go build ./... && go vet ./...
make lint
make test-unit
```

Expected: build and vet clean; lint clean; every package passes except `utils` `TestCreateModuleAccount`, which fails on `main` too (pre-existing, not this branch's). `go test ./app/upgrades/v35/... -v` must show `TestMainnetExportTestSuite/TestUpgradeFromMainnetExport` as `PASS`, not `SKIP` (the fixture from Task 2 is present).

- [ ] **Step 2: Staleness gate and constants**

```bash
python3 app/upgrades/v35/testdata/verify_constants.py
go test ./x/stakeibc/types/... -run 'TestOperatorAddresses|TestHostToOsmosisTransferChannel|TestSweepUnwindChannels' -v
git diff wind-down-pr5-sweep-tx...HEAD -- go.mod
```

Expected: all `[PASS]` and exit 0; the three Go tests `PASS`; the `go.mod` diff is empty (module path untouched).

- [ ] **Step 3: Localstride dry run (ops, not a code step)**

`make upgrade-localstride` needs two binaries. The tagged previous release drops the triggered plan (its `InitStrideAppForTestnet` writes to a discarded CheckTx cache; `main` has the `NewUncachedContext` fix), and a binary that already registers the `v35` handler panics `BINARY UPDATED BEFORE TRIGGER` at the plan height. So: build the previous release from `git archive` into the scratchpad with only the `NewUncachedContext` line patched into `InitStrideAppForTestnet` (no v35 handler), sync with the plain previous-release binary first, testnetify with the patched one until the log shows `UPGRADE "v35" NEEDED`, then swap to this branch's binary. Never `pkill -f "strided start"` from an agent shell (the pattern matches the shell itself); use `pkill -x strided`. Rehearsal write-ups go in `integration-tests/rehearsal/` on a `REHEARSAL ONLY` branch; `localstride/scratch` and `exports` are gitignored.

What the dry run must show, in order: one `v35:` log line per handler helper in the order of the PR 3 plan's Task 10; afterwards `strided q stakeibc host-zone haqq_11235-1` with every `delegation_changes_in_progress` at 0 and the delta table's validators at their reconciled amounts; `strided q autopilot params` with `stakeibc_active: false`; `strided q interchainaccounts host params` without the liquid-stake and redeem URLs; `strided q wasm params` with upload access `AnyOfAddresses [gov]`; `strided q ratelimit list-rate-limits` empty; then, with the branch binary, `strided tx stakeibc liquid-stake ...` failing with `can't route message` and one `MsgTransferFromIca` submitted from the admin key and visible as an ICA packet on the delegation port (the full ops flow of §11's localstride bullet — redemption before the upgrade, day-epoch unbond, drain, sweep, claim, one ICA transfer to a second local chain and one sweep batch — is the rehearsal write-up's scope and is recorded there, not here).

- [ ] **Step 4: The §9 pre-proposal checklist, each line mapped to what satisfies it**

Reproduce this table in the rehearsal write-up and tick every row before the proposal is submitted:

| §9 line | Satisfied by |
|---|---|
| Full accounting check: every in-scope validator's recorded delegation equals the delegation ICA's on-chain delegation, differences in the haqq table or explained | `python3 scripts/wind-down/measure_delegation_drift.py` at height H (zero over-recorded validators outside haqq; haqq differences equal the table) and `verify_constants.py` `[PASS]` on every `haqq delta` line |
| Haqq delta table, drift measurement and mainnet-export tests match the chain at one recent height | `testdata/README.md` provenance height H == the drift measurement height == the height `gen_delta_table.py` ran at; `TestMainnetExportTestSuite` `PASS` |
| Delegation ICA withdraw address on every in-scope host is the zone's withdrawal ICA | per host: `<hostd> q distribution withdraw-address <delegation_ica_address>` equals the zone's `withdrawal_ica_address` from `strided q stakeibc host-zone <chain>` (eleven commands, paste the outputs) |
| The two address constants, the channel map and `SweepUnwindChannels` re-verified: maps against the hosts, each address by a test transfer to it and a signed spend from it | `verify_constants.py` (both map sections `[PASS]`); the four tx hashes from Task 1 Step 2, re-checked with `strided q tx <hash>` / `osmosisd q tx <hash>` |
| Staketia and stakedym operators ready to act on day 0 | written confirmation from each operator, linked in the write-up |
| (§12, before the proposal) Band's light client of Stride recovered | `curl <band REST>/ibc/core/client/v1/client_states/07-tendermint-169` shows a fresh `latest_height` and the delegation ICA restore on channel-768 no longer `STATE_INIT` |
| (§12, before the proposal) the stuck Cosmos Hub pipeline cleared | the cause is two over-recorded validators, not the stale flags (§3): `strided tx stakeibc update-delegation cosmoshub-4 <NodeGuardians valoper>` and the same for Forbole have been run from the admin key and their callbacks applied (`strided q stakeibc host-zone cosmoshub-4` shows both at the chain's delegation with a sub-one `shares_to_tokens_rate`); the next four-day day epoch has passed; `strided q records list-epoch-unbonding-record` shows no cosmoshub-4 record in `UNBONDING_RETRY_QUEUE` |

- [ ] **Step 5: Tag readiness note (no commit)**

Post the output of Steps 1 to 4 on the PR. The release tag and the `/v35` module-path bump happen after this PR merges, as a separate manual step outside every plan.

---

## Self-review

**Spec coverage.** §4 addresses filled and proven (Task 1); §9 pre-proposal checklist mapped line by line (Task 6 Step 4) and the address proof (Task 1 Step 2); §10 coverage check as a script with the exact three inequalities (Task 4); §11 "Handler" mainnet-export suite asserting every §5 effect: autopilot, allow-list, wasm params and admins, comdex, trade route, oracles, rate limiter, both ICQ purges, haqq deltas applied in full with the real table, stale flags reset exactly where nothing is in flight (Task 3); §11 "Ops scripts" coverage check tested against a synthetic export (Task 4; the batch builder is PR 5's); §12 item 6 changelog and the two constants (Tasks 5, 1); §13's `verify_constants` pattern and the localstride two-binary note (Tasks 2, 6). The decode test, the "no handler" guard and the export-suite-independent handler tests are PRs 1 and 3.

**Placeholders.** The two address values are read from spec §4 on the day (today's values are shown in every command that uses them) and guarded by the double-check and the signed spends of Task 1 Step 2 plus `verify_constants.py`'s sequence check, not by a test of which key they are; `<H>`, `<date>`, `<n>` in the README and the PR numbers in the changelog are filled at execution time by the commands beside them. Every code step carries its code.

**Review fixes folded in (2026-09-29).** No admin-distinctness assertion on either address (the vault is the admin multisig by design); the Hub checklist row follows the spec's diagnosis (two over-recorded validators, cleared by `update-delegation`); `DelegationChangesInProgress` is `int64`; the wasm section instantiates a stand-in hackatom per fixture contract instead of a genesis import wasmd rejects, using `testdata.HackatomContractWasm()` like PR 3; the rate-limiting REST paths are ibc-go v11's (`rate-limiting`, `ratelimit/blacklisted_denoms`, `ratelimit/whitelisted_addresses`); the coverage check errors on a stToken missing from the supply and derives escrow addresses from the export's channel list; the bech32 smoke check compares against the chain's escrow-address endpoint; records are compared by `String()`; channels are listed per connection; `urllib.parse.quote`; `verify_constants.py` checks both address constants have a non-zero sequence.

**Consistency with the other plans.** Helper and constant names come from the PR 3 plan (Task 10 Step 1 and the helper signatures); the hackatom store-and-instantiate shape is PR 3's `storeAndInstantiateHackatom`; the constants file and its existing tests from the PR 4 plan Task 1 (this plan only replaces `TestOperatorAddressesParse`); the fixture sections supersede PR 3's README list (two entries renamed to genesis names, `wasm.codes` dropped, `records` added) and use the modules' genesis field names, which is what `AppCodec().UnmarshalJSON` into each `GenesisState` needs; the sweep operator's fail-closed behaviour and PR 4's transfer tests set the vars themselves, so filling them does not change their outcomes (Task 1 Step 4 reruns them).

**Review tags.** Task 1 (the two strings every transfer depends on), Task 3 (the release gate) and Task 4 (the off-chain assertion the design relies on) and Task 6 (release decision) are `yes`; the fixture assembly and the changelog are `no`.

**Address keys.** Both §4 values are the protocol-admin multisig (the sweep operator on Stride, the same bytes with the `osmo` prefix as the vault); that is the decided setup, so no test compares them against `utils.Admins`. The guard is the second-person check against §4, the signed spends, and `verify_constants.py`'s sequence check.
