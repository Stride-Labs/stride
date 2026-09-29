# Wind-Down PR 3: Upgrade Handler Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the `v35` upgrade package: one handler that flips the entry points, moves wasm control to gov, deprecates comdex, deletes the dYdX trade route, deactivates the ICA oracles, empties the rate limiter, resets stale in-progress flags, purges two kinds of pending ICQ, and trues up the haqq delegations, each step a unit-tested helper.

**Architecture:** `app/upgrades/v35/` mirrors v34: `constants.go`, `upgrades.go` (the handler calls one helper per concern, in a fixed order), one file per concern with its test in `UpgradeTestSuite`, and the v34 delta helper copied verbatim for the haqq table. Every helper logs and skips on missing state; only the wasm upload-access write returns an error. The handler is wired in `app/upgrades.go` with no store upgrades. The haqq table is produced by a new generator script from the drift measurement and is regenerated right before the proposal.

**Tech Stack:** Go 1.25, cosmos-sdk v0.54.3, ibc-go v11.2.0 (ICA controller/host, rate-limiting middleware), wasmd v0.70.2, testify suites via `app/apptesting`, Python 3 for the generator.

Spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` §5 (all of it), §11 "Handler", §12 item 3, §13 "Upgrade handler (PR 3)".

> **Branching:** PR 1 branches off `wind-down-design-consolidation`. Each later PR branches
> off the previous PR's branch (PR 2 off PR 1, PR 3 off PR 2, and so on) and the PRs are
> implemented and merged strictly in order: 1, 2, 3, 4, 5, 6. Branch names:
> `wind-down-pr1-remove-handlers`, `wind-down-pr2-freeze-by-code`,
> `wind-down-pr3-upgrade-handler`, `wind-down-pr4-admin-txs`, `wind-down-pr5-sweep-tx`,
> `wind-down-pr6-release-gate`.
> The Go module path stays `github.com/Stride-Labs/stride/v34` in every PR; the bump to
> `/v35` is a manual step after all six land and is out of scope for every plan.

This PR's branch is `wind-down-pr3-upgrade-handler`, created from `wind-down-pr2-freeze-by-code`.

## Global Constraints

- Import paths are `github.com/Stride-Labs/stride/v34/...` everywhere; the package directory is `app/upgrades/v35` and its Go package name is `v35`. Do not touch `go.mod`.
- `UpgradeName = "v35"`.
- Handler step order (spec §5, this plan): `RunMigrations` → autopilot → ICA host → wasm (upload access, then contract admins) → comdex `Deprecated` → dYdX trade route → oracles → rate limits → stale in-progress flags → haqq slash-path ICQ purge → withdrawal-balance ICQ purge → haqq delta table. The haqq purge must run before the delta table.
- Only `SetWasmUploadAccessToGov` may return an error from the handler. Every other helper logs and skips on missing or mismatched state and returns nothing.
- Wasm upload access becomes `AccessConfig{Permission: AccessTypeAnyOfAddresses, Addresses: [gov module address]}`; the gov module address is `authtypes.NewModuleAddress(govtypes.ModuleName).String()`. Contract admins move only for contracts whose current admin equals `WasmDeployKey = "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh"`.
- ICA host allow-list: remove exactly `/stride.stakeibc.MsgLiquidStake` and `/stride.stakeibc.MsgRedeemStake` from the existing list; keep `/stride.stakeibc.MsgClaimUndelegatedTokens` and everything else; never rewrite the list from a constant.
- Trade route to delete: reward denom `uusdc`, host denom `adydx`.
- ICQ purge, haqq: delete queries with `ChainId == "haqq_11235-1"` and `CallbackId ∈ {"delegation", "validator", "calibrate"}`; clear `SlashQueryInProgress` on every haqq validator. ICQ purge, withdrawal: delete queries with `CallbackId == "withdrawalbalance"` on every chain. Nothing else is deleted.
- Stale-flag reset: only zones with `Deprecated == false`; only when the delegation ICA has an OPEN active channel with zero packet commitments; then every validator's `DelegationChangesInProgress` becomes 0. Otherwise log and skip the zone.
- Haqq deltas: `Delta = actual on-chain − tracked` in `aISLM`; both signs; applied all-or-nothing through the v34 helper copied as-is (only `package v35` and the `v35:` log prefix change). The net must be negative (over-recorded) for the 2026-09-29 table.
- Tests live in package `v35_test`, suite `UpgradeTestSuite` (`apptesting.AppTestHelper`), run with `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/<Name>' -v`.
- Commit messages end with `Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>`.

---

## File map

| File | Responsibility |
| --- | --- |
| `app/upgrades/v35/constants.go` | `UpgradeName`, chain ids, trade-route denoms, `WasmDeployKey`, `InScopeChainIds` |
| `app/upgrades/v35/upgrades.go` | `CreateUpgradeHandler`: migrations, then the helpers in order |
| `app/upgrades/v35/delegation_deltas.go` | v34's `DelegationDelta`, `mustInt`, `reconcileHostZoneDelegations`, `validateDelegationDeltas` (verbatim copy) |
| `app/upgrades/v35/haqq.go` | `HaqqDelegationDeltas`, `ReconcileHaqqDelegations` |
| `app/upgrades/v35/entry_points.go` | `DisableAutopilotStakeibc`, `RemoveStakeibcFromICAHostAllowList` |
| `app/upgrades/v35/wasm.go` | `SetWasmUploadAccessToGov`, `MoveDeployKeyContractAdminsToGov` |
| `app/upgrades/v35/stakeibc_state.go` | `DeprecateComdex`, `DeleteDydxTradeRoute` |
| `app/upgrades/v35/ibc_state.go` | `DeactivateICAOracles`, `RemoveAllRateLimits` |
| `app/upgrades/v35/stale_flags.go` | `ResetStaleDelegationChangesInProgress` |
| `app/upgrades/v35/icq_purge.go` | `PurgeHaqqSlashQueries`, `PurgeWithdrawalBalanceQueries` |
| `app/upgrades/v35/*_test.go` | one test file per source file, plus `upgrades_test.go` (suite + full-handler test) |
| `app/upgrades/v35/testdata/README.md` | what the PR 6 mainnet-export suite will need in `mainnet_export.json.gz` |
| `app/upgrades.go` | wiring |
| `scripts/wind-down/gen_delta_table.py`, `test_gen_delta_table.py` | drift.json → Go table literal |
| `scripts/wind-down/measure_delegation_drift.py`, `.gitignore` | `--output-dir` (default `scripts/wind-down/drift/`, ignored) and `--chain-id` so the measurement runs from a clean checkout |

---

### Task 1: Package skeleton, wiring, suite

**Files:**
- Create: `app/upgrades/v35/constants.go`
- Create: `app/upgrades/v35/upgrades.go`
- Create: `app/upgrades/v35/upgrades_test.go`
- Modify: `app/upgrades.go` (imports at lines 21-53 and the handler block after line 463)

**Interfaces:**
- Produces: `v35.UpgradeName`, `v35.CreateUpgradeHandler(...)` with the parameter list below (later tasks add helper calls inside it, never new parameters), `UpgradeTestSuite` with `SetupTest`.
- Review: no

- [ ] **Step 1: Write `constants.go`**

```go
package v35

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v35"

	HaqqChainId   = "haqq_11235-1"
	ComdexChainId = "comdex-1"

	// The one live trade route (dYdX USDC rewards → DYDX), keyed by the denoms as they
	// appear on the reward and host zones, not the IBC hashes (spec §13)
	DydxTradeRouteRewardDenom = "uusdc"
	DydxTradeRouteHostDenom   = "adydx"

	// The Stride deploy key that is admin of four Hyperlane contracts (spec §3); the
	// handler moves those admins to the gov module
	WasmDeployKey = "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh"
)

// InScopeChainIds are the non-deprecated stakeibc host zones the wind-down drains (spec §2).
// Used only by tests and the export README; the handler reads Deprecated from state.
var InScopeChainIds = []string{
	"celestia",
	"cosmoshub-4",
	"dydx-mainnet-1",
	"haqq_11235-1",
	"injective-1",
	"juno-1",
	"laozi-mainnet",
	"osmosis-1",
	"phoenix-1",
	"sommelier-3",
	"ssc-1",
}
```

- [ ] **Step 2: Write `upgrades.go` with migrations only**

```go
package v35

import (
	"context"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"
	ratelimitkeeper "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v34/x/autopilot/keeper"
	icaoraclekeeper "github.com/Stride-Labs/stride/v34/x/icaoracle/keeper"
	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

// CreateUpgradeHandler returns the v35 upgrade handler, the wind-down upgrade
// (docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md §5). Every step
// logs and skips on missing state; only the wasm upload-access write can fail the upgrade.
//
// icaHostKeeper and ratelimitKeeper are pointers because their methods have pointer
// receivers. The ICA controller and channel keepers used by the stale-flag reset are read
// through the stakeibc keeper's exported fields.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
	autopilotKeeper autopilotkeeper.Keeper,
	icaHostKeeper *icahostkeeper.Keeper,
	wasmKeeper wasmkeeper.Keeper,
	ratelimitKeeper *ratelimitkeeper.Keeper,
	icaOracleKeeper icaoraclekeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (protocol wind-down)...", UpgradeName))

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Helpers are added here by the later tasks, in the order fixed by the plan

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}
```

The unused keeper parameters compile (Go allows unused function parameters). `go vet` does not flag them.

- [ ] **Step 3: Wire the handler in `app/upgrades.go`**

Add the import beside the v34 line:

```go
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
```

Add after the v34 `SetUpgradeHandler` block (after line 463):

```go
	// v35 upgrade handler (protocol wind-down)
	app.UpgradeKeeper.SetUpgradeHandler(
		v35.UpgradeName,
		v35.CreateUpgradeHandler(
			app.ModuleManager,
			app.configurator,
			app.StakeibcKeeper,
			app.InterchainqueryKeeper,
			app.AutopilotKeeper,
			app.ICAHostKeeper,
			app.WasmKeeper,
			&app.RatelimitKeeper,
			app.ICAOracleKeeper,
		),
	)
```

No `case "v35"` in the store-upgrades switch: v35 adds and removes no stores.

- [ ] **Step 4: Write `upgrades_test.go` with the suite and an empty-state handler run**

```go
package v35_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// The handler must complete on a chain that has none of the mainnet state it acts on
// (every helper skips with a log); this is also the non-mainnet localnet case.
func (s *UpgradeTestSuite) TestUpgrade_EmptyState() {
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)
}
```

- [ ] **Step 5: Build and run**

Run: `go build ./... && go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/TestUpgrade_EmptyState' -v`
Expected: `--- PASS: TestUpgradeTestSuite/TestUpgrade_EmptyState`

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35/constants.go app/upgrades/v35/upgrades.go app/upgrades/v35/upgrades_test.go app/upgrades.go
git commit -m "feat(v35): upgrade package skeleton and wiring

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 2: Copy the v34 delegation delta helper

**Files:**
- Create: `app/upgrades/v35/delegation_deltas.go` (from `app/upgrades/v34/delegation_deltas.go`)

**Interfaces:**
- Produces: `DelegationDelta{Name, Address string; Delta sdkmath.Int}`, `mustInt(s string) sdkmath.Int`, `reconcileHostZoneDelegations(ctx, sk stakeibckeeper.Keeper, chainId string, deltas []DelegationDelta) (appliedDelta sdkmath.Int, applied bool)`, `validateDelegationDeltas(ctx, hostZone, deltas) (sdkmath.Int, bool)`.
- Review: no

- [ ] **Step 1: Copy and rename**

```bash
cp app/upgrades/v34/delegation_deltas.go app/upgrades/v35/delegation_deltas.go
sed -i '' -e 's/^package v34$/package v35/' -e 's/"v34: /"v35: /g' -e 's/(injective.go, celestia.go)/(haqq.go)/' -e 's/It is shared by the Injective and\n\/\/ Celestia reconciliations./It is used by the haqq reconciliation./' app/upgrades/v35/delegation_deltas.go
```

Then open the file and fix the two comment lines the `sed` could not join: the "Chain-agnostic helpers" comment should end with `(haqq.go).` and the `reconcileHostZoneDelegations` comment should say `It is used by the haqq reconciliation.` No code line changes.

- [ ] **Step 2: Verify the copy is code-identical**

Run: `diff <(grep -v '^//' app/upgrades/v34/delegation_deltas.go | sed 's/v34/v35/g') <(grep -v '^//' app/upgrades/v35/delegation_deltas.go)`
Expected: no output.

- [ ] **Step 3: Build**

Run: `go build ./app/...`
Expected: no output (the helpers are unused until Task 9, which is fine for unexported functions? No: Go reports unused *imports*, not unused functions; `mustInt` and the two functions compile unused).

- [ ] **Step 4: Commit**

```bash
git add app/upgrades/v35/delegation_deltas.go
git commit -m "feat(v35): copy the v34 delegation delta helper

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Parallel-safe tasks

Tasks 3 to 9 depend only on Tasks 1 and 2. Each adds one source file, one test file, and one call line inside the `// Helpers are added here` region of `upgrades.go`; merge the call lines in the order of the Global Constraints when the worktrees come back (the merge conflicts are one-line and expected). Task 10 is a post-wave integration task and runs after the merges.

### Task 3: Entry points (autopilot, ICA host allow-list)

**Files:**
- Create: `app/upgrades/v35/entry_points.go`
- Create: `app/upgrades/v35/entry_points_test.go`
- Modify: `app/upgrades/v35/upgrades.go` (call lines)

**Interfaces:**
- Consumes: `autopilotkeeper.Keeper.GetParams/SetParams`, `icahostkeeper.Keeper.GetParams/SetParams`.
- Produces: `DisableAutopilotStakeibc(ctx sdk.Context, k autopilotkeeper.Keeper)`, `RemoveStakeibcFromICAHostAllowList(ctx sdk.Context, k *icahostkeeper.Keeper)`.
- Depends on: Task 1
- Review: no

- [ ] **Step 1: Write the failing tests**

```go
package v35_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	icahosttypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/types"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	autopilottypes "github.com/Stride-Labs/stride/v34/x/autopilot/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func (s *UpgradeTestSuite) TestDisableAutopilotStakeibc() {
	s.App.AutopilotKeeper.SetParams(s.Ctx, autopilottypes.Params{StakeibcActive: true, ClaimActive: true})

	v35.DisableAutopilotStakeibc(s.Ctx, s.App.AutopilotKeeper)

	params := s.App.AutopilotKeeper.GetParams(s.Ctx)
	s.Require().False(params.StakeibcActive, "stakeibc autopilot should be off")
	s.Require().True(params.ClaimActive, "claim autopilot should be untouched")
}

func (s *UpgradeTestSuite) TestRemoveStakeibcFromICAHostAllowList() {
	// The mainnet list (queried 2026-09-29), in its mainnet order
	before := []string{
		sdk.MsgTypeURL(&banktypes.MsgSend{}),
		sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
		"/cosmos.staking.v1beta1.MsgDelegate",
		"/cosmos.staking.v1beta1.MsgUndelegate",
		"/cosmos.staking.v1beta1.MsgBeginRedelegate",
		"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
		"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
		"/ibc.applications.transfer.v1.MsgTransfer",
		"/cosmos.gov.v1beta1.MsgVote",
		sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
	}
	s.App.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.Params{HostEnabled: true, AllowMessages: before})

	v35.RemoveStakeibcFromICAHostAllowList(s.Ctx, s.App.ICAHostKeeper)

	after := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().True(after.HostEnabled, "host stays enabled")
	s.Require().Equal(append(append([]string{}, before[:9]...), before[11]), after.AllowMessages,
		"exactly the liquid stake and redeem stake entries are removed, order preserved")
}

// Running the removal twice must be a no-op the second time (idempotent for a re-run on localnet)
func (s *UpgradeTestSuite) TestRemoveStakeibcFromICAHostAllowList_Idempotent() {
	s.App.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.Params{HostEnabled: true, AllowMessages: []string{
		"/cosmos.bank.v1beta1.MsgSend",
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
	}})

	v35.RemoveStakeibcFromICAHostAllowList(s.Ctx, s.App.ICAHostKeeper)

	after := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().Equal([]string{"/cosmos.bank.v1beta1.MsgSend", "/stride.stakeibc.MsgClaimUndelegatedTokens"}, after.AllowMessages)
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(DisableAutopilotStakeibc|RemoveStakeibcFromICAHostAllowList)' -v`
Expected: build failure `undefined: v35.DisableAutopilotStakeibc`.

- [ ] **Step 3: Write `entry_points.go`**

```go
package v35

import (
	"fmt"

	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v34/x/autopilot/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// DisableAutopilotStakeibc turns off the autopilot liquid-stake and redeem paths, which
// bypass the msg service router that PR 1 emptied (spec §5 "Entry points that bypass the router").
func DisableAutopilotStakeibc(ctx sdk.Context, k autopilotkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.StakeibcActive = false
	k.SetParams(ctx, params)
	ctx.Logger().Info("v35: autopilot StakeibcActive set to false")
}

// RemoveStakeibcFromICAHostAllowList drops MsgLiquidStake and MsgRedeemStake from the ICA host
// allow-list so interchain accounts on Stride cannot reach the removed handlers either.
// MsgClaimUndelegatedTokens stays: ICA-originated claims keep paying the open redemptions.
// The existing list is filtered in place, never rewritten from a constant.
func RemoveStakeibcFromICAHostAllowList(ctx sdk.Context, k *icahostkeeper.Keeper) {
	removed := map[string]bool{
		sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}): true,
		sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}): true,
	}

	params := k.GetParams(ctx)
	kept := make([]string, 0, len(params.AllowMessages))
	for _, msgTypeUrl := range params.AllowMessages {
		if removed[msgTypeUrl] {
			ctx.Logger().Info(fmt.Sprintf("v35: removing %s from the ICA host allow-list", msgTypeUrl))
			continue
		}
		kept = append(kept, msgTypeUrl)
	}
	params.AllowMessages = kept
	k.SetParams(ctx, params)
}
```

- [ ] **Step 4: Add the calls to `upgrades.go`**

Replace the `// Helpers are added here ...` comment region with (keeping it as the anchor for the other tasks):

```go
		// Entry points that bypass the msg service router (spec §5)
		DisableAutopilotStakeibc(ctx, autopilotKeeper)
		RemoveStakeibcFromICAHostAllowList(ctx, icaHostKeeper)

		// Helpers are added here by the later tasks, in the order fixed by the plan
```

- [ ] **Step 5: Run the tests**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(DisableAutopilotStakeibc|RemoveStakeibcFromICAHostAllowList|Upgrade_EmptyState)' -v`
Expected: all `PASS`.

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35/entry_points.go app/upgrades/v35/entry_points_test.go app/upgrades/v35/upgrades.go
git commit -m "feat(v35): disable autopilot stakeibc and drop stakeibc msgs from the ICA host allow-list

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 4: Wasm to gov

**Files:**
- Create: `app/upgrades/v35/wasm.go`
- Create: `app/upgrades/v35/wasm_test.go`
- Modify: `app/upgrades/v35/upgrades.go` (call lines)

**Interfaces:**
- Consumes: `wasmkeeper.Keeper.GetParams/SetParams/IterateContractInfo`, `wasmkeeper.NewGovPermissionKeeper(k).UpdateContractAdmin(ctx, contractAddr, caller, newAdmin)`.
- Produces: `SetWasmUploadAccessToGov(ctx, k wasmkeeper.Keeper) error`, `MoveDeployKeyContractAdminsToGov(ctx, k wasmkeeper.Keeper)`, `GovModuleAddress() sdk.AccAddress`.
- Depends on: Task 1
- Review: yes (the one error path that halts the upgrade; the only step that changes who controls code on chain)

- [ ] **Step 1: Write the failing tests**

The admin test instantiates a real contract: wasmd ships `hackatom.wasm` in its testdata package (`$(go env GOMODCACHE)/github.com/!cosm!wasm/wasmd@v0.70.2/x/wasm/keeper/testdata/hackatom.wasm`, exported as `testdata.HackatomContractWasm()`; confirmed present). The gov permission keeper is used to store and instantiate it, so the test app's upload access setting does not matter. Hackatom's init message is `{"verifier": "<addr>", "beneficiary": "<addr>"}`.

```go
package v35_test

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/testdata"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
)

func (s *UpgradeTestSuite) TestSetWasmUploadAccessToGov() {
	deployKey := sdk.MustAccAddressFromBech32(v35.WasmDeployKey)
	params := s.App.WasmKeeper.GetParams(s.Ctx)
	params.CodeUploadAccess = wasmtypes.AccessTypeAnyOfAddresses.With(deployKey, apptesting.CreateRandomAccounts(1)[0])
	s.Require().NoError(s.App.WasmKeeper.SetParams(s.Ctx, params))

	err := v35.SetWasmUploadAccessToGov(s.Ctx, s.App.WasmKeeper)
	s.Require().NoError(err)

	after := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, after.CodeUploadAccess.Permission)
	s.Require().Equal([]string{v35.GovModuleAddress().String()}, after.CodeUploadAccess.Addresses)
	s.Require().Equal(params.InstantiateDefaultPermission, after.InstantiateDefaultPermission, "instantiate permission untouched")
}

// storeAndInstantiateHackatom stores hackatom.wasm through the gov permission keeper (which
// bypasses upload access) and instantiates it with the given admin. Returns the contract address.
func (s *UpgradeTestSuite) storeAndInstantiateHackatom(admin sdk.AccAddress) sdk.AccAddress {
	govKeeper := wasmkeeper.NewGovPermissionKeeper(s.App.WasmKeeper)
	creator := apptesting.CreateRandomAccounts(1)[0]

	codeId, _, err := govKeeper.Create(s.Ctx, creator, testdata.HackatomContractWasm(), nil)
	s.Require().NoError(err, "store hackatom")

	initMsg, err := json.Marshal(map[string]string{
		"verifier":    creator.String(),
		"beneficiary": creator.String(),
	})
	s.Require().NoError(err)

	contractAddr, _, err := govKeeper.Instantiate(s.Ctx, codeId, creator, admin, initMsg, "hackatom", sdk.NewCoins())
	s.Require().NoError(err, "instantiate hackatom")
	return contractAddr
}

func (s *UpgradeTestSuite) TestMoveDeployKeyContractAdminsToGov() {
	deployKey := sdk.MustAccAddressFromBech32(v35.WasmDeployKey)
	otherAdmin := apptesting.CreateRandomAccounts(1)[0]

	deployKeyContract := s.storeAndInstantiateHackatom(deployKey)
	otherAdminContract := s.storeAndInstantiateHackatom(otherAdmin)
	noAdminContract := s.storeAndInstantiateHackatom(nil)

	v35.MoveDeployKeyContractAdminsToGov(s.Ctx, s.App.WasmKeeper)

	gov := v35.GovModuleAddress().String()
	s.Require().Equal(gov, s.App.WasmKeeper.GetContractInfo(s.Ctx, deployKeyContract).Admin, "deploy-key contract moved to gov")
	s.Require().Equal(otherAdmin.String(), s.App.WasmKeeper.GetContractInfo(s.Ctx, otherAdminContract).Admin, "other admin untouched")
	s.Require().Equal("", s.App.WasmKeeper.GetContractInfo(s.Ctx, noAdminContract).Admin, "admin-less contract untouched")
}

func (s *UpgradeTestSuite) TestMoveDeployKeyContractAdminsToGov_NoContracts() {
	s.Require().NotPanics(func() { v35.MoveDeployKeyContractAdminsToGov(s.Ctx, s.App.WasmKeeper) })
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(SetWasmUploadAccessToGov|MoveDeployKeyContractAdminsToGov)' -v`
Expected: build failure `undefined: v35.SetWasmUploadAccessToGov`.

- [ ] **Step 3: Write `wasm.go`**

```go
package v35

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GovModuleAddress is the gov module account, the only address that may upload code or
// administer the deploy key's contracts after the upgrade.
func GovModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(govtypes.ModuleName)
}

// SetWasmUploadAccessToGov restricts code upload to the gov module. This is the one handler
// step that fails the upgrade on error: it can only fail on invalid params, and leaving
// upload open to the two deploy keys with nothing but a log line is the worse outcome (spec §5).
func SetWasmUploadAccessToGov(ctx sdk.Context, k wasmkeeper.Keeper) error {
	params := k.GetParams(ctx)
	params.CodeUploadAccess = wasmtypes.AccessTypeAnyOfAddresses.With(GovModuleAddress())
	if err := k.SetParams(ctx, params); err != nil {
		return errorsmod.Wrap(err, "v35: unable to set wasm code upload access to the gov module")
	}
	ctx.Logger().Info("v35: wasm code upload access restricted to the gov module")
	return nil
}

// MoveDeployKeyContractAdminsToGov sets the admin of every contract currently administered by
// WasmDeployKey to the gov module. The gov permission keeper's authorization policy allows
// the change without the current admin's signature. A failed update is logged and skipped.
func MoveDeployKeyContractAdminsToGov(ctx sdk.Context, k wasmkeeper.Keeper) {
	govKeeper := wasmkeeper.NewGovPermissionKeeper(k)
	gov := GovModuleAddress()

	// Collect first: IterateContractInfo must not see the store change under it
	var toMove []sdk.AccAddress
	k.IterateContractInfo(ctx, func(contractAddr sdk.AccAddress, info wasmtypes.ContractInfo) bool {
		if info.Admin == WasmDeployKey {
			toMove = append(toMove, contractAddr)
		}
		return false
	})

	for _, contractAddr := range toMove {
		if err := govKeeper.UpdateContractAdmin(ctx, contractAddr, gov, gov); err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: unable to move admin of %s to gov, skipping: %s", contractAddr, err))
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: admin of %s moved from the deploy key to gov", contractAddr))
	}
	ctx.Logger().Info(fmt.Sprintf("v35: %d contract admin(s) moved to gov", len(toMove)))
}
```

- [ ] **Step 4: Add the calls to `upgrades.go`** (immediately after the entry-point calls, before the anchor comment)

```go
		// Wasm control to gov: the upload-access write is the one step that may fail the upgrade
		if err := SetWasmUploadAccessToGov(ctx, wasmKeeper); err != nil {
			return vm, err
		}
		MoveDeployKeyContractAdminsToGov(ctx, wasmKeeper)
```

- [ ] **Step 5: Run the tests**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(SetWasmUploadAccessToGov|MoveDeployKeyContractAdminsToGov|Upgrade_EmptyState)' -v`
Expected: all `PASS`. If `Create` fails with a wasmvm error in the test app, that is a real finding (the test app cannot execute wasm) and must be reported, not worked around by seeding `ContractInfo` some other way.

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35/wasm.go app/upgrades/v35/wasm_test.go app/upgrades/v35/upgrades.go
git commit -m "feat(v35): wasm upload access and deploy-key contract admins to gov

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 5: Comdex flag and dYdX trade route

**Files:**
- Create: `app/upgrades/v35/stakeibc_state.go`
- Create: `app/upgrades/v35/stakeibc_state_test.go`
- Modify: `app/upgrades/v35/upgrades.go` (call lines)

**Interfaces:**
- Consumes: `stakeibckeeper.Keeper.GetHostZone/SetHostZone/GetTradeRoute/RemoveTradeRoute/GetAllTradeRoutes`.
- Produces: `DeprecateComdex(ctx, k stakeibckeeper.Keeper)`, `DeleteDydxTradeRoute(ctx, k stakeibckeeper.Keeper)`.
- Depends on: Task 1
- Review: no

- [ ] **Step 1: Write the failing tests**

```go
package v35_test

import (
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func (s *UpgradeTestSuite) TestDeprecateComdex() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: v35.ComdexChainId, Halted: false, Deprecated: false})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: "juno-1", Deprecated: false})

	v35.DeprecateComdex(s.Ctx, s.App.StakeibcKeeper)

	comdex, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(found)
	s.Require().True(comdex.Deprecated, "comdex should be deprecated")
	s.Require().False(comdex.Halted, "halted is not touched")

	juno, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "juno-1")
	s.Require().False(juno.Deprecated, "other zones untouched")
}

func (s *UpgradeTestSuite) TestDeprecateComdex_MissingZone() {
	s.Require().NotPanics(func() { v35.DeprecateComdex(s.Ctx, s.App.StakeibcKeeper) })
}

func (s *UpgradeTestSuite) TestDeleteDydxTradeRoute() {
	dydx := stakeibctypes.TradeRoute{
		RewardDenomOnRewardZone: v35.DydxTradeRouteRewardDenom,
		HostDenomOnHostZone:     v35.DydxTradeRouteHostDenom,
	}
	other := stakeibctypes.TradeRoute{RewardDenomOnRewardZone: "uusdc", HostDenomOnHostZone: "uatom"}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, dydx)
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, other)

	v35.DeleteDydxTradeRoute(s.Ctx, s.App.StakeibcKeeper)

	_, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, v35.DydxTradeRouteRewardDenom, v35.DydxTradeRouteHostDenom)
	s.Require().False(found, "dydx route deleted")
	s.Require().Len(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), 1, "the other route stays")
}

func (s *UpgradeTestSuite) TestDeleteDydxTradeRoute_MissingRoute() {
	s.Require().NotPanics(func() { v35.DeleteDydxTradeRoute(s.Ctx, s.App.StakeibcKeeper) })
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(DeprecateComdex|DeleteDydxTradeRoute)' -v`
Expected: build failure `undefined: v35.DeprecateComdex`.

- [ ] **Step 3: Write `stakeibc_state.go`**

```go
package v35

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

// DeprecateComdex flags comdex-1 as deprecated so it carries the same flag as the other three
// dead zones (spec §5 "Comdex"). Halted is not touched; the flag is documentation and the
// wind-down txs refuse deprecated zones.
func DeprecateComdex(ctx sdk.Context, k stakeibckeeper.Keeper) {
	hostZone, found := k.GetHostZone(ctx, ComdexChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v35: host zone %s not found, skipping deprecation", ComdexChainId))
		return
	}
	hostZone.Deprecated = true
	k.SetHostZone(ctx, hostZone)
	ctx.Logger().Info(fmt.Sprintf("v35: %s marked deprecated", ComdexChainId))
}

// DeleteDydxTradeRoute removes the one live trade route (spec §5 "Trade route"). Its USDC is
// swept from the withdrawal ICA by MsgTransferFromIca; the converter ICAs are written off.
func DeleteDydxTradeRoute(ctx sdk.Context, k stakeibckeeper.Keeper) {
	if _, found := k.GetTradeRoute(ctx, DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom); !found {
		ctx.Logger().Info(fmt.Sprintf("v35: trade route %s/%s not found, skipping deletion",
			DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom))
		return
	}
	k.RemoveTradeRoute(ctx, DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom)
	ctx.Logger().Info(fmt.Sprintf("v35: trade route %s/%s deleted", DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom))
}
```

- [ ] **Step 4: Add the calls to `upgrades.go`** (after the wasm calls)

```go
		// Stakeibc state flips (spec §5)
		DeprecateComdex(ctx, stakeibcKeeper)
		DeleteDydxTradeRoute(ctx, stakeibcKeeper)
```

- [ ] **Step 5: Run the tests**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(DeprecateComdex|DeleteDydxTradeRoute|Upgrade_EmptyState)' -v`
Expected: all `PASS`.

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35/stakeibc_state.go app/upgrades/v35/stakeibc_state_test.go app/upgrades/v35/upgrades.go
git commit -m "feat(v35): deprecate comdex-1 and delete the dYdX trade route

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 6: Oracles off, rate limiter emptied

**Files:**
- Create: `app/upgrades/v35/ibc_state.go`
- Create: `app/upgrades/v35/ibc_state_test.go`
- Modify: `app/upgrades/v35/upgrades.go` (call lines)

**Interfaces:**
- Consumes: `icaoraclekeeper.Keeper.GetAllOracles/ToggleOracle`, `ratelimitkeeper.Keeper.GetAllRateLimits/RemoveRateLimit/GetAllBlacklistedDenoms/RemoveDenomFromBlacklist/GetAllWhitelistedAddressPairs/RemoveWhitelistedAddressPair`.
- Produces: `DeactivateICAOracles(ctx, k icaoraclekeeper.Keeper)`, `RemoveAllRateLimits(ctx, k *ratelimitkeeper.Keeper)`.
- Depends on: Task 1
- Review: no

- [ ] **Step 1: Write the failing tests**

```go
package v35_test

import (
	ratelimittypes "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/types"

	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"
)

func (s *UpgradeTestSuite) TestDeactivateICAOracles() {
	for _, oracle := range []icaoracletypes.Oracle{
		{ChainId: "injective-1", ConnectionId: "connection-1", ChannelId: "channel-1", PortId: "port-1", IcaAddress: "inj1", ContractAddress: "inj1c", Active: true},
		{ChainId: "neutron-1", ConnectionId: "connection-2", ChannelId: "channel-2", PortId: "port-2", IcaAddress: "neutron1", ContractAddress: "neutron1c", Active: true},
		{ChainId: "osmosis-1", ConnectionId: "connection-3", ChannelId: "channel-3", PortId: "port-3", IcaAddress: "osmo1", ContractAddress: "osmo1c", Active: false},
	} {
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)
	}

	v35.DeactivateICAOracles(s.Ctx, s.App.ICAOracleKeeper)

	oracles := s.App.ICAOracleKeeper.GetAllOracles(s.Ctx)
	s.Require().Len(oracles, 3, "no oracle is removed")
	for _, oracle := range oracles {
		s.Require().False(oracle.Active, "%s should be inactive", oracle.ChainId)
	}
}

func (s *UpgradeTestSuite) TestRemoveAllRateLimits() {
	rateLimit := func(denom, channel string) ratelimittypes.RateLimit {
		return ratelimittypes.RateLimit{
			Path:  &ratelimittypes.Path{Denom: denom, ChannelOrClientId: channel},
			Quota: &ratelimittypes.Quota{MaxPercentSend: sdkmath.NewInt(10), MaxPercentRecv: sdkmath.NewInt(10), DurationHours: 24},
			Flow:  &ratelimittypes.Flow{Inflow: sdkmath.ZeroInt(), Outflow: sdkmath.ZeroInt(), ChannelValue: sdkmath.NewInt(1000)},
		}
	}
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit("stuatom", "channel-0"))
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit("stuatom", "channel-5"))
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit("stutia", "channel-162"))
	s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "stuevmos")
	s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, ratelimittypes.WhitelistedAddressPair{Sender: "stride1a", Receiver: "stride1b"})
	s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, ratelimittypes.WhitelistedAddressPair{Sender: "stride1c", Receiver: "stride1d"})

	v35.RemoveAllRateLimits(s.Ctx, &s.App.RatelimitKeeper)

	s.Require().Empty(s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx), "rate limits")
	s.Require().Empty(s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx), "blacklist")
	s.Require().Empty(s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx), "whitelist")
}

func (s *UpgradeTestSuite) TestRemoveAllRateLimits_Empty() {
	s.Require().NotPanics(func() { v35.RemoveAllRateLimits(s.Ctx, &s.App.RatelimitKeeper) })
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(DeactivateICAOracles|RemoveAllRateLimits)' -v`
Expected: build failure `undefined: v35.DeactivateICAOracles`.

- [ ] **Step 3: Write `ibc_state.go`**

```go
package v35

import (
	"fmt"

	ratelimitkeeper "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icaoraclekeeper "github.com/Stride-Labs/stride/v34/x/icaoracle/keeper"
)

// DeactivateICAOracles turns every ICA oracle off (spec §5 "Oracles and rate limits"). The
// redemption rate no longer moves, so there is nothing to post. ToggleOracle(false) skips the
// channel validation the true case does, so a half-registered oracle still deactivates.
func DeactivateICAOracles(ctx sdk.Context, k icaoraclekeeper.Keeper) {
	for _, oracle := range k.GetAllOracles(ctx) {
		if !oracle.Active {
			continue
		}
		if err := k.ToggleOracle(ctx, oracle.ChainId, false); err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: unable to deactivate oracle %s, skipping: %s", oracle.ChainId, err))
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: oracle %s deactivated", oracle.ChainId))
	}
}

// RemoveAllRateLimits empties the rate limiter: every limit, blacklisted denom and whitelisted
// address pair. The token sweep sends most of each stToken's on-Stride supply out over channel-5
// in a few days, which no limit would allow, and there is no mint path left to protect. The
// module and middleware stay in the stack with empty state.
func RemoveAllRateLimits(ctx sdk.Context, k *ratelimitkeeper.Keeper) {
	for _, rateLimit := range k.GetAllRateLimits(ctx) {
		k.RemoveRateLimit(ctx, rateLimit.Path.Denom, rateLimit.Path.ChannelOrClientId)
		ctx.Logger().Info(fmt.Sprintf("v35: rate limit removed for %s on %s", rateLimit.Path.Denom, rateLimit.Path.ChannelOrClientId))
	}
	for _, denom := range k.GetAllBlacklistedDenoms(ctx) {
		k.RemoveDenomFromBlacklist(ctx, denom)
		ctx.Logger().Info(fmt.Sprintf("v35: %s removed from the rate-limit blacklist", denom))
	}
	for _, pair := range k.GetAllWhitelistedAddressPairs(ctx) {
		k.RemoveWhitelistedAddressPair(ctx, pair.Sender, pair.Receiver)
		ctx.Logger().Info(fmt.Sprintf("v35: whitelisted pair %s -> %s removed", pair.Sender, pair.Receiver))
	}
}
```

- [ ] **Step 4: Add the calls to `upgrades.go`** (after the stakeibc state calls)

```go
		// Oracles and rate limits (spec §5)
		DeactivateICAOracles(ctx, icaOracleKeeper)
		RemoveAllRateLimits(ctx, ratelimitKeeper)
```

- [ ] **Step 5: Run the tests**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(DeactivateICAOracles|RemoveAllRateLimits|Upgrade_EmptyState)' -v`
Expected: all `PASS`.

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35/ibc_state.go app/upgrades/v35/ibc_state_test.go app/upgrades/v35/upgrades.go
git commit -m "feat(v35): deactivate ICA oracles and empty the rate limiter

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 7: Stale in-progress flag reset

**Files:**
- Create: `app/upgrades/v35/stale_flags.go`
- Create: `app/upgrades/v35/stale_flags_test.go`
- Modify: `app/upgrades/v35/upgrades.go` (call line)

**Interfaces:**
- Consumes: `stakeibckeeper.Keeper.GetAllHostZone/SetHostZone`, the keeper's exported `ICAControllerKeeper.GetOpenActiveChannel(ctx, connectionId, portId)` and `IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctx, portId, channelId)`, `icatypes.NewControllerPortID`, `stakeibctypes.FormatHostZoneICAOwner`.
- Produces: `ResetStaleDelegationChangesInProgress(ctx, k stakeibckeeper.Keeper)`.
- Depends on: Task 1
- Review: yes (it rewrites the flag the drain and the day epoch gate on; a wrong reset on a zone with a packet in flight lets a batch be double-submitted)

- [ ] **Step 1: Write the failing tests**

```go
package v35_test

import (
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"

	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// seedFlaggedZone stores a host zone with two validators carrying non-zero in-progress counters
func (s *UpgradeTestSuite) seedFlaggedZone(chainId, connectionId string, deprecated bool) {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:      chainId,
		ConnectionId: connectionId,
		Deprecated:   deprecated,
		Validators: []*stakeibctypes.Validator{
			{Address: chainId + "valoper1", Delegation: sdkmath.NewInt(100), DelegationChangesInProgress: 12},
			{Address: chainId + "valoper2", Delegation: sdkmath.NewInt(100), DelegationChangesInProgress: 1},
			{Address: chainId + "valoper3", Delegation: sdkmath.NewInt(100), DelegationChangesInProgress: 0},
		},
	})
}

func (s *UpgradeTestSuite) flags(chainId string) []uint64 {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
	s.Require().True(found)
	flags := []uint64{}
	for _, validator := range hostZone.Validators {
		flags = append(flags, validator.DelegationChangesInProgress)
	}
	return flags
}

func (s *UpgradeTestSuite) mockDelegationChannel(chainId, connectionId, channelId string) (portId string) {
	owner := stakeibctypes.FormatHostZoneICAOwner(chainId, stakeibctypes.ICAAccountType_DELEGATION)
	s.MockICAChannel(connectionId, channelId, owner, chainId+"ica")
	portId, _ = icatypes.NewControllerPortID(owner)
	return portId
}

func (s *UpgradeTestSuite) TestResetStaleDelegationChangesInProgress() {
	// cosmoshub-4: open channel, nothing in flight -> reset
	s.seedFlaggedZone("cosmoshub-4", "connection-0", false)
	s.mockDelegationChannel("cosmoshub-4", "connection-0", "channel-863")

	// juno-1: open channel with an unacked packet -> skipped
	s.seedFlaggedZone("juno-1", "connection-1", false)
	junoPort := s.mockDelegationChannel("juno-1", "connection-1", "channel-10")
	s.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(s.Ctx, junoPort, "channel-10", 7, []byte{1})

	// haqq_11235-1: no active channel at all -> skipped
	s.seedFlaggedZone("haqq_11235-1", "connection-2", false)

	// evmos_9001-2: deprecated, open channel -> skipped
	s.seedFlaggedZone("evmos_9001-2", "connection-3", true)
	s.mockDelegationChannel("evmos_9001-2", "connection-3", "channel-20")

	v35.ResetStaleDelegationChangesInProgress(s.Ctx, s.App.StakeibcKeeper)

	s.Require().Equal([]uint64{0, 0, 0}, s.flags("cosmoshub-4"), "open channel, no packets: reset")
	s.Require().Equal([]uint64{12, 1, 0}, s.flags("juno-1"), "unacked packet: untouched")
	s.Require().Equal([]uint64{12, 1, 0}, s.flags("haqq_11235-1"), "no channel: untouched")
	s.Require().Equal([]uint64{12, 1, 0}, s.flags("evmos_9001-2"), "deprecated: untouched")
}

func (s *UpgradeTestSuite) TestResetStaleDelegationChangesInProgress_NoZones() {
	s.Require().NotPanics(func() { v35.ResetStaleDelegationChangesInProgress(s.Ctx, s.App.StakeibcKeeper) })
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/TestResetStaleDelegationChangesInProgress' -v`
Expected: build failure `undefined: v35.ResetStaleDelegationChangesInProgress`.

- [ ] **Step 3: Write `stale_flags.go`**

```go
package v35

import (
	"fmt"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// ResetStaleDelegationChangesInProgress zeroes DelegationChangesInProgress on every validator
// of each in-scope zone whose delegation ICA channel is open with no unacked packet: exactly
// what RestoreInterchainAccount does after a channel restore (spec §5 "Stale in-progress
// flags"). A flag with no ICA behind it is stale by definition, since the callback that clears
// it can never fire, and stale flags are what has the Cosmos Hub pipeline stuck. A zone with a
// packet in flight, no open channel, or the Deprecated flag is logged and left alone.
func ResetStaleDelegationChangesInProgress(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if hostZone.Deprecated {
			continue
		}

		owner := stakeibctypes.FormatHostZoneICAOwner(hostZone.ChainId, stakeibctypes.ICAAccountType_DELEGATION)
		portId, err := icatypes.NewControllerPortID(owner)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: %s: unable to build the delegation port id, skipping flag reset: %s", hostZone.ChainId, err))
			continue
		}

		channelId, found := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, hostZone.ConnectionId, portId)
		if !found {
			ctx.Logger().Info(fmt.Sprintf("v35: %s: no open delegation channel, skipping flag reset (restore the channel after the upgrade)", hostZone.ChainId))
			continue
		}
		if unacked := k.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctx, portId, channelId); len(unacked) > 0 {
			ctx.Logger().Info(fmt.Sprintf("v35: %s: %d unacked packet(s) on %s, skipping flag reset", hostZone.ChainId, len(unacked), channelId))
			continue
		}

		// Validators are stored as pointers, so clearing the flag updates hostZone in place
		numReset := 0
		for _, validator := range hostZone.Validators {
			if validator.DelegationChangesInProgress == 0 {
				continue
			}
			ctx.Logger().Info(fmt.Sprintf("v35: %s: resetting %s DelegationChangesInProgress %d -> 0",
				hostZone.ChainId, validator.Address, validator.DelegationChangesInProgress))
			validator.DelegationChangesInProgress = 0
			numReset++
		}
		k.SetHostZone(ctx, hostZone)
		ctx.Logger().Info(fmt.Sprintf("v35: %s: %d stale in-progress flag(s) reset on %s", hostZone.ChainId, numReset, channelId))
	}
}
```

- [ ] **Step 4: Add the call to `upgrades.go`** (after the oracles/rate-limit calls)

```go
		// Stale DelegationChangesInProgress flags on zones with nothing in flight (spec §5)
		ResetStaleDelegationChangesInProgress(ctx, stakeibcKeeper)
```

- [ ] **Step 5: Run the tests**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(ResetStaleDelegationChangesInProgress|Upgrade_EmptyState)' -v`
Expected: all `PASS`.

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35/stale_flags.go app/upgrades/v35/stale_flags_test.go app/upgrades/v35/upgrades.go
git commit -m "feat(v35): reset stale DelegationChangesInProgress on zones with no ICA in flight

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 8: Pending ICQ purges

**Files:**
- Create: `app/upgrades/v35/icq_purge.go`
- Create: `app/upgrades/v35/icq_purge_test.go`
- Modify: `app/upgrades/v35/upgrades.go` (call lines)

**Interfaces:**
- Consumes: `icqkeeper.Keeper.AllQueries/DeleteQuery`, `stakeibckeeper.Keeper.GetHostZone/SetHostZone`, the callback id constants in `x/stakeibc/keeper/icqcallbacks.go`.
- Produces: `PurgeHaqqSlashQueries(ctx, icq icqkeeper.Keeper, sk stakeibckeeper.Keeper)`, `PurgeWithdrawalBalanceQueries(ctx, icq icqkeeper.Keeper)`.
- Depends on: Task 1
- Review: no

- [ ] **Step 1: Write the failing tests**

```go
package v35_test

import (
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func (s *UpgradeTestSuite) seedQueries() {
	for _, query := range []icqtypes.Query{
		{Id: "haqq-delegation", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Delegation},
		{Id: "haqq-validator", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Validator},
		{Id: "haqq-calibrate", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Calibrate},
		{Id: "haqq-withdrawal", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_WithdrawalHostBalance},
		{Id: "haqq-fee", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_FeeBalance},
		{Id: "juno-withdrawal", ChainId: "juno-1", CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_WithdrawalHostBalance},
		{Id: "juno-delegation", ChainId: "juno-1", CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Delegation},
		{Id: "comdex-calibrate", ChainId: "comdex-1", CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Calibrate},
	} {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
	}
}

func (s *UpgradeTestSuite) queryIds() []string {
	ids := []string{}
	for _, query := range s.App.InterchainqueryKeeper.AllQueries(s.Ctx) {
		ids = append(ids, query.Id)
	}
	return ids
}

func (s *UpgradeTestSuite) TestPurgeHaqqSlashQueries() {
	s.seedQueries()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: v35.HaqqChainId,
		Validators: []*stakeibctypes.Validator{
			{Address: "haqqvaloper1", SlashQueryInProgress: true},
			{Address: "haqqvaloper2", SlashQueryInProgress: false},
		},
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:    "juno-1",
		Validators: []*stakeibctypes.Validator{{Address: "junovaloper1", SlashQueryInProgress: true}},
	})

	v35.PurgeHaqqSlashQueries(s.Ctx, s.App.InterchainqueryKeeper, s.App.StakeibcKeeper)

	s.Require().ElementsMatch(
		[]string{"haqq-withdrawal", "haqq-fee", "juno-withdrawal", "juno-delegation", "comdex-calibrate"},
		s.queryIds(), "only haqq's slash-path queries are deleted")

	haqq, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, validator := range haqq.Validators {
		s.Require().False(validator.SlashQueryInProgress, "%s flag cleared", validator.Address)
	}
	juno, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "juno-1")
	s.Require().True(juno.Validators[0].SlashQueryInProgress, "other zones' flags untouched")
}

func (s *UpgradeTestSuite) TestPurgeHaqqSlashQueries_NoZone() {
	s.seedQueries()
	s.Require().NotPanics(func() { v35.PurgeHaqqSlashQueries(s.Ctx, s.App.InterchainqueryKeeper, s.App.StakeibcKeeper) })
	s.Require().Len(s.queryIds(), 5, "queries are still purged when the zone is missing")
}

func (s *UpgradeTestSuite) TestPurgeWithdrawalBalanceQueries() {
	s.seedQueries()

	v35.PurgeWithdrawalBalanceQueries(s.Ctx, s.App.InterchainqueryKeeper)

	s.Require().ElementsMatch(
		[]string{"haqq-delegation", "haqq-validator", "haqq-calibrate", "haqq-fee", "juno-delegation", "comdex-calibrate"},
		s.queryIds(), "only withdrawal-balance queries are deleted, on every chain")
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/TestPurge' -v`
Expected: build failure `undefined: v35.PurgeHaqqSlashQueries`.

- [ ] **Step 3: Write `icq_purge.go`**

```go
package v35

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

// haqqSlashPathCallbacks are the ICQ callbacks that correct a validator's delegation from
// on-chain shares. A query submitted against the pre-delta state has no reason to exist once
// the delta table is applied, so all of them are thrown out first.
var haqqSlashPathCallbacks = map[string]bool{
	stakeibckeeper.ICQCallbackID_Delegation: true,
	stakeibckeeper.ICQCallbackID_Validator:  true,
	stakeibckeeper.ICQCallbackID_Calibrate:  true,
}

// PurgeHaqqSlashQueries deletes every pending slash-path query for haqq_11235-1 and clears
// SlashQueryInProgress on every haqq validator (spec §5 "Pending ICQs"). Unlike v34's pinned
// query ids this is dynamic, so it cannot go stale between measurement and execution. It
// must run before ReconcileHaqqDelegations.
func PurgeHaqqSlashQueries(ctx sdk.Context, icq icqkeeper.Keeper, sk stakeibckeeper.Keeper) {
	numDeleted := 0
	for _, query := range icq.AllQueries(ctx) {
		if query.ChainId != HaqqChainId || !haqqSlashPathCallbacks[query.CallbackId] {
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: deleting pending %s ICQ %s for %s", query.CallbackId, query.Id, HaqqChainId))
		icq.DeleteQuery(ctx, query.Id)
		numDeleted++
	}
	ctx.Logger().Info(fmt.Sprintf("v35: %d pending slash-path ICQ(s) deleted for %s", numDeleted, HaqqChainId))

	hostZone, found := sk.GetHostZone(ctx, HaqqChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v35: host zone %s not found, skipping slash query flag reset", HaqqChainId))
		return
	}
	// Validators are stored as pointers, so clearing the flag updates hostZone in place
	for _, validator := range hostZone.Validators {
		if validator.SlashQueryInProgress {
			ctx.Logger().Info(fmt.Sprintf("v35: clearing SlashQueryInProgress on %s", validator.Address))
			validator.SlashQueryInProgress = false
		}
	}
	sk.SetHostZone(ctx, hostZone)
}

// PurgeWithdrawalBalanceQueries deletes every pending withdrawal-balance query on every
// chain: its callback delegates the withdrawal ICA balance back to validators, and PR 2
// removed the epoch call that submits it (spec §6), so a late answer must not be acted on.
func PurgeWithdrawalBalanceQueries(ctx sdk.Context, icq icqkeeper.Keeper) {
	numDeleted := 0
	for _, query := range icq.AllQueries(ctx) {
		if query.CallbackId != stakeibckeeper.ICQCallbackID_WithdrawalHostBalance {
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: deleting pending withdrawal-balance ICQ %s for %s", query.Id, query.ChainId))
		icq.DeleteQuery(ctx, query.Id)
		numDeleted++
	}
	ctx.Logger().Info(fmt.Sprintf("v35: %d pending withdrawal-balance ICQ(s) deleted", numDeleted))
}
```

- [ ] **Step 4: Add the calls to `upgrades.go`** (after the stale-flag call; the haqq purge must precede the delta table, which Task 9 adds after these lines)

```go
		// Pending ICQs (spec §5): the haqq slash-path purge runs before the haqq delta table
		PurgeHaqqSlashQueries(ctx, icqKeeper, stakeibcKeeper)
		PurgeWithdrawalBalanceQueries(ctx, icqKeeper)
```

- [ ] **Step 5: Run the tests**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(Purge|Upgrade_EmptyState)' -v`
Expected: all `PASS`.

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35/icq_purge.go app/upgrades/v35/icq_purge_test.go app/upgrades/v35/upgrades.go
git commit -m "feat(v35): purge haqq slash-path ICQs and every withdrawal-balance ICQ

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

### Task 9: Haqq delta table, its generator, and the reconciliation

**Files:**
- Create: `scripts/wind-down/gen_delta_table.py`
- Create: `scripts/wind-down/test_gen_delta_table.py`
- Create: `app/upgrades/v35/haqq.go`
- Create: `app/upgrades/v35/haqq_test.go`
- Modify: `app/upgrades/v35/upgrades.go` (call line)
- Modify: `scripts/wind-down/measure_delegation_drift.py` (docstring lines 5-6, `SCRIPT_DIR` line 17, `main` lines 492-563)
- Modify: `.gitignore` (add `scripts/wind-down/drift/`)

**Interfaces:**
- Consumes: Task 2's `DelegationDelta`, `mustInt`, `reconcileHostZoneDelegations`; `scripts/wind-down/measure_delegation_drift.py`'s `drift.json` (`zones.<chain_id>.validators[]` rows with `validator_address`, `moniker`, `diff` = recorded − actual, `in_stride_list`).
- Produces: `HaqqDelegationDeltas []DelegationDelta`, `ReconcileHaqqDelegations(ctx, sk stakeibckeeper.Keeper) (appliedDelta sdkmath.Int, applied bool)`.
- Depends on: Tasks 1, 2
- Review: yes (money: rewrites tracked delegations on a live zone)

The drift measurement was rerun for this plan on 2026-09-29 at 21:36 UTC (REST `haqq-rest.publicnode.com`, 52 Stride validators, 42 with an on-chain delegation): 16 non-zero deltas, 13 over-recorded and 3 under, net **−1,758,262,481,209,726,849,963 aISLM** (−1,758.26 ISLM), unchanged from the four earlier measurements in the spec (§5, §9a). The table below is that run's output through the generator. **It is regenerated right before the proposal** (spec §9 checklist) with the commands in Step 7, and the v35 mainnet-export suite (PR 6) is what turns a stale table into a red build.

- [ ] **Step 1: Make the drift script reproducible (`--output-dir`, `--chain-id`)**

The table is a release-critical accounting input, so the measurement must run from a clean checkout with one exact command. Today `scripts/wind-down/measure_delegation_drift.py` writes into a hard-coded, session-private temp directory and always measures every zone. Edit it as follows (there are no tests for this script; Step 7 is the verification).

Docstring, lines 5-6, replace

```python
Read-only. Writes drift.json and report.md into this directory.
```

with

```python
Read-only. Writes drift.json and report.md into --output-dir (default scripts/wind-down/drift/,
git-ignored). --chain-id (repeatable) restricts the measurement to those Stride host zones.
```

Imports, lines 9-13: add `import argparse` and `import pathlib` (keep the module-qualified style; alphabetical order: `argparse`, `json`, `pathlib`, `time`, `urllib.request`, `urllib.error`).

Delete line 17 entirely:

```python
SCRIPT_DIR = "/private/tmp/claude-501/-Users-sampocs-Documents-Projects-stride/4ac0fcca-c8bc-4981-b63e-b0f199cc9eaa/scratchpad/drift"
```

Replace the head of `main` (line 492 to the `for` at line 500)

```python
def main():
    stride_zones = fetch_stride_host_zones()
    stride_by_chain_id = {z["chain_id"]: z for z in stride_zones}

    all_zone_data = {}
    all_rows = {}
    all_summaries = []

    for chain_id, cfg in ZONES.items():
```

with

```python
def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--output-dir",
        type=pathlib.Path,
        default=pathlib.Path("scripts/wind-down/drift"),
        help="directory for drift.json and report.md (created if missing)",
    )
    parser.add_argument(
        "--chain-id",
        action="append",
        dest="chain_ids",
        default=None,
        help="measure only this Stride host zone (repeatable); default is every zone in ZONES",
    )
    return parser.parse_args()


def selected_zones(chain_ids: list[str] | None) -> dict[str, dict]:
    if chain_ids is None:
        return ZONES
    unknown = sorted(set(chain_ids) - set(ZONES))
    if unknown:
        raise SystemExit(f"unknown chain id(s): {', '.join(unknown)}; known: {', '.join(ZONES)}")
    return {chain_id: ZONES[chain_id] for chain_id in chain_ids}


def main() -> None:
    args = parse_args()
    zones = selected_zones(chain_ids=args.chain_ids)
    output_dir: pathlib.Path = args.output_dir
    output_dir.mkdir(parents=True, exist_ok=True)

    stride_zones = fetch_stride_host_zones()
    stride_by_chain_id = {z["chain_id"]: z for z in stride_zones}

    all_zone_data = {}
    all_rows = {}
    all_summaries = []

    for chain_id, cfg in zones.items():
```

Then, still inside `main`, the three remaining references to the full zone table become the selection: line 520 `for chain_id in ZONES:` → `for chain_id in zones:`; line 538 `decimals_map = {chain_id: cfg["decimals"] for chain_id, cfg in ZONES.items()}` → `... for chain_id, cfg in zones.items()}`; line 548 `for chain_id in ZONES:` → `for chain_id in zones:`. The two writes and the final print:

```python
    with open(f"{SCRIPT_DIR}/drift.json", "w") as f:
```
→
```python
    with open(output_dir / "drift.json", "w") as f:
```

```python
    with open(f"{SCRIPT_DIR}/report.md", "w") as f:
```
→
```python
    with open(output_dir / "report.md", "w") as f:
```

```python
    print("\nDone. Wrote drift.json and report.md")
```
→
```python
    print(f"\nDone. Wrote {output_dir / 'drift.json'} and {output_dir / 'report.md'}")
```

Append to `.gitignore` (the existing `scripts/state` and `scripts/logs` entries do not cover it):

```
scripts/wind-down/drift/
```

Check: `grep -n SCRIPT_DIR scripts/wind-down/measure_delegation_drift.py` prints nothing; `python3 scripts/wind-down/measure_delegation_drift.py --help` prints both options; `python3 scripts/wind-down/measure_delegation_drift.py --chain-id nope` exits with `unknown chain id(s): nope; known: ...`.

- [ ] **Step 2: Write the generator's failing test**

```python
# scripts/wind-down/test_gen_delta_table.py
import json
import pathlib
import tempfile
import unittest

import gen_delta_table


def write_drift(rows: list[dict]) -> pathlib.Path:
    payload = {"generated_at": "2026-09-29T00:00:00Z", "zones": {"haqq_11235-1": {"validators": rows}}}
    path = pathlib.Path(tempfile.mkdtemp()) / "drift.json"
    path.write_text(json.dumps(payload))
    return path


class GenDeltaTableTest(unittest.TestCase):
    def test_emits_one_entry_per_nonzero_delta_sorted_by_magnitude(self) -> None:
        drift_path = write_drift(
            [
                {"validator_address": "haqqvaloper1small", "moniker": "Small Node", "diff": -5, "in_stride_list": True},
                {"validator_address": "haqqvaloper1big", "moniker": "Big [Node] ⚡", "diff": 853, "in_stride_list": True},
                {"validator_address": "haqqvaloper1zero", "moniker": "Zero", "diff": 0, "in_stride_list": True},
                {"validator_address": "haqqvaloper1foreign", "moniker": "Foreign", "diff": 7, "in_stride_list": False},
            ]
        )

        entries = gen_delta_table.load_deltas(drift_path=drift_path, chain_id="haqq_11235-1", names={})

        self.assertEqual(
            [("bignode", "haqqvaloper1big", -853), ("smallnode", "haqqvaloper1small", 5)],
            [(entry.name, entry.address, entry.delta) for entry in entries],
        )

    def test_names_come_from_the_host_zone_when_given(self) -> None:
        drift_path = write_drift([{"validator_address": "haqqvaloper1big", "moniker": "Big", "diff": 1, "in_stride_list": True}])

        entries = gen_delta_table.load_deltas(drift_path=drift_path, chain_id="haqq_11235-1", names={"haqqvaloper1big": "bigstride"})

        self.assertEqual("bigstride", entries[0].name)

    def test_render_go_literal(self) -> None:
        rendered = gen_delta_table.render_go(
            entries=[gen_delta_table.DeltaEntry(name="big", address="haqqvaloper1big", delta=-853)],
            var_name="HaqqDelegationDeltas",
        )

        self.assertEqual(
            'var HaqqDelegationDeltas = []DelegationDelta{\n'
            '\t{Name: "big", Address: "haqqvaloper1big", Delta: mustInt("-853")},\n'
            '}\n',
            rendered,
        )


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 3: Run to verify it fails**

Run: `cd scripts/wind-down && python3 -m unittest test_gen_delta_table -v`
Expected: `ModuleNotFoundError: No module named 'gen_delta_table'`.

- [ ] **Step 4: Write `scripts/wind-down/gen_delta_table.py`**

```python
#!/usr/bin/env python3
"""
Turn measure_delegation_drift.py's drift.json into the Go delta table an upgrade handler applies
through reconcileHostZoneDelegations (app/upgrades/v35/delegation_deltas.go).

drift.json stores diff = recorded - actual; the Go table stores Delta = actual - recorded, so
every sign flips here. Zero diffs and validators Stride does not track are left out.

Usage:
    python3 gen_delta_table.py DRIFT_JSON CHAIN_ID [--host-zone-json HOST_ZONE_JSON] [--var-name NAME]

HOST_ZONE_JSON is the REST response of /Stride-Labs/stride/stakeibc/host_zone/CHAIN_ID; it maps
validator addresses to Stride's validator names for the Name column. Without it the host
moniker (lower-cased, alphanumerics only) is used.
"""

import argparse
import json
import pathlib
import re
from dataclasses import dataclass


@dataclass(frozen=True)
class DeltaEntry:
    name: str
    address: str
    delta: int  # actual on-chain minus tracked, in the host base denom


def load_deltas(drift_path: pathlib.Path, chain_id: str, names: dict[str, str]) -> list[DeltaEntry]:
    drift = json.loads(drift_path.read_text())
    rows = drift["zones"][chain_id]["validators"]

    entries = [
        DeltaEntry(
            name=names.get(row["validator_address"], sanitize_name(row.get("moniker") or "")),
            address=row["validator_address"],
            delta=-int(row["diff"]),
        )
        for row in rows
        if row["in_stride_list"] and int(row["diff"]) != 0
    ]
    return sorted(entries, key=lambda entry: -abs(entry.delta))


def load_host_zone_names(host_zone_path: pathlib.Path | None) -> dict[str, str]:
    if host_zone_path is None:
        return {}
    host_zone = json.loads(host_zone_path.read_text())["host_zone"]
    return {validator["address"]: validator["name"] for validator in host_zone["validators"]}


def sanitize_name(moniker: str) -> str:
    return re.sub(r"[^a-z0-9]", "", moniker.lower())


def render_go(entries: list[DeltaEntry], var_name: str) -> str:
    lines = [f"var {var_name} = []DelegationDelta{{"]
    lines += [f'\t{{Name: "{entry.name}", Address: "{entry.address}", Delta: mustInt("{entry.delta}")}},' for entry in entries]
    lines.append("}")
    return "\n".join(lines) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("drift_json", type=pathlib.Path)
    parser.add_argument("chain_id")
    parser.add_argument("--host-zone-json", type=pathlib.Path, default=None)
    parser.add_argument("--var-name", default="HaqqDelegationDeltas")
    args = parser.parse_args()

    names = load_host_zone_names(host_zone_path=args.host_zone_json)
    entries = load_deltas(drift_path=args.drift_json, chain_id=args.chain_id, names=names)
    net = sum(entry.delta for entry in entries)

    print(f"// {len(entries)} deltas, net {net} (actual minus tracked)")
    print(render_go(entries=entries, var_name=args.var_name), end="")


if __name__ == "__main__":
    main()
```

- [ ] **Step 5: Run the generator tests**

Run: `cd scripts/wind-down && python3 -m unittest test_gen_delta_table -v`
Expected: `Ran 3 tests ... OK`.

- [ ] **Step 6: Write the failing Go tests**

```go
package v35_test

import (
	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Seeds the haqq host zone with every validator in the delta table, each tracked at a round
// number of whole ISLM so the expected post-reconciliation values are obvious. The full table
// is needed because the reconciliation refuses to apply a partial one.
func (s *UpgradeTestSuite) setupHaqqHostZone() (tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int) {
	tracked = map[string]sdkmath.Int{}
	trackedTotal = sdkmath.ZeroInt()

	validators := []*stakeibctypes.Validator{}
	for i, entry := range v35.HaqqDelegationDeltas {
		delegation := sdkmath.NewInt(int64(1000 + i)).Mul(sdkmath.NewInt(1e18))
		validators = append(validators, &stakeibctypes.Validator{
			Name:       entry.Name,
			Address:    entry.Address,
			Delegation: delegation,
		})
		tracked[entry.Address] = delegation
		trackedTotal = trackedTotal.Add(delegation)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v35.HaqqChainId,
		HostDenom:        "aISLM",
		TotalDelegations: trackedTotal,
		Validators:       validators,
	})
	return tracked, trackedTotal
}

func (s *UpgradeTestSuite) TestHaqqDelegationDeltas_TableShape() {
	seen := map[string]bool{}
	hasPositive, hasNegative := false, false
	net := sdkmath.ZeroInt()
	for _, entry := range v35.HaqqDelegationDeltas {
		s.Require().False(seen[entry.Address], "duplicate address %s", entry.Address)
		seen[entry.Address] = true
		s.Require().False(entry.Delta.IsZero(), "%s has a zero delta", entry.Name)
		s.Require().Contains(entry.Address, "haqqvaloper1", "%s is not a haqq operator address", entry.Address)
		hasPositive = hasPositive || entry.Delta.IsPositive()
		hasNegative = hasNegative || entry.Delta.IsNegative()
		net = net.Add(entry.Delta)
	}
	s.Require().True(hasPositive && hasNegative, "both signs are applied (spec §5)")
	s.Require().True(net.IsNegative(), "the 2026-09-29 table nets to an over-recording, so TotalDelegations must drop")
}

func (s *UpgradeTestSuite) TestReconcileHaqqDelegations() {
	tracked, trackedTotal := s.setupHaqqHostZone()

	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().True(applied)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().True(found)

	expectedDelta := sdkmath.ZeroInt()
	for _, entry := range v35.HaqqDelegationDeltas {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().True(found, "validator %s should still exist", entry.Name)
		s.Require().Equal(tracked[entry.Address].Add(entry.Delta), validator.Delegation, "%s delegation moves by its delta", entry.Name)
		expectedDelta = expectedDelta.Add(entry.Delta)
	}
	s.Require().Equal(expectedDelta, appliedDelta, "returned delta is the sum over the whole table")
	s.Require().Equal(trackedTotal.Add(appliedDelta), hostZone.TotalDelegations, "TotalDelegations adjusted by exactly the net")
	s.Require().True(hostZone.TotalDelegations.LT(trackedTotal), "TotalDelegations drops")

	sum := sdkmath.ZeroInt()
	for _, validator := range hostZone.Validators {
		sum = sum.Add(validator.Delegation)
	}
	s.Require().Equal(sum, hostZone.TotalDelegations, "TotalDelegations == sum(validator.Delegation)")
}

// Dropping one entry from the host zone must skip the whole reconciliation
func (s *UpgradeTestSuite) TestReconcileHaqqDelegations_MissingValidatorSkipsAll() {
	tracked, trackedTotal := s.setupHaqqHostZone()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	hostZone.Validators = hostZone.Validators[1:]
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().False(applied)
	s.Require().True(appliedDelta.IsZero())

	after, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().Equal(trackedTotal, after.TotalDelegations, "nothing written")
	for _, validator := range after.Validators {
		s.Require().Equal(tracked[validator.Address], validator.Delegation, "%s untouched", validator.Name)
	}
}

func (s *UpgradeTestSuite) TestReconcileHaqqDelegations_MissingZone() {
	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().False(applied)
	s.Require().True(appliedDelta.IsZero())
}
```

- [ ] **Step 7: Regenerate the table (record the commands; the values below are today's)**

```bash
# From the repo root, after Step 1. Both outputs land in the git-ignored scripts/wind-down/drift/.
python3 scripts/wind-down/measure_delegation_drift.py --chain-id haqq_11235-1 --output-dir scripts/wind-down/drift
curl -s -A "Mozilla/5.0" "https://stride-api.polkachu.com/Stride-Labs/stride/stakeibc/host_zone/haqq_11235-1" > scripts/wind-down/drift/haqq_host_zone.json
python3 scripts/wind-down/gen_delta_table.py scripts/wind-down/drift/drift.json haqq_11235-1 --host-zone-json scripts/wind-down/drift/haqq_host_zone.json
```

This is also the verification of Step 1: the run must complete from a clean checkout with no edits, `git status` must show nothing under `scripts/wind-down/drift/`, and the generator's output must match the table committed in `haqq.go` line for line (`diff <(python3 scripts/wind-down/gen_delta_table.py scripts/wind-down/drift/drift.json haqq_11235-1 --host-zone-json scripts/wind-down/drift/haqq_host_zone.json) <(sed -n '/^var HaqqDelegationDeltas/,/^}/p' app/upgrades/v35/haqq.go)` after the table is in place; a non-empty diff means the chain moved since the last measurement and the table is re-pasted, never hand-edited).

Expected (2026-09-29 21:36 UTC): the first line `// 16 deltas, net -1758262481209726849963 (actual minus tracked)` and the table pasted in Step 8.

- [ ] **Step 8: Write `haqq.go`**

```go
package v35

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

// HaqqDelegationDeltas trues up the haqq_11235-1 host zone's tracked delegations to what is
// actually staked on Haqq (spec §5 "Haqq delegation reconciliation", §9a).
//
// Thirteen validators are over-recorded (undetected downtime slashes and sub-token rounding,
// the largest 853.8 ISLM) and three are under-recorded by sub-token dust. Both signs are
// applied so every tracked delegation equals the chain's; the net is a decrease of about
// 1,758 ISLM, so TotalDelegations drops. The rate update was deleted in PR 2, so this no
// longer reaches the stISLM redemption rate. The stored SharesToTokensRate is deliberately
// left as is: the day-0 refresh updates it, and the slash callback then finds tracked
// delegation equal to on-chain shares × the refreshed rate, so nothing is applied twice.
//
// Generated 2026-09-29 21:36 UTC by scripts/wind-down/gen_delta_table.py from
// measure_delegation_drift.py (Haqq REST haqq-rest.publicnode.com; 52 tracked validators).
// The net has been stable since 2026-09-21. Regenerate right before the upgrade proposal;
// the mainnet-export suite (PR 6) fails if this table no longer matches state.
var HaqqDelegationDeltas = []DelegationDelta{
	{Name: "neuler", Address: "haqqvaloper1a57vprf7lswm3aqy2g5gy509235wmtsfvf9q73", Delta: mustInt("-853800913870811785762")},
	{Name: "kioqqhaqqsh", Address: "haqqvaloper1a4qnqnk5ag0um6z3unkdth92v9c46x0dcpe2n0", Delta: mustInt("-448303270428234546889")},
	{Name: "gmocoin", Address: "haqqvaloper1f2j8t0ddtak6z9td28wmv60xj6mlykzx65jr8w", Delta: mustInt("-364790807431421779754")},
	{Name: "takamulfivalidator", Address: "haqqvaloper1gcw6ru5akcmzvr3anqre8f3m8eml2rmrz8m6sl", Delta: mustInt("-91337544757772840484")},
	{Name: "stakingcabin", Address: "haqqvaloper1f6hy5d68hkx9wtp8er3mgdfwjfjr7tun5yu7us", Delta: mustInt("-29944721483185574")},
	{Name: "islamicstaking", Address: "haqqvaloper1p8k6xk94u24vv9dmxu3vkgg43fs3v72grkpjhm", Delta: mustInt("-3216140")},
	{Name: "foreststaking", Address: "haqqvaloper1ktu3f367c0j0yet6apuefy8xt5n7eswns7yuff", Delta: mustInt("677576")},
	{Name: "surestake", Address: "haqqvaloper16hy887wxzjmmkkfrdxzgz9dlv6mfru56q539cw", Delta: mustInt("-203557")},
	{Name: "masterblox", Address: "haqqvaloper1ja29wpj6l5t42jj67vgqcuj8pu046uqklss524", Delta: mustInt("31310")},
	{Name: "stakeme", Address: "haqqvaloper1p02zk5ecdanap637e2wtt82cucjlxtkrhus623", Delta: mustInt("-1234")},
	{Name: "haqqassociation", Address: "haqqvaloper16lp0xpq87cre5z4jkfddq78r5l4vcd7el2jlmj", Delta: mustInt("555")},
	{Name: "noders", Address: "haqqvaloper1hgggrfgjeu4d5nveh03c6w37magsuqcy84p44t", Delta: mustInt("-4")},
	{Name: "staketake", Address: "haqqvaloper1tnm7y48w5nh8wt0s2u0fxwu607xtqhk6v99773", Delta: mustInt("-2")},
	{Name: "alxvoyanodeteam", Address: "haqqvaloper1wgm35c4nzs6ssgktd0zj4pefdr3h8ms5252mqc", Delta: mustInt("-2")},
	{Name: "segastakers", Address: "haqqvaloper10jqmd8rvggegva0r5smarr7avwwa8vce0q5gsh", Delta: mustInt("-1")},
	{Name: "palamar", Address: "haqqvaloper1xp597fjhgu6dx3a525htulkn36fqqntjaqvhct", Delta: mustInt("-1")},
}

// ReconcileHaqqDelegations applies HaqqDelegationDeltas to the haqq_11235-1 host zone so that
// tracked validator delegations (and TotalDelegations) match what is staked on Haqq. It
// returns the net delta applied and whether the table was applied at all; the table is
// applied all-or-nothing and never as an upgrade error (see reconcileHostZoneDelegations).
// Nothing is queued afterwards: the wind-down drains every delegation by admin tx.
func ReconcileHaqqDelegations(ctx sdk.Context, sk stakeibckeeper.Keeper) (appliedDelta sdkmath.Int, applied bool) {
	return reconcileHostZoneDelegations(ctx, sk, HaqqChainId, HaqqDelegationDeltas)
}
```

- [ ] **Step 9: Add the call to `upgrades.go`** (last, after the ICQ purges)

```go
		// Haqq delegation reconciliation, after its slash-path ICQs are gone (spec §5)
		ReconcileHaqqDelegations(ctx, stakeibcKeeper)
```

- [ ] **Step 10: Run the tests**

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/Test(HaqqDelegationDeltas|ReconcileHaqqDelegations|Upgrade_EmptyState)' -v`
Expected: all `PASS`.

- [ ] **Step 11: Commit**

```bash
git add scripts/wind-down/measure_delegation_drift.py .gitignore scripts/wind-down/gen_delta_table.py scripts/wind-down/test_gen_delta_table.py app/upgrades/v35/haqq.go app/upgrades/v35/haqq_test.go app/upgrades/v35/upgrades.go
git commit -m "feat(v35): haqq delegation delta table, its generator, and the reconciliation; drift script takes --output-dir/--chain-id

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Post-wave integration (serial, after Tasks 3 to 9 are merged)

### Task 10: Handler order, full-handler test, export README

**Files:**
- Modify: `app/upgrades/v35/upgrades.go` (verify the order and the doc comment)
- Modify: `app/upgrades/v35/upgrades_test.go` (add `TestUpgrade`)
- Create: `app/upgrades/v35/testdata/README.md`

**Interfaces:**
- Consumes: every helper from Tasks 3 to 9 and their test fixtures (`storeAndInstantiateHackatom`, `seedFlaggedZone`, `mockDelegationChannel`, `flags`, `seedQueries`, `queryIds`, `setupHaqqHostZone`).
- Depends on: Tasks 3, 4, 5, 6, 7, 8, 9 (post-wave; run this task alone after the merges)
- Review: yes (this is where the order and the one error path are verified end to end)

- [ ] **Step 1: Verify the handler body reads exactly as follows** (fix the merge if it does not; the anchor comment may be deleted now)

```go
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Entry points that bypass the msg service router (spec §5)
		DisableAutopilotStakeibc(ctx, autopilotKeeper)
		RemoveStakeibcFromICAHostAllowList(ctx, icaHostKeeper)

		// Wasm control to gov: the upload-access write is the one step that may fail the upgrade
		if err := SetWasmUploadAccessToGov(ctx, wasmKeeper); err != nil {
			return vm, err
		}
		MoveDeployKeyContractAdminsToGov(ctx, wasmKeeper)

		// Stakeibc state flips (spec §5)
		DeprecateComdex(ctx, stakeibcKeeper)
		DeleteDydxTradeRoute(ctx, stakeibcKeeper)

		// Oracles and rate limits (spec §5)
		DeactivateICAOracles(ctx, icaOracleKeeper)
		RemoveAllRateLimits(ctx, ratelimitKeeper)

		// Stale DelegationChangesInProgress flags on zones with nothing in flight (spec §5)
		ResetStaleDelegationChangesInProgress(ctx, stakeibcKeeper)

		// Pending ICQs (spec §5): the haqq slash-path purge runs before the haqq delta table
		PurgeHaqqSlashQueries(ctx, icqKeeper, stakeibcKeeper)
		PurgeWithdrawalBalanceQueries(ctx, icqKeeper)

		// Haqq delegation reconciliation, after its slash-path ICQs are gone (spec §5)
		ReconcileHaqqDelegations(ctx, stakeibcKeeper)

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
```

Update the `CreateUpgradeHandler` doc comment to list the steps in this order in one sentence each.

- [ ] **Step 2: Add the full-handler test to `upgrades_test.go`**

```go
// TestUpgrade runs the whole handler through the upgrade module on a state that exercises
// every step at once, and asserts the state each helper's own test checks in isolation.
func (s *UpgradeTestSuite) TestUpgrade() {
	// ----- arrange -----
	s.App.AutopilotKeeper.SetParams(s.Ctx, autopilottypes.Params{StakeibcActive: true, ClaimActive: true})
	s.App.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.Params{HostEnabled: true, AllowMessages: []string{
		"/cosmos.bank.v1beta1.MsgSend",
		sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
	}})
	deployKeyContract := s.storeAndInstantiateHackatom(sdk.MustAccAddressFromBech32(v35.WasmDeployKey))
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: v35.ComdexChainId})
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, stakeibctypes.TradeRoute{
		RewardDenomOnRewardZone: v35.DydxTradeRouteRewardDenom, HostDenomOnHostZone: v35.DydxTradeRouteHostDenom,
	})
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, icaoracletypes.Oracle{ChainId: "osmosis-1", ConnectionId: "connection-9", Active: true})
	s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "stuevmos")
	s.seedFlaggedZone("cosmoshub-4", "connection-0", false)
	s.mockDelegationChannel("cosmoshub-4", "connection-0", "channel-863")
	s.seedQueries()
	haqqTracked, haqqTrackedTotal := s.setupHaqqHostZone()
	haqqZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	haqqZone.Validators[0].SlashQueryInProgress = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haqqZone)

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)

	// ----- assert -----
	s.Require().False(s.App.AutopilotKeeper.GetParams(s.Ctx).StakeibcActive, "autopilot")
	s.Require().Equal([]string{"/cosmos.bank.v1beta1.MsgSend", "/stride.stakeibc.MsgClaimUndelegatedTokens"},
		s.App.ICAHostKeeper.GetParams(s.Ctx).AllowMessages, "ICA host allow-list")
	s.Require().Equal([]string{v35.GovModuleAddress().String()}, s.App.WasmKeeper.GetParams(s.Ctx).CodeUploadAccess.Addresses, "wasm upload")
	s.Require().Equal(v35.GovModuleAddress().String(), s.App.WasmKeeper.GetContractInfo(s.Ctx, deployKeyContract).Admin, "contract admin")
	comdex, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(comdex.Deprecated, "comdex")
	s.Require().Empty(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), "trade route")
	oracle, _ := s.App.ICAOracleKeeper.GetOracle(s.Ctx, "osmosis-1")
	s.Require().False(oracle.Active, "oracle")
	s.Require().Empty(s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx), "rate limiter")
	s.Require().Equal([]uint64{0, 0, 0}, s.flags("cosmoshub-4"), "stale flags")
	s.Require().ElementsMatch([]string{"haqq-fee", "juno-delegation", "comdex-calibrate"}, s.queryIds(), "both ICQ purges")
	haqq, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().False(haqq.Validators[0].SlashQueryInProgress, "haqq slash flag")
	for _, entry := range v35.HaqqDelegationDeltas {
		validator, _, _ := stakeibckeeper.GetValidatorFromAddress(haqq.Validators, entry.Address)
		s.Require().Equal(haqqTracked[entry.Address].Add(entry.Delta), validator.Delegation, "haqq delta %s", entry.Name)
	}
	s.Require().True(haqq.TotalDelegations.LT(haqqTrackedTotal), "haqq total dropped")
}
```

Add the imports the test needs to `upgrades_test.go`: `sdk "github.com/cosmos/cosmos-sdk/types"`, `icahosttypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/types"`, `autopilottypes`, `icaoracletypes`, `stakeibckeeper`, `stakeibctypes` (module paths as in the earlier tasks).

- [ ] **Step 3: Run the whole package**

Run: `go test ./app/upgrades/v35/... -v`
Expected: every test `PASS`, including `TestUpgrade`.

- [ ] **Step 4: Write `app/upgrades/v35/testdata/README.md`**

```markdown
# v35 mainnet export fixture

`mainnet_export.json.gz` is not committed by this PR. The release gate (PR 6) adds the
mainnet-export suite that replays the v35 handler against real state with the real
constants; it skips when the file is absent. This README fixes what that fixture must
contain so it can be assembled from the public REST API right before the proposal, the
way v34's was (see `app/upgrades/v34/testdata/README.md` for the assembly commands and
the trimming rules).

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
```

- [ ] **Step 5: Commit**

```bash
git add app/upgrades/v35/upgrades.go app/upgrades/v35/upgrades_test.go app/upgrades/v35/testdata/README.md
git commit -m "feat(v35): full-handler test and the export fixture contract for the release gate

Co-Authored-By: Claude Fable 5.1 <noreply@anthropic.com>"
```

---

## Ops step after the PR: localstride dry run

Not a code task; recorded here because the plan is where the next engineer looks. `make upgrade-localstride` (in-place testnet with `--trigger-testnet-upgrade`) needs two binaries: the tagged previous release drops the triggered plan (its `InitStrideAppForTestnet` writes to a discarded CheckTx cache; main has the `NewUncachedContext` fix), and a binary that already registers the `v35` handler panics `BINARY UPDATED BEFORE TRIGGER` at the plan height. So: build the previous release from `git archive` into the scratchpad with only the `NewUncachedContext` line patched into `InitStrideAppForTestnet` (no v35 handler), testnetify with it until the log shows `UPGRADE "v35" NEEDED`, then swap to this branch's binary. Sync with the plain previous-release binary first. Never `pkill -f "strided start"` from an agent shell (the pattern matches the shell itself); use `pkill -x strided`. Rehearsal write-ups go in `integration-tests/rehearsal/` on a `REHEARSAL ONLY` branch; `localstride/scratch` and `exports` are gitignored. The dry run checks, in the node log, one `v35:` line per helper in the order of Task 10 Step 1, and afterwards `strided q stakeibc host-zone haqq_11235-1` showing every validator's `delegation_changes_in_progress` at 0 and the delta table's validators at their reconciled amounts.

---

## Self-review

**Spec coverage (§5, §11 handler bullet):** autopilot param (Task 3), ICA host allow-list (3), wasm params and contract admins with a real contract (4), comdex flag (5), trade-route deletion (5), oracle deactivation (6), rate-limit removal (6), stale-flag reset cleared on a zone with no unacked packet and left alone on one that does (7), haqq purge deleting only that chain's slash-path queries and clearing the flags, withdrawal-balance purge leaving every other query (8), haqq delta table applied and skipped on a stale constant, both signs, generated from the drift script (9), handler order with the purge before the table and the single error path (10), the fixture contract for the PR 6 export suite (10). The "two ValidateBasic gates and the lifted calibration cap" and the "removed messages no longer exist" tests in §11 belong to PRs 2 and 1. No store upgrade is needed.

**Placeholders:** none; every step carries its code, every test its body, the haqq table its measured values with an exact regeneration command that needs no edits to the repo (Task 9 Step 1 gives the drift script `--output-dir` and `--chain-id`; the old session-private `SCRIPT_DIR` is gone).

**Type consistency:** `CreateUpgradeHandler` parameters are fixed in Task 1 and only consumed afterwards; helper names match between the source, the tests and the Task 10 handler body; `reconcileHostZoneDelegations` returns `(sdkmath.Int, bool)` as in v34 and `ReconcileHaqqDelegations` passes that through (v34's Injective wrapper dropped the bool; haqq's keeps it, which its tests use). Test fixtures reused by Task 10 (`storeAndInstantiateHackatom`, `seedFlaggedZone`, `mockDelegationChannel`, `flags`, `seedQueries`, `queryIds`, `setupHaqqHostZone`) are defined once in their own task's test file and shared through the package.

**Review tags:** wasm (4), stale flags (7), haqq deltas (9) and the integration (10) are `yes`; the rest are parameter flips with no error path.

**Known risks to surface if hit:** the wasm test depends on the test app executing wasm (`Create`/`Instantiate` through the gov permission keeper); a wasmvm failure there is a finding for the reviewer, not something to route around.
