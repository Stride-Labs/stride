# Wind-Down Upgrade 1 (v35): Close the Doors — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the v35 binary that removes every liquid-stake, redeem and create-things message, closes the autopilot and ICA-host entry points, hands wasm control to gov, deprecates Comdex, deletes the dYdX trade route, and trues up Haqq's validator delegations, while every other flow keeps running.

**Architecture:** One SDK upgrade package `app/upgrades/v35` (handler + constants + tests, v34 style), message removals done per module by deleting the `rpc` from `tx.proto`, the msg-server handler, the codec registration and the CLI command while keeping the proto `message` definitions and any keeper logic that internal callers still use. The stakeibc `LiquidStake` handler body moves to a keeper method because the reward collector and autopilot call it. Haqq reconciliation reuses the v34 delta helper verbatim.

**Tech Stack:** Go 1.2x, cosmos-sdk v0.54.3, ibc-go v11.2.0, wasmd v0.70.2, gogoproto via `make proto-gen` (docker), testify suites via `app/apptesting`.

Spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` (§5 is this upgrade; the spec now describes a two-upgrade wind-down and upgrade 2 has its own plan, `docs/superpowers/plans/2026-09-24-wind-down-upgrade-2.md`). The module path bump to `/v35` is done manually after this plan, not by it.

## Global Constraints

- The Go module path stays `github.com/Stride-Labs/stride/v34` for every file this plan touches; the bump to `/v35` is a manual step the team runs after the plan lands and is deliberately not a task here. The upgrade package is still `app/upgrades/v35` and the plan name is `"v35"`; those are independent of the module path.
- Only `rpc` lines are removed from `.proto` files. Proto `message` definitions stay (some are used internally, and keeping them avoids churn). After any `.proto` edit run `make proto-gen` and commit only the regenerated `tx.pb.go` of the module you changed. With the module path untouched proto-gen should leave every other `*.pb.go` alone; if it still rewrites descriptor bytes elsewhere (a protoc image drift), revert that noise with `git checkout <base> -- <files>` before committing and do not chase it.
- Removed message types STAY registered in the interface registry: keep every `registry.RegisterImplementations((*sdk.Msg)(nil), &Msg...{})` entry in each module's `RegisterInterfaces`, and remove only the `rpc`, the msg-server handler, the `legacy.RegisterAminoMsg` line and the CLI command. Without the registration a v35 node can no longer decode any historical tx containing the message (`strided q tx`, `/cosmos/tx/v1beta1/txs/{hash}` fail with "unable to resolve type URL"). New submissions are still rejected because `MsgServiceRouter().Handler(msg)` is nil ("can't route message"). Add a comment above each block saying so.
- Keeper logic stays wherever something still calls it (reward collector, autopilot, epoch hooks). Delete keeper code only when nothing references it after the handler is gone.
- Upgrade handler helpers never return an error for a missing-state case; they log and continue (v34 convention). Only `RunMigrations` errors propagate.
- macOS host: use `sed -i ''` (BSD sed).
- Every commit message ends with the attribution lines from the session's system reminder.
- Do not push. Branch: `wind-down-design` (already checked out; branch from it per task in worktrees as the sub-skill directs).
- Run `go build ./...` before every commit; run the named package tests in each task.

### Notes from a dry run of this plan (2026-09-21, reverted)

The plan was executed once end to end on a scratch branch and reverted; these are the facts it surfaced, already folded into the tasks above:
- Trade-route store key is `RewardDenomOnRewardZone + "-" + HostDenomOnHostZone`; mainnet's route is `uusdc`/`adydx`.
- The Haqq measurement yields 17 deltas (14 negative slashes, 3 positive sub-token dust); both signs apply; the table matched live state on two measurements 3.5 h apart and again a day later.
- `RegisterHostZone` tests are all keeper tests behind a pure-delegate handler: rewrite, don't delete (18 cases).
- `community_pool.go`, `handler.go` and `cli/tx_test.go` also reference the removed stakeibc handlers.
- Every task that runs `make proto-gen` must revert descriptor-only churn in unrelated `*.pb.go` (see Global Constraints).
- Keeping `RegisterImplementations` is required for historical tx decoding; the first pass dropped them.
- The full suite passed except the pre-existing `utils` `TestCreateModuleAccount` failure that also fails on main.
- Unreferenced after the removals and left in place (candidate follow-up cleanup, no behavior impact): `StartLSMLiquidStake`, `SubmitValidatorSlashQuery`, `ShouldCheckIfValidatorWasSlashed`, `EmitPendingLSMLiquidStakeEvent`, `BuildTradeAuthzMsg`, `GetTradeRouteFromTradeAccountChainId`, `RegisterTradeRouteICAAccount`, `EnableRedemptions` (stakeibc).

---

## Foundation tasks (serial)

### Task 1: v35 upgrade package skeleton and wiring

**Files:**
- Create: `app/upgrades/v35/constants.go`
- Create: `app/upgrades/v35/upgrades.go`
- Create: `app/upgrades/v35/upgrades_test.go`
- Modify: `app/upgrades.go` (imports near line 47, handler registration after the v34 block ending near line 464)

**Interfaces:**
- Produces: `v35.UpgradeName = "v35"`; `v35.CreateUpgradeHandler(mm, configurator, stakeibcKeeper, icqKeeper, autopilotKeeper, icaHostKeeper, wasmKeeper)`; test suite `UpgradeTestSuite` with `s.Setup()` and `s.ConfirmUpgradeSucceeded(v35.UpgradeName)`. Tasks 7 and 8 add calls inside the handler body between the two `Logger` lines.
- Review: no

- [ ] **Step 1: Write the failing test**

`app/upgrades/v35/upgrades_test.go`:

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

// TestUpgradeRuns is the smoke test: the handler is registered under the plan name and completes
// on an empty app. Each later task adds a focused test for its own handler step.
func (s *UpgradeTestSuite) TestUpgradeRuns() {
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./app/upgrades/v35/... 2>&1 | tail -3`
Expected: build failure, `package github.com/Stride-Labs/stride/v34/app/upgrades/v35 is not in std` or `undefined: v35.UpgradeName`

- [ ] **Step 3: Create the package**

`app/upgrades/v35/constants.go`:

```go
package v35

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v35"
)
```

`app/upgrades/v35/upgrades.go`:

```go
package v35

import (
	"context"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v34/x/autopilot/keeper"
	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

// CreateUpgradeHandler returns the v35 upgrade handler: the first wind-down upgrade
// (docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md §5). The message removals
// live in the binary itself; the handler flips the parameters and state that the removals do
// not cover, deprecates comdex-1, deletes the dYdX trade route, and trues up haqq_11235-1's
// validator delegations. Every step logs and skips on missing state rather than erroring.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
	autopilotKeeper autopilotkeeper.Keeper,
	icaHostKeeper *icahostkeeper.Keeper,
	wasmKeeper wasmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(goCtx)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s (wind-down 1: close the doors)...", UpgradeName))

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}
```

Wire it in `app/upgrades.go`. Add the import next to the v34 one:

```go
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
```

and directly after the v34 `SetUpgradeHandler(...)` call (before `upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()`):

```go
	// v35 upgrade handler
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
		),
	)
```

The unused keeper parameters are used by Tasks 7 and 8; Go does not flag unused function parameters, so this compiles.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go build ./... && go test ./app/upgrades/v35/... 2>&1 | tail -3`
Expected: `ok  	github.com/Stride-Labs/stride/v34/app/upgrades/v35`

- [ ] **Step 5: Commit**

```bash
git add app/upgrades/v35 app/upgrades.go
git commit -m "feat(upgrade): v35 handler skeleton and wiring"
```

## Parallel-safe tasks

Every task below depends only on Task 1. Tasks 2-5 each edit a disjoint set of modules. Tasks 6 and 7 both add lines to `app/upgrades/v35/upgrades.go`; that is textual overlap only, each adds its own helper file and its own calls.

### Task 2: stakeibc — remove the liquid-stake, redeem and create-things messages

**Files:**
- Modify: `proto/stride/stakeibc/tx.proto` (rpc lines 17-20 and the `CreateTradeRoute`, `UpdateTradeRoute`, `DeleteTradeRoute`, `SetCommunityPoolRebate`, `ToggleTradeController` rpcs)
- Create: `x/stakeibc/keeper/liquid_stake.go`
- Modify: `x/stakeibc/keeper/msg_server.go` (delete `RegisterHostZone` ~42-222, `LiquidStake` 224-309, `RedeemStake` 311-332, `LSMLiquidStake` 334-392, `CreateTradeRoute` 394-481, `DeleteTradeRoute` 483-517, `UpdateTradeRoute` 519-~560, `SetCommunityPoolRebate` 770-803, `ToggleTradeController` 807-end)
- Modify: `x/stakeibc/keeper/reward_allocation.go:47`
- Modify: `x/stakeibc/keeper/community_pool.go` (two `NewMsgServerImpl(k).LiquidStake/RedeemStake` calls → `k.LiquidStake` / `k.RedeemStake`)
- Modify: `x/stakeibc/handler.go` (legacy `MsgServiceHandler` switch: delete the nine cases for the removed messages)
- Modify: `x/autopilot/keeper/liquidstake.go:76-105`, `x/autopilot/keeper/redeem_stake.go:60-84`
- Modify: `x/stakeibc/types/codec.go` (amino lines 13,14,16,17,29,30,31,34,35 only; the `RegisterImplementations` entries STAY, see Global Constraints)
- Modify: `x/stakeibc/client/cli/tx.go` (AddCommand lines 43-46, 60-61 and the `Cmd*` funcs), `x/stakeibc/client/cli/gov.go` (trade route proposal commands, if present)
- Delete: `x/stakeibc/types/message_register_host_zone.go` (+`_test`), `message_lsm_liquid_stake.go` (+`_test`), `message_create_trade_route.go` (+`_test`), `message_update_trade_route.go` (+`_test`), `message_delete_trade_route.go` (+`_test`), `message_set_community_pool_rebate.go` (+`_test`), `message_toggle_trade_controller.go` (+`_test`) — only if nothing else references their helpers (check with `grep -rn NewMsgRegisterHostZone x/ app/`)
- Keep: `x/stakeibc/types/message_liquid_stake.go`, `message_redeem_stake.go` (their `NewMsg*`/`ValidateBasic` are used by the reward collector and autopilot)
- Test: `x/stakeibc/keeper/msg_server_test.go` (rewrite the LiquidStake tests to call the keeper; delete the tests of removed handlers), `x/stakeibc/keeper/redeem_stake_test.go` (its 18 calls go through `GetMsgServer().RedeemStake`; rewrite them to `s.App.StakeibcKeeper.RedeemStake(s.Ctx, &msg)`, bodies unchanged), `x/stakeibc/keeper/registration_test.go` (all 18 `TestRegisterHostZone_*` cases exercise `Keeper.RegisterHostZone`, which survives for `x/staketia/keeper/migration.go`; rewrite `s.GetMsgServer().RegisterHostZone(` → `s.App.StakeibcKeeper.RegisterHostZone(`, do not delete the file), `x/stakeibc/client/cli/tx_test.go` (delete the tests of the six removed commands), `x/stakeibc/keeper/community_pool_test.go` (delete the `SetCommunityPoolRebate` tests), `x/stakeibc/keeper/lsm_test.go` (delete the `LSMLiquidStake` handler tests, keep keeper-level ones), `x/stakeibc/keeper/liquid_stake_test.go` (new home for the LiquidStake keeper tests)
- Also delete the CLI flag constants that lose their only users (`FlagMinRedemptionRate`, `FlagMaxRedemptionRate`, `FlagCommunityPoolTreasuryAddress`, `FlagMaxMessagesPerIcaTx`, `FlagLegacy` in `x/stakeibc/client/cli/tx.go`)

**Interfaces:**
- Produces: `func (k Keeper) LiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error)` in `x/stakeibc/keeper/liquid_stake.go`. `func (k Keeper) RedeemStake(ctx sdk.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error)` already exists in `redeem_stake.go` and is unchanged.
- Depends on: Task 1
- Review: yes (money path: the LiquidStake body moves; the reward collector and autopilot must still mint correctly)

- [ ] **Step 1: Write the failing keeper test for the moved LiquidStake**

Create `x/stakeibc/keeper/liquid_stake_test.go` by moving the existing `TestLiquidStake*` cases out of `msg_server_test.go`. Every call of the form

```go
_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
```

becomes

```go
_, err := s.App.StakeibcKeeper.LiquidStake(s.Ctx, &msg)
```

Keep the test bodies otherwise identical (they exercise the deposit record, mint, and safety-bounds logic the reward collector still relies on). Add one new case at the end of the file:

```go
// The LiquidStake message has no tx handler after v35, but the keeper path must keep working
// because the reward collector liquid stakes the validators' fee share every mint epoch
func (s *KeeperTestSuite) TestLiquidStake_NoMsgHandlerButKeeperWorks() {
	tc := s.SetupLiquidStake()

	s.Require().Nil(s.App.MsgServiceRouter().Handler(&tc.validMsg), "MsgLiquidStake must have no tx handler")

	_, err := s.App.StakeibcKeeper.LiquidStake(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "keeper liquid stake")
}
```

(`SetupLiquidStake` and `tc.validMsg` already exist in the moved tests; reuse their names as-is.)

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./x/stakeibc/keeper/... -run TestKeeperTestSuite/TestLiquidStake_NoMsgHandlerButKeeperWorks 2>&1 | tail -3`
Expected: build failure `s.App.StakeibcKeeper.LiquidStake undefined`

- [ ] **Step 3: Remove the rpcs from the proto and regenerate**

In `proto/stride/stakeibc/tx.proto` delete these nine `rpc` lines (and nothing else):

```
rpc LiquidStake(...)
rpc LSMLiquidStake(...)
rpc RedeemStake(...)
rpc RegisterHostZone(...)
rpc CreateTradeRoute(...)
rpc DeleteTradeRoute(...)
rpc UpdateTradeRoute(...)
rpc SetCommunityPoolRebate(...)
rpc ToggleTradeController(...)
```

Run: `make proto-gen`
Expected: `x/stakeibc/types/tx.pb.go` regenerated; `grep -c "MethodName" x/stakeibc/types/tx.pb.go` drops by 9.

- [ ] **Step 4: Move the LiquidStake body to the keeper and delete the handlers**

Create `x/stakeibc/keeper/liquid_stake.go` with the body of the old `msgServer.LiquidStake` (lines 224-309 of `msg_server.go`), signature changed to take `sdk.Context`:

```go
package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
	epochtypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// LiquidStake mints stTokens for native tokens at the current redemption rate.
// Since v35 there is no tx handler for MsgLiquidStake; the only callers are the reward
// collector (validator fee share) and autopilot (disabled by param from v35).
func (k Keeper) LiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	// ... exact body of the former msgServer.LiquidStake, minus the
	//     `ctx := sdk.UnwrapSDKContext(goCtx)` line ...
}
```

Then in `x/stakeibc/keeper/msg_server.go` delete the nine handler functions listed in **Files**. Remove imports that become unused.

Update the two internal callers:

`x/stakeibc/keeper/reward_allocation.go:47`
```go
		liquidStakeResp, err := k.LiquidStake(ctx, msg)
```

`x/autopilot/keeper/liquidstake.go` (replace the `msgServer := ...` and `msgServer.LiquidStake(` lines)
```go
	msgResponse, err := k.stakeibcKeeper.LiquidStake(ctx, msg)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to liquid stake")
	}
```

`x/autopilot/keeper/redeem_stake.go`
```go
	if _, err = k.stakeibcKeeper.RedeemStake(ctx, msg); err != nil {
		return errorsmod.Wrapf(err, "redeem stake failed")
	}
```

Remove the now-unused `stakeibckeeper` import from both autopilot files if it was only used for `NewMsgServerImpl`.

- [ ] **Step 5: Remove codec registrations, CLI commands and dead message helpers**

`x/stakeibc/types/codec.go`: delete the `legacy.RegisterAminoMsg` lines for the nine messages. Keep their `RegisterImplementations` entries (historical tx decoding, see Global Constraints) and add a comment saying why. `RegisterInterfaces` must still end with `msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)`.

`x/stakeibc/client/cli/tx.go`: delete the `cmd.AddCommand(...)` lines and the `Cmd*` functions for `LiquidStake`, `LSMLiquidStake`, `RegisterHostZone`, `RedeemStake`, `SetCommunityPoolRebate`, `ToggleTradeController`. In `x/stakeibc/client/cli/gov.go` delete any trade-route proposal commands and their `AddCommand` lines.

Delete the `types/message_*.go` files (and their tests) listed under **Files**, after confirming with `grep -rn "NewMsg<Name>\|Msg<Name>{" x/ app/ --include='*.go' | grep -v pb.go` that only the deleted handler/CLI/test referenced them. Keep `message_liquid_stake.go` and `message_redeem_stake.go`.

- [ ] **Step 6: Fix the remaining tests**

- `msg_server_test.go`: delete every test of a removed handler (`TestRegisterHostZone*`, `TestLSMLiquidStake*`, `TestRedeemStake*` that go through `GetMsgServer()`, `TestCreateTradeRoute*`, `TestUpdateTradeRoute*`, `TestDeleteTradeRoute*`, `TestToggleTradeController*`). Keeper-level `TestRedeemStake*` in `redeem_stake_test.go` stay.
- `registration_test.go`: rewrite every `GetMsgServer().RegisterHostZone(` call to the keeper; the old handler was a pure delegate, so the tests are unchanged otherwise.
- `community_pool_test.go`: delete the `TestSetCommunityPoolRebate*` cases.
- `lsm_test.go`: delete the cases that call `GetMsgServer().LSMLiquidStake`; keep keeper-level LSM tests (the LSM callbacks still run in window 1).
- `x/autopilot/keeper/*_test.go`: should pass unchanged; fix any reference to `stakeibckeeper.NewMsgServerImpl`.

- [ ] **Step 7: Build and run the affected packages**

Run: `go build ./... && go test ./x/stakeibc/... ./x/autopilot/... 2>&1 | tail -5`
Expected: all `ok`

Run: `grep -c "^\s*rpc " proto/stride/stakeibc/tx.proto`
Expected: `14`

Run: `git diff --stat HEAD -- 'x/*/types/*.pb.go' | grep -v tx.pb.go`
Expected: nothing (descriptor noise reverted; see Global Constraints)

- [ ] **Step 8: Commit**

```bash
git add -A proto/stride/stakeibc x/stakeibc x/autopilot
git commit -m "feat(stakeibc): remove liquid stake, redeem and create-things tx handlers for wind-down"
```

### Task 3: staketia and stakedym — remove LiquidStake and RedeemStake

**Files:**
- Modify: `proto/stride/staketia/tx.proto` (rpc lines 27, 30), `proto/stride/stakedym/tx.proto` (rpc lines 27, 30)
- Modify: `x/staketia/keeper/msg_server.go` (delete `LiquidStake` 26-28, `RedeemStake` 31-39), `x/stakedym/keeper/msg_server.go` (delete `LiquidStake` 27-35, `RedeemStake` 37-45)
- Modify: `x/staketia/keeper/unbonding.go` (delete keeper `RedeemStake` 21-153 and `HandleRedemptionSpillover` 155-186), `x/stakedym/keeper/unbonding.go` (delete keeper `RedeemStake` 20-~105)
- Keep: `x/stakedym/keeper/delegation.go` `LiquidStake` (used by `LiquidStakeAndDistributeFees` in the hook) and everything else
- Modify: `x/staketia/types/msgs.go` and `x/stakedym/types/msgs.go` (delete `TypeMsgLiquidStake`, `TypeMsgRedeemStake`, `NewMsgLiquidStake`, `NewMsgRedeemStake` and their `Type/Route/GetSignBytes/ValidateBasic` methods), `x/staketia/types/codec.go`, `x/stakedym/types/codec.go` (remove the amino lines only; keep `RegisterImplementations`)
- Also dead after this task, delete: `EmitSuccessfulRedeemStakeEvent` in both modules' `keeper/events.go`, and `StakeibcKeeper.RedeemStake` in `x/staketia/types/expected_keepers.go`
- Modify: `x/staketia/client/cli/tx.go` (`CmdRedeemStake`), `x/stakedym/client/cli/tx.go` (`CmdLiquidStake`, `CmdRedeemStake`)
- Test: `x/staketia/keeper/unbonding_test.go`, `x/stakedym/keeper/unbonding_test.go`, `x/stakedym/keeper/delegation_test.go`, `x/stakedym/keeper/msg_server_test.go`, `x/staketia/types/msgs_test.go`, `x/stakedym/types/msgs_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: nothing other tasks use. `stakeibckeeper.EnableRedemptions` loses its only caller and is left in place.
- Depends on: Task 1
- Review: yes (staketia's confirm/sweep/distribute flow must survive untouched; it runs in window 1)

- [ ] **Step 1: Write the failing test**

Append to `x/staketia/keeper/unbonding_test.go`:

```go
// v35 removes staketia redemptions; the operator confirm flow must be untouched
func (s *KeeperTestSuite) TestRedeemStakeHasNoHandler() {
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgRedeemStake{}))
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgLiquidStake{}))
	s.Require().NotNil(s.App.MsgServiceRouter().Handler(&types.MsgConfirmUndelegation{}))
	s.Require().NotNil(s.App.MsgServiceRouter().Handler(&types.MsgConfirmUnbondedTokenSweep{}))
}
```

and the same four lines to `x/stakedym/keeper/unbonding_test.go` with stakedym's `types`.

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./x/staketia/keeper/... ./x/stakedym/keeper/... -run 'TestKeeperTestSuite/TestRedeemStakeHasNoHandler' 2>&1 | tail -4`
Expected: FAIL, `Expected nil, but got: ...` for the redeem handler

- [ ] **Step 3: Remove the rpcs and regenerate**

Delete `rpc LiquidStake(...)` and `rpc RedeemStake(...)` from both `proto/stride/staketia/tx.proto` and `proto/stride/stakedym/tx.proto`.

Run: `make proto-gen`

- [ ] **Step 4: Delete handlers, dead keeper code, codec entries, CLI and helpers**

Delete the four msg-server functions, staketia's keeper `RedeemStake` + `HandleRedemptionSpillover`, stakedym's keeper `RedeemStake`, the two messages from both `codec.go` files, the three CLI commands, and the `NewMsg*`/method sets in both `msgs.go`. Remove imports that become unused (`stakeibctypes` in staketia's `unbonding.go` if only the spillover used it).

- [ ] **Step 5: Fix tests**

Delete `TestRedeemStake*`, `TestHandleRedemptionSpillover*` and `TestLiquidStake*` handler tests in the four keeper test files (keep stakedym's keeper-level `TestLiquidStake*` in `delegation_test.go` if they call `s.App.StakedymKeeper.LiquidStake` directly), and the `MsgLiquidStake`/`MsgRedeemStake` cases in both `msgs_test.go`.

- [ ] **Step 6: Build and test**

Run: `go build ./... && go test ./x/staketia/... ./x/stakedym/... 2>&1 | tail -5`
Expected: all `ok`

- [ ] **Step 7: Commit**

```bash
git add -A proto/stride/staketia proto/stride/stakedym x/staketia x/stakedym
git commit -m "feat(staketia,stakedym): remove liquid stake and redeem tx handlers for wind-down"
```

### Task 4: icaoracle and icqoracle — remove the registration messages

**Files:**
- Modify: `proto/stride/icaoracle/tx.proto` (rpc lines 15, 17), `proto/stride/icqoracle/tx.proto` (rpc lines 17, 21)
- Modify: `x/icaoracle/keeper/msg_server.go` (delete `AddOracle`, `InstantiateOracle`), `x/icqoracle/keeper/msg_server.go` (delete `RegisterTokenPriceQuery`, `RemoveTokenPriceQuery`; keep `UpdateParams`)
- Modify: `x/icaoracle/types/codec.go`, `x/icqoracle/types/codec.go` (amino lines only; keep `RegisterImplementations`)
- Modify: `x/icaoracle/README.md:117-122` (drop `AddOracle`/`InstantiateOracle` from the Transactions list, note they were removed in v35)
- Modify: `x/icaoracle/client/cli/tx.go` (`CmdAddOracle`, `CmdInstantiateOracle`), `x/icqoracle/client/cli/tx.go` (`CmdAddTokenPrice`, `CmdRemoveTokenPrice`)
- Delete: `x/icaoracle/types/message_add_oracle.go` (+`_test`), `message_instantiate_oracle.go` (+`_test`); in `x/icqoracle/types/msgs.go` delete the two messages' helpers
- Keep: all icaoracle keeper logic (the instantiate ICA callback and metric posting still run) and `MsgRestoreOracleICA`, `MsgToggleOracle`, `MsgRemoveOracle`
- Test: the modules' `msg_server_test.go` and `msgs_test.go`

**Interfaces:**
- Depends on: Task 1
- Review: no

- [ ] **Step 1: Write the failing test**

Append to `x/icaoracle/keeper/msg_server_test.go`:

```go
func (s *KeeperTestSuite) TestRemovedOracleHandlers() {
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgAddOracle{}))
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgInstantiateOracle{}))
	s.Require().NotNil(s.App.MsgServiceRouter().Handler(&types.MsgToggleOracle{}))
}
```

and to `x/icqoracle/keeper/msg_server_test.go`:

```go
func (s *KeeperTestSuite) TestRemovedTokenPriceHandlers() {
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgRegisterTokenPriceQuery{}))
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgRemoveTokenPriceQuery{}))
	s.Require().NotNil(s.App.MsgServiceRouter().Handler(&types.MsgUpdateParams{}))
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./x/icaoracle/keeper/... ./x/icqoracle/keeper/... -run 'TestKeeperTestSuite/TestRemoved' 2>&1 | tail -4`
Expected: FAIL on the `Nil` assertions

- [ ] **Step 3: Remove rpcs, regenerate, delete handlers/codec/CLI/helpers, fix tests**

Delete the four rpc lines, run `make proto-gen`, delete the four handler functions, the codec entries, the four CLI commands, the icaoracle message files and the icqoracle helpers, and the tests that called the removed handlers.

- [ ] **Step 4: Build and test**

Run: `go build ./... && go test ./x/icaoracle/... ./x/icqoracle/... 2>&1 | tail -4`
Expected: all `ok`

- [ ] **Step 5: Commit**

```bash
git add -A proto/stride/icaoracle proto/stride/icqoracle x/icaoracle x/icqoracle
git commit -m "feat(icaoracle,icqoracle): remove oracle and price-query registration handlers for wind-down"
```

### Task 5: auction, airdrop and legacy claim — remove every message

**Files:**
- Modify: `proto/stride/auction/tx.proto` (rpc lines 19, 24, 26), `proto/stride/airdrop/tx.proto` (all seven rpcs), `proto/stride/claim/tx.proto` (all four rpcs). Each `service Msg { ... }` block stays, with its `option (cosmos.msg.v1.service) = true;` and no rpcs.
- Modify: `x/auction/keeper/msg_server.go`, `x/airdrop/keeper/msg_server.go`, `x/claim/keeper/msg_server.go`: delete every handler; keep the `msgServer` type and `NewMsgServerImpl` because `module.go` registers them.
- Modify: `x/auction/types/codec.go`, `x/airdrop/types/codec.go`, `x/claim/types/codec.go`: remove the amino registrations; keep every `RegisterImplementations` entry and `RegisterMsgServiceDesc`.
- Modify: `x/claim/keeper/claim.go` `CreateAirdropAndEpoch` calls `msg.ValidateBasic()` on a `MsgCreateAirdrop` (used by the v3/v8/v14 upgrade handlers); inline the identical checks (distributor bech32, non-empty identifier/chain-id/denom, valid denom, non-zero start and duration) and add a table test for them in `x/claim/keeper/claim_test.go`, since `msgs_test.go` goes.
- Modify: `x/auction/client/cli/tx.go`, `x/airdrop/client/cli/tx.go`: delete all `Cmd*` tx commands so `GetTxCmd` returns the bare parent command. `x/claim/client/cli/tx.go`: delete the four `AddCommand` lines and delete `tx_claim_free_amount.go`, `tx_create_airdrop.go`, `tx_delete_airdrop.go`, `tx_set_airdrop_allocations.go`.
- Modify: `x/auction/types/msgs.go`, `x/airdrop/types/msgs.go`, `x/claim/types/msgs.go`: delete every message's helpers.
- Keep: all keeper and epoch-hook logic and genesis; module stores are untouched.
- Test: each module's `msg_server_test.go` and `msgs_test.go`, `x/claim/client/cli/cli_test.go`

**Interfaces:**
- Depends on: Task 1
- Review: no

- [ ] **Step 1: Write the failing test**

Append to `x/auction/keeper/msg_server_test.go` (create the file with the standard `KeeperTestSuite` boilerplate from a sibling test if it does not exist):

```go
func (s *KeeperTestSuite) TestAllAuctionHandlersRemoved() {
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgPlaceBid{}))
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgCreateAuction{}))
	s.Require().Nil(s.App.MsgServiceRouter().Handler(&types.MsgUpdateAuction{}))
}
```

Same shape in airdrop (`MsgClaimDaily`, `MsgClaimEarly`, `MsgCreateAirdrop`, `MsgUpdateAirdrop`, `MsgAddAllocations`, `MsgUpdateUserAllocation`, `MsgLinkAddresses`) and claim (`MsgSetAirdropAllocations`, `MsgClaimFreeAmount`, `MsgCreateAirdrop`, `MsgDeleteAirdrop`).

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./x/auction/keeper/... ./x/airdrop/keeper/... ./x/claim/keeper/... -run 'TestKeeperTestSuite/TestAll' 2>&1 | tail -4`
Expected: FAIL on the `Nil` assertions

- [ ] **Step 3: Remove rpcs, regenerate, delete handlers/codec/CLI/helpers, fix tests**

As in Tasks 4 and 5. After `make proto-gen`, each module's `MsgServer` interface is empty; `NewMsgServerImpl` still satisfies it.

- [ ] **Step 4: Build and test**

Run: `go build ./... && go test ./x/auction/... ./x/airdrop/... ./x/claim/... 2>&1 | tail -5`
Expected: all `ok`

- [ ] **Step 5: Commit**

```bash
git add -A proto/stride/auction proto/stride/airdrop proto/stride/claim x/auction x/airdrop x/claim
git commit -m "feat(auction,airdrop,claim): remove all tx handlers for wind-down"
```

### Task 6: Handler steps — comdex, trade route, autopilot, ICA host, wasm

**Files:**
- Create: `app/upgrades/v35/params.go`
- Create: `app/upgrades/v35/params_test.go`
- Modify: `app/upgrades/v35/constants.go` (add the constants below)
- Modify: `app/upgrades/v35/upgrades.go` (add the calls after `RunMigrations`)

**Interfaces:**
- Produces: `DeprecateComdex(ctx, stakeibcKeeper)`, `DeleteDydxTradeRoute(ctx, stakeibcKeeper)`, `DisableAutopilotStakeibc(ctx, autopilotKeeper)`, `RemoveIcaHostLiquidStakingMessages(ctx, icaHostKeeper)`, `SetWasmUploadAccessToGov(ctx, wasmKeeper) error`, `TransferWasmContractAdminsToGov(ctx, wasmKeeper)` which delegates to an unexported `transferWasmContractAdminsToGov(ctx, wasmKeeper, contractAddresses []string)` so a test can pass its own contracts. `SetWasmUploadAccessToGov`'s error IS propagated by the handler (it can only fail on invalid params, and leaving upload open to the two keys with just a log is worse than a failed upgrade); every other helper logs and continues.
- Extra test (the plan's original skip-path-only test is not enough): copy `hackatom.wasm` from `$(go list -m -f '{{.Dir}}' github.com/CosmWasm/wasmd)/x/wasm/keeper/testdata/` into `app/upgrades/v35/testdata/`, store and instantiate it twice via `wasmkeeper.NewDefaultPermissionKeeper(s.App.WasmKeeper)` (`Create`, then `Instantiate` with a test admin and init msg `{"verifier":"<addr>","beneficiary":"<addr>"}`), run `transferWasmContractAdminsToGov` on the first only, and assert `GetContractInfo(first).Admin == GovModuleAddress.String()` while the second is unchanged. The test app has a working wasmvm.
- Also: `x/autopilot/README.md:34-42` gets a note that stakeibc autopilot actions are disabled by param from v35.
- Depends on: Task 1
- Review: yes (wasm admin transfer and ICA host allow-list are security-relevant)

- [ ] **Step 1: Add the constants**

Append to `app/upgrades/v35/constants.go`:

```go
import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	ComdexChainId = "comdex-1"

	// The one live trade route on mainnet (dYdX USDC rewards → Noble → Osmosis). The store is
	// keyed by the reward denom as it appears on the REWARD zone (Noble's native uusdc) plus the
	// host denom on the host zone; TradeRoute.GetKey() uses RewardDenomOnRewardZone.
	DydxTradeRouteRewardDenom = "uusdc"
	DydxTradeRouteHostDenom   = "adydx"
)

var (
	// GovModuleAddress becomes the wasm code-upload authority and the admin of every contract
	// that a Stride key currently administers.
	GovModuleAddress = authtypes.NewModuleAddress(govtypes.ModuleName)

	// WasmContractsToTransfer are the mainnet contracts whose admin is the Stride deploy key
	// stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh (listed 2026-09-21: 21 contracts total,
	// 17 with no admin, these 4 with that key). Re-list before the proposal.
	WasmContractsToTransfer = []string{
		"stride1vqk4huclshlfp0up9f0wdckv3nmfnzngwyju40kazgyc8jugmf2qe6k7mv", // code 30, hpl_igp
		"stride1ytdpedv8tkt364mzjyyrdw3nqcj3nlw7axvtfpp9wwcyjnfeclxqe2dg8v", // code 30, hpl_igp_hyperlane
		"stride1907l3m8649c7dma9a7lpqau06yqykka4xmukxlguu7j9ddlu2k6skpm57w", // code 31, hpl_hook_aggregate_hyperlane
		"stride1j50chhlj7g9prlzfh9guhpxlp6h0wnp9y3g499ry7x3cyeg0r6pspuw8tz", // code 44, hpl_ism_multisig
	}
)
```

(Merge the `import` block with the file's existing one.)

- [ ] **Step 2: Write the failing tests**

`app/upgrades/v35/params_test.go`:

```go
package v35_test

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	icahosttypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	autopilottypes "github.com/Stride-Labs/stride/v34/x/autopilot/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func (s *UpgradeTestSuite) TestDeprecateComdex() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: v35.ComdexChainId, Halted: false, Deprecated: false})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: "osmosis-1"})

	v35.DeprecateComdex(s.Ctx, s.App.StakeibcKeeper)

	comdex, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(found)
	s.Require().True(comdex.Deprecated, "comdex deprecated")
	s.Require().False(comdex.Halted, "comdex must not be halted by this upgrade")

	osmosis, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "osmosis-1")
	s.Require().False(osmosis.Deprecated, "other zones untouched")
}

func (s *UpgradeTestSuite) TestDeprecateComdexSkipsWhenMissing() {
	v35.DeprecateComdex(s.Ctx, s.App.StakeibcKeeper) // no panic, no error
	_, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().False(found)
}

func (s *UpgradeTestSuite) TestDeleteDydxTradeRoute() {
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, stakeibctypes.TradeRoute{
		RewardDenomOnRewardZone: v35.DydxTradeRouteRewardDenom,
		HostDenomOnHostZone:     v35.DydxTradeRouteHostDenom,
	})
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, stakeibctypes.TradeRoute{
		RewardDenomOnRewardZone: "ibc/other",
		HostDenomOnHostZone:     "uother",
	})

	v35.DeleteDydxTradeRoute(s.Ctx, s.App.StakeibcKeeper)

	routes := s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx)
	s.Require().Len(routes, 1, "only the dYdX route is removed")
	s.Require().Equal("uother", routes[0].HostDenomOnHostZone)
}

func (s *UpgradeTestSuite) TestDisableAutopilotStakeibc() {
	s.App.AutopilotKeeper.SetParams(s.Ctx, autopilottypes.Params{StakeibcActive: true, ClaimActive: true})

	v35.DisableAutopilotStakeibc(s.Ctx, s.App.AutopilotKeeper)

	params := s.App.AutopilotKeeper.GetParams(s.Ctx)
	s.Require().False(params.StakeibcActive, "stakeibc autopilot disabled")
	s.Require().True(params.ClaimActive, "claim autopilot untouched")
}

func (s *UpgradeTestSuite) TestRemoveIcaHostLiquidStakingMessages() {
	s.App.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
			sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}),
			sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}),
			sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
		},
	})

	v35.RemoveIcaHostLiquidStakingMessages(s.Ctx, s.App.ICAHostKeeper)

	params := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().True(params.HostEnabled)
	s.Require().Equal([]string{
		sdk.MsgTypeURL(&banktypes.MsgSend{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
	}, params.AllowMessages, "liquid stake and redeem removed, claim kept until upgrade 2")
}

func (s *UpgradeTestSuite) TestSetWasmUploadAccessToGov() {
	err := s.App.WasmKeeper.SetParams(s.Ctx, wasmtypes.Params{
		CodeUploadAccess:             wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeAnyOfAddresses, Addresses: []string{"stride1aaa", "stride1bbb"}},
		InstantiateDefaultPermission: wasmtypes.AccessTypeAnyOfAddresses,
	})
	s.Require().NoError(err)

	s.Require().NoError(v35.SetWasmUploadAccessToGov(s.Ctx, s.App.WasmKeeper))

	params := s.App.WasmKeeper.GetParams(s.Ctx)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, params.CodeUploadAccess.Permission)
	s.Require().Equal([]string{v35.GovModuleAddress.String()}, params.CodeUploadAccess.Addresses)
	s.Require().Equal(wasmtypes.AccessTypeAnyOfAddresses, params.InstantiateDefaultPermission, "instantiate default untouched")
}

// The unit app has no stored contracts, so the transfer must skip every listed address with a
// log and never panic or error. The real transfer is verified on localstride from a mainnet
// export (spec §8), where the four contracts exist.
func (s *UpgradeTestSuite) TestTransferWasmContractAdminsToGovSkipsMissingContracts() {
	s.Require().NotPanics(func() {
		v35.TransferWasmContractAdminsToGov(s.Ctx, s.App.WasmKeeper)
	})
	for _, address := range v35.WasmContractsToTransfer {
		_, err := sdk.AccAddressFromBech32(address)
		s.Require().NoError(err, "constant %s must be a valid bech32 address", address)
	}
}
```

- [ ] **Step 3: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... 2>&1 | tail -3`
Expected: build failure `undefined: v35.DeprecateComdex` (and the others)

- [ ] **Step 4: Implement `params.go`**

```go
package v35

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	icahostkeeper "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v34/x/autopilot/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// DeprecateComdex sets the Deprecated flag on comdex-1 so it carries the same annotation as
// evmos, stargaze and umee. Halted is deliberately left alone: deprecated zones are not
// touched by the wind-down; the flag is documentation, nothing reads it.
func DeprecateComdex(ctx sdk.Context, stakeibcKeeper stakeibckeeper.Keeper) {
	hostZone, found := stakeibcKeeper.GetHostZone(ctx, ComdexChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v35: host zone %s not found, skipping deprecation", ComdexChainId))
		return
	}
	hostZone.Deprecated = true
	stakeibcKeeper.SetHostZone(ctx, hostZone)
	ctx.Logger().Info(fmt.Sprintf("v35: marked %s deprecated", ComdexChainId))
}

// DeleteDydxTradeRoute removes the only live trade route. Its conversions are worth a few USDC
// a day; stopping them now removes the off-chain trade controller from the wind-down and the
// need to prove the route quiet before upgrade 2. The USDC that keeps landing in the dYdX
// withdrawal ICA is swept in window 2.
func DeleteDydxTradeRoute(ctx sdk.Context, stakeibcKeeper stakeibckeeper.Keeper) {
	if _, found := stakeibcKeeper.GetTradeRoute(ctx, DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom); !found {
		ctx.Logger().Info("v35: dYdX trade route not found, skipping deletion")
		return
	}
	stakeibcKeeper.RemoveTradeRoute(ctx, DydxTradeRouteRewardDenom, DydxTradeRouteHostDenom)
	ctx.Logger().Info("v35: deleted the dYdX trade route")
}

// DisableAutopilotStakeibc closes the IBC-memo entry point to liquid stake and redeem. The
// handlers are gone from the binary, but autopilot calls the keeper directly, so it needs its
// own switch.
func DisableAutopilotStakeibc(ctx sdk.Context, autopilotKeeper autopilotkeeper.Keeper) {
	params := autopilotKeeper.GetParams(ctx)
	params.StakeibcActive = false
	autopilotKeeper.SetParams(ctx, params)
	ctx.Logger().Info("v35: autopilot stakeibc actions disabled")
}

// RemoveIcaHostLiquidStakingMessages drops MsgLiquidStake and MsgRedeemStake from the ICA
// host allow-list. MsgClaimUndelegatedTokens stays until upgrade 2 so ICA-originated claims
// work through window 1.
func RemoveIcaHostLiquidStakingMessages(ctx sdk.Context, icaHostKeeper *icahostkeeper.Keeper) {
	removed := map[string]bool{
		sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}): true,
		sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}): true,
	}

	params := icaHostKeeper.GetParams(ctx)
	kept := make([]string, 0, len(params.AllowMessages))
	for _, typeUrl := range params.AllowMessages {
		if !removed[typeUrl] {
			kept = append(kept, typeUrl)
		}
	}
	params.AllowMessages = kept
	icaHostKeeper.SetParams(ctx, params)
	ctx.Logger().Info(fmt.Sprintf("v35: ICA host allow-list now %v", kept))
}

// SetWasmUploadAccessToGov restricts code upload to the gov module address. The instantiate
// default permission is left as it is.
func SetWasmUploadAccessToGov(ctx sdk.Context, wasmKeeper wasmkeeper.Keeper) error {
	params := wasmKeeper.GetParams(ctx)
	params.CodeUploadAccess = wasmtypes.AccessConfig{
		Permission: wasmtypes.AccessTypeAnyOfAddresses,
		Addresses:  []string{GovModuleAddress.String()},
	}
	if err := wasmKeeper.SetParams(ctx, params); err != nil {
		return err
	}
	ctx.Logger().Info("v35: wasm code upload restricted to the gov module")
	return nil
}

// TransferWasmContractAdminsToGov sets the gov module as admin of every contract in
// WasmContractsToTransfer. It uses the gov permission keeper, whose authorization policy
// allows the admin change without the current admin's signature. A contract that no longer
// exists is logged and skipped so a stale constant cannot fail the upgrade.
func TransferWasmContractAdminsToGov(ctx sdk.Context, wasmKeeper wasmkeeper.Keeper) {
	govPermissionKeeper := wasmkeeper.NewGovPermissionKeeper(wasmKeeper)
	for _, address := range WasmContractsToTransfer {
		contractAddress, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: invalid contract address constant %s, skipping", address))
			continue
		}
		if wasmKeeper.GetContractInfo(ctx, contractAddress) == nil {
			ctx.Logger().Error(fmt.Sprintf("v35: contract %s not found, skipping admin transfer", address))
			continue
		}
		if err := govPermissionKeeper.UpdateContractAdmin(ctx, contractAddress, GovModuleAddress, GovModuleAddress); err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: unable to transfer admin of %s to gov: %s", address, err.Error()))
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: contract %s admin transferred to gov", address))
	}
}
```

If `GetTradeRoute` does not exist with that signature, check `x/stakeibc/keeper/trade_route.go` for the getter name and use it; the two-denom key is the module's convention.

- [ ] **Step 5: Call the steps from the handler**

In `app/upgrades/v35/upgrades.go`, after the `RunMigrations` block and before the completion log:

```go
		DeprecateComdex(ctx, stakeibcKeeper)
		DeleteDydxTradeRoute(ctx, stakeibcKeeper)
		DisableAutopilotStakeibc(ctx, autopilotKeeper)
		RemoveIcaHostLiquidStakingMessages(ctx, icaHostKeeper)
		if err := SetWasmUploadAccessToGov(ctx, wasmKeeper); err != nil {
			return vm, err
		}
		TransferWasmContractAdminsToGov(ctx, wasmKeeper)
```

- [ ] **Step 6: Run the tests**

Run: `go build ./... && go test ./app/upgrades/v35/... 2>&1 | tail -3`
Expected: `ok`

- [ ] **Step 7: Commit**

```bash
git add app/upgrades/v35
git commit -m "feat(upgrade): v35 - deprecate comdex, delete dYdX trade route, close autopilot and ICA host entry points, wasm to gov"
```

### Task 7: Handler steps — Haqq slash-query purge and delegation delta table

**Files:**
- Create: `app/upgrades/v35/delegation_deltas.go` (copy of `app/upgrades/v34/delegation_deltas.go`)
- Create: `app/upgrades/v35/haqq.go`
- Create: `app/upgrades/v35/haqq_test.go`
- Create: `scripts/wind-down/gen_delta_table.py`
- Modify: `app/upgrades/v35/upgrades.go` (add the two calls after `RunMigrations`)

**Interfaces:**
- Consumes: `scripts/wind-down/measure_delegation_drift.py` output `scripts/wind-down/drift.json` (`zones.<chain_id>.validators[]` rows with `validator_address`, `moniker`, `recorded`, `actual` in base units).
- Produces: `PurgeHaqqSlashQueries(ctx, stakeibcKeeper, icqKeeper)`, `ReconcileHaqqDelegations(ctx, stakeibcKeeper) (appliedDelta sdkmath.Int)`, `HaqqChainId`, `HaqqDelegationDeltas []DelegationDelta`, the shared `DelegationDelta`/`mustInt`/`reconcileHostZoneDelegations`.
- Depends on: Task 1
- Review: yes (accounting: changes validator delegations and TotalDelegations)

- [ ] **Step 1: Copy the delta helper**

```bash
sed 's/^package v34$/package v35/; s/"v34: /"v35: /g' app/upgrades/v34/delegation_deltas.go > app/upgrades/v35/delegation_deltas.go
```

Open the file and confirm every log line now starts with `v35:`. The imports stay on `stride/v34` (the module path is not bumped in this plan).

- [ ] **Step 2: Write the delta-table generator**

`scripts/wind-down/gen_delta_table.py`:

```python
"""Emit a Go DelegationDelta table for one zone from measure_delegation_drift.py's drift.json.

Usage: python3 gen_delta_table.py haqq_11235-1 > /tmp/haqq_deltas.go.txt

Delta = actual on-chain delegation minus Stride's tracked delegation, in base units. Zero
rows are omitted. Paste the output over HaqqDelegationDeltas in app/upgrades/v35/haqq.go.
"""

import json
import pathlib
import sys

SCRIPT_DIR = pathlib.Path(__file__).resolve().parent


def main() -> None:
    chain_id = sys.argv[1]
    drift = json.loads((SCRIPT_DIR / "drift.json").read_text())
    zone = drift["zones"][chain_id]

    entries = []
    for row in zone["validators"]:
        delta = int(row["actual"]) - int(row["recorded"])
        if delta == 0:
            continue
        entries.append((row["moniker"], row["validator_address"], delta))

    print(f"// Generated by scripts/wind-down/gen_delta_table.py from drift.json ({drift['generated_at']})")
    print(f"var HaqqDelegationDeltas = []DelegationDelta{{")
    for moniker, address, delta in sorted(entries, key=lambda e: e[2]):
        print(f'\t{{Name: "{moniker}", Address: "{address}", Delta: mustInt("{delta}")}},')
    print("}")


if __name__ == "__main__":
    main()
```

Run: `python3 scripts/wind-down/measure_delegation_drift.py && python3 scripts/wind-down/gen_delta_table.py haqq_11235-1`
Expected: a Go var block with about 17 entries: 14 negative (undetected downtime slashes, the largest about `-853800000000000000000`, 853.8 ISLM in aISLM) and 3 positive sub-token dust entries where the chain holds slightly more than tracked. Both signs are applied; the net is negative. In `haqq_test.go` assert every delta is non-zero and the sum is negative, not that each is negative. If `drift.json` lacks `actual`/`recorded` as integers, adapt the two `int(...)` reads to the field the script actually writes (see `compute_zone_rows` in `measure_delegation_drift.py`).

- [ ] **Step 3: Write the failing tests**

`app/upgrades/v35/haqq_test.go`:

```go
package v35_test

import (
	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// seedHaqqHostZone stores a haqq host zone whose validators match the delta table, each with a
// tracked delegation large enough that every (negative) delta leaves it positive
func (s *UpgradeTestSuite) seedHaqqHostZone() (tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int) {
	tracked = map[string]sdkmath.Int{}
	trackedTotal = sdkmath.ZeroInt()
	validators := []*stakeibctypes.Validator{}
	for i, delta := range v35.HaqqDelegationDeltas {
		delegation := delta.Delta.Abs().MulRaw(10).AddRaw(int64(i + 1))
		validators = append(validators, &stakeibctypes.Validator{
			Name:                 delta.Name,
			Address:              delta.Address,
			Delegation:           delegation,
			SlashQueryInProgress: i == 0,
		})
		tracked[delta.Address] = delegation
		trackedTotal = trackedTotal.Add(delegation)
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v35.HaqqChainId,
		HostDenom:        "aISLM",
		Validators:       validators,
		TotalDelegations: trackedTotal,
	})
	return tracked, trackedTotal
}

func (s *UpgradeTestSuite) TestReconcileHaqqDelegations() {
	tracked, trackedTotal := s.seedHaqqHostZone()

	applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)

	expectedTotalDelta := sdkmath.ZeroInt()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, delta := range v35.HaqqDelegationDeltas {
		s.Require().True(delta.Delta.IsNegative(), "every haqq delta is a missed slash: %s", delta.Name)
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, delta.Address)
		s.Require().True(found)
		s.Require().Equal(tracked[delta.Address].Add(delta.Delta).String(), validator.Delegation.String(), delta.Name)
		expectedTotalDelta = expectedTotalDelta.Add(delta.Delta)
	}
	s.Require().Equal(expectedTotalDelta.String(), applied.String(), "applied delta")
	s.Require().Equal(trackedTotal.Add(expectedTotalDelta).String(), hostZone.TotalDelegations.String(), "TotalDelegations")
}

func (s *UpgradeTestSuite) TestReconcileHaqqDelegationsSkipsWhenHostZoneMissing() {
	applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().True(applied.IsZero())
}

func (s *UpgradeTestSuite) TestReconcileHaqqDelegationsSkipsStaleTable() {
	s.seedHaqqHostZone()
	// Drop the first validator so the table no longer matches the host zone
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	hostZone.Validators = hostZone.Validators[1:]
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	before, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)

	applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)

	after, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().True(applied.IsZero(), "nothing applied")
	s.Require().Equal(before.TotalDelegations.String(), after.TotalDelegations.String(), "untouched")
}

func (s *UpgradeTestSuite) TestPurgeHaqqSlashQueries() {
	s.seedHaqqHostZone()
	queries := []icqtypes.Query{
		{Id: "haqq-validator", ChainId: v35.HaqqChainId, CallbackModule: "stakeibc", CallbackId: stakeibckeeper.ICQCallbackID_Validator},
		{Id: "haqq-delegation", ChainId: v35.HaqqChainId, CallbackModule: "stakeibc", CallbackId: stakeibckeeper.ICQCallbackID_Delegation},
		{Id: "haqq-calibrate", ChainId: v35.HaqqChainId, CallbackModule: "stakeibc", CallbackId: stakeibckeeper.ICQCallbackID_Calibrate},
		{Id: "haqq-withdrawal", ChainId: v35.HaqqChainId, CallbackModule: "stakeibc", CallbackId: stakeibckeeper.ICQCallbackID_WithdrawalHostBalance},
		{Id: "osmo-validator", ChainId: "osmosis-1", CallbackModule: "stakeibc", CallbackId: stakeibckeeper.ICQCallbackID_Validator},
	}
	for _, query := range queries {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
	}

	v35.PurgeHaqqSlashQueries(s.Ctx, s.App.StakeibcKeeper, s.App.InterchainqueryKeeper)

	remaining := map[string]bool{}
	for _, query := range s.App.InterchainqueryKeeper.AllQueries(s.Ctx) {
		remaining[query.Id] = true
	}
	s.Require().Equal(map[string]bool{"haqq-withdrawal": true, "osmo-validator": true}, remaining,
		"only haqq slash-path queries are deleted")

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, validator := range hostZone.Validators {
		s.Require().False(validator.SlashQueryInProgress, "flag cleared on %s", validator.Name)
	}
}

func (s *UpgradeTestSuite) TestPurgeHaqqSlashQueriesSkipsWhenHostZoneMissing() {
	s.App.InterchainqueryKeeper.SetQuery(s.Ctx, icqtypes.Query{Id: "haqq-validator", ChainId: v35.HaqqChainId, CallbackModule: "stakeibc", CallbackId: stakeibckeeper.ICQCallbackID_Validator})

	v35.PurgeHaqqSlashQueries(s.Ctx, s.App.StakeibcKeeper, s.App.InterchainqueryKeeper)

	s.Require().Len(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), 0, "queries are purged even without a host zone")
}
```

- [ ] **Step 4: Run to verify they fail**

Run: `go test ./app/upgrades/v35/... 2>&1 | tail -3`
Expected: build failure `undefined: v35.HaqqDelegationDeltas`

- [ ] **Step 5: Implement `haqq.go`**

```go
package v35

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

const HaqqChainId = "haqq_11235-1"

// HaqqDelegationDeltas trues up haqq_11235-1's tracked validator delegations to the delegation
// ICA's on-chain balances (spec §5, §8a). Every entry is an undetected downtime slash or
// sub-token dust, so every delta is negative. Generated by
// scripts/wind-down/gen_delta_table.py; regenerate and re-measure right before the proposal.
//
// PASTE THE GENERATOR OUTPUT HERE (Task 7 step 2). The table below is the 2026-09-21 measurement.
var HaqqDelegationDeltas = []DelegationDelta{
	// {Name: "...", Address: "haqqvaloper1...", Delta: mustInt("-...")},
}

// ReconcileHaqqDelegations applies HaqqDelegationDeltas with the shared all-or-nothing helper
// and returns the total applied (zero when skipped). The stored SharesToTokensRate is left
// alone on purpose: the slash callback computes a slash as tracked delegation minus on-chain
// shares × the stored rate, so once tracked equals on-chain the later rate refresh finds
// nothing to apply and the slash is not double-counted.
func ReconcileHaqqDelegations(ctx sdk.Context, stakeibcKeeper stakeibckeeper.Keeper) (appliedDelta sdkmath.Int) {
	appliedDelta, _ = reconcileHostZoneDelegations(ctx, stakeibcKeeper, HaqqChainId, HaqqDelegationDeltas)
	return appliedDelta
}

// PurgeHaqqSlashQueries deletes every pending interchain query for haqq whose callback is on
// the slash path (validator exchange rate, delegator shares, calibration) and clears
// SlashQueryInProgress on every haqq validator. A query submitted against the pre-delta
// state has no reason to exist; the epoch hooks resubmit fresh ones. It is dynamic rather
// than a pinned id list so nothing can go stale between measurement and execution.
func PurgeHaqqSlashQueries(ctx sdk.Context, stakeibcKeeper stakeibckeeper.Keeper, icqKeeper icqkeeper.Keeper) {
	slashPathCallbacks := map[string]bool{
		stakeibckeeper.ICQCallbackID_Validator:  true,
		stakeibckeeper.ICQCallbackID_Delegation: true,
		stakeibckeeper.ICQCallbackID_Calibrate:  true,
	}

	deleted := 0
	for _, query := range icqKeeper.AllQueries(ctx) {
		if query.ChainId != HaqqChainId || !slashPathCallbacks[query.CallbackId] {
			continue
		}
		icqKeeper.DeleteQuery(ctx, query.Id)
		deleted++
	}
	ctx.Logger().Info(fmt.Sprintf("v35: deleted %d open haqq slash-path ICQs", deleted))

	hostZone, found := stakeibcKeeper.GetHostZone(ctx, HaqqChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v35: host zone %s not found, skipping slash query flag reset", HaqqChainId))
		return
	}
	for _, validator := range hostZone.Validators {
		validator.SlashQueryInProgress = false
	}
	stakeibcKeeper.SetHostZone(ctx, hostZone)
}
```

Replace the placeholder comment inside `HaqqDelegationDeltas` with the generator output from step 2 (14 entries).

- [ ] **Step 6: Call the steps from the handler**

In `app/upgrades/v35/upgrades.go`, after `RunMigrations` (order relative to Task 6's calls does not matter; put these first):

```go
		// Purge before applying the table so no in-flight query answers against pre-delta state
		PurgeHaqqSlashQueries(ctx, stakeibcKeeper, icqKeeper)
		ReconcileHaqqDelegations(ctx, stakeibcKeeper)
```

- [ ] **Step 7: Run the tests**

Run: `go build ./... && go test ./app/upgrades/v35/... 2>&1 | tail -3`
Expected: `ok`

- [ ] **Step 8: Commit**

```bash
git add app/upgrades/v35 scripts/wind-down/gen_delta_table.py
git commit -m "feat(upgrade): v35 - purge haqq slash queries and true up haqq validator delegations"
```

## Integration tasks (serial, after every parallel task has merged)

### Task 8: Removed-message guard, mainnet export suite, changelog

**Files:**
- Create: `app/upgrades/v35/removed_messages_test.go`
- Create: `app/upgrades/v35/mainnet_export_test.go`
- Create: `app/upgrades/v35/testdata/README.md`
- Create: `app/upgrades/v35/testdata/mainnet_export.json.gz` (generated, committed)
- Modify: `CHANGELOG.md` (new `## [v35.0.0]` section at the top of the version list)

**Interfaces:**
- Consumes: every helper from Tasks 6 and 7; the message types left in place by Tasks 2-5.
- Depends on: Tasks 1-7
- Review: yes (release gate)

- [ ] **Step 1: Write the removed-message guard**

`app/upgrades/v35/removed_messages_test.go`:

```go
package v35_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	airdroptypes "github.com/Stride-Labs/stride/v34/x/airdrop/types"
	auctiontypes "github.com/Stride-Labs/stride/v34/x/auction/types"
	claimtypes "github.com/Stride-Labs/stride/v34/x/claim/types"
	icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"
	icqoracletypes "github.com/Stride-Labs/stride/v34/x/icqoracle/types"
	stakedymtypes "github.com/Stride-Labs/stride/v34/x/stakedym/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

// Every message the wind-down spec §5 removes must have no tx handler; the ones it keeps must
func (s *UpgradeTestSuite) TestRemovedMessagesHaveNoHandler() {
	removed := []sdk.Msg{
		&stakeibctypes.MsgLiquidStake{}, &stakeibctypes.MsgLSMLiquidStake{}, &stakeibctypes.MsgRedeemStake{},
		&stakeibctypes.MsgRegisterHostZone{}, &stakeibctypes.MsgCreateTradeRoute{}, &stakeibctypes.MsgUpdateTradeRoute{},
		&stakeibctypes.MsgDeleteTradeRoute{}, &stakeibctypes.MsgSetCommunityPoolRebate{}, &stakeibctypes.MsgToggleTradeController{},
		&staketiatypes.MsgLiquidStake{}, &staketiatypes.MsgRedeemStake{},
		&stakedymtypes.MsgLiquidStake{}, &stakedymtypes.MsgRedeemStake{},
		&icaoracletypes.MsgAddOracle{}, &icaoracletypes.MsgInstantiateOracle{},
		&icqoracletypes.MsgRegisterTokenPriceQuery{}, &icqoracletypes.MsgRemoveTokenPriceQuery{},
		&auctiontypes.MsgPlaceBid{}, &auctiontypes.MsgCreateAuction{}, &auctiontypes.MsgUpdateAuction{},
		&airdroptypes.MsgClaimDaily{}, &airdroptypes.MsgClaimEarly{}, &airdroptypes.MsgCreateAirdrop{},
		&airdroptypes.MsgUpdateAirdrop{}, &airdroptypes.MsgAddAllocations{}, &airdroptypes.MsgUpdateUserAllocation{},
		&airdroptypes.MsgLinkAddresses{},
		&claimtypes.MsgSetAirdropAllocations{}, &claimtypes.MsgClaimFreeAmount{}, &claimtypes.MsgCreateAirdrop{},
		&claimtypes.MsgDeleteAirdrop{},
	}
	for _, msg := range removed {
		s.Require().Nil(s.App.MsgServiceRouter().Handler(msg), "%s must have no handler", sdk.MsgTypeURL(msg))
	}

	kept := []sdk.Msg{
		&stakeibctypes.MsgClaimUndelegatedTokens{}, &stakeibctypes.MsgRestoreInterchainAccount{},
		&stakeibctypes.MsgUpdateValidatorSharesExchRate{}, &stakeibctypes.MsgCalibrateDelegation{},
		&stakeibctypes.MsgRebalanceValidators{}, &stakeibctypes.MsgResumeHostZone{},
		&staketiatypes.MsgConfirmUndelegation{}, &staketiatypes.MsgConfirmUnbondedTokenSweep{},
		&stakedymtypes.MsgConfirmUndelegation{}, &stakedymtypes.MsgConfirmUnbondedTokenSweep{},
		&icaoracletypes.MsgToggleOracle{}, &icqoracletypes.MsgUpdateParams{},
	}
	for _, msg := range kept {
		s.Require().NotNil(s.App.MsgServiceRouter().Handler(msg), "%s must keep its handler", sdk.MsgTypeURL(msg))
	}
}
```

Add a second test in the same file, `TestRemovedMessagesStillDecode`: for one message per module (e.g. `&stakeibctypes.MsgLiquidStake{Creator: <valid bech32>, Amount: sdkmath.NewInt(1), HostDenom: "uatom"}`), build a tx with `s.App.TxConfig().NewTxBuilder()` + `SetMsgs`, encode with `TxConfig().TxEncoder()`, decode with `TxConfig().TxDecoder()`, and assert no error and one msg of the right type. This is what proves historical txs still decode (Global Constraints).

Run: `go test ./app/upgrades/v35/... -run 'TestUpgradeTestSuite/TestRemovedMessages' 2>&1 | tail -3`
Expected: `ok`. If a message in `removed` still has a handler, the corresponding Task 2-5 missed it; fix that module, do not edit the list.

- [ ] **Step 2: Assemble the mainnet export fixture**

`app/upgrades/v35/testdata/README.md`:

```markdown
# v35 mainnet export fixture

`mainnet_export.json.gz` is post-v34 mainnet state trimmed to what the v35 mainnet-export
suite consumes:

- `app_state.stakeibc.host_zone_list` — haqq_11235-1 and comdex-1 host zones.
- `app_state.stakeibc.trade_routes` — every trade route (one, dYdX).
- `app_state.interchainquery.queries` — every pending ICQ.

The suite skips when the file is absent. Regenerate right before the proposal, at one height,
together with `scripts/wind-down/drift.json` and the `HaqqDelegationDeltas` table so the three
describe the same snapshot:

```bash
API=https://stride-api.polkachu.com
H="x-cosmos-block-height: <HEIGHT>"
curl -s -A curl/8.0 -H "$H" $API/Stride-Labs/stride/stakeibc/host_zone/haqq_11235-1 > haqq_hz.json
curl -s -A curl/8.0 -H "$H" $API/Stride-Labs/stride/stakeibc/host_zone/comdex-1     > comdex_hz.json
curl -s -A curl/8.0 -H "$H" $API/Stride-Labs/stride/stakeibc/trade_routes            > routes.json
curl -s -A curl/8.0 -H "$H" "$API/Stride-Labs/stride/interchainquery/pending_queries" > icqs.json

jq -n --slurpfile h haqq_hz.json --slurpfile c comdex_hz.json --slurpfile r routes.json --slurpfile q icqs.json \
  '{app_state: {stakeibc: {host_zone_list: [$h[0].host_zone, $c[0].host_zone], trade_routes: $r[0].trade_routes},
                interchainquery: {queries: $q[0].pending_queries}}}' \
  | gzip -9 > mainnet_export.json.gz
```

Record the height and date here when regenerating.
```

Run the commands with `<HEIGHT>` set to the latest height (`curl -s -A curl/8.0 $API/cosmos/base/tendermint/v1beta1/blocks/latest | jq -r .block.header.height`) and commit the `.gz`. Also run `python3 scripts/wind-down/measure_delegation_drift.py` at the same time and regenerate `HaqqDelegationDeltas` if it changed.

- [ ] **Step 3: Write the mainnet export suite**

`app/upgrades/v35/mainnet_export_test.go`:

```go
package v35_test

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const mainnetExportPath = "testdata/mainnet_export.json.gz"

// MainnetExportTestSuite replays the v35 handler against real mainnet state so the haqq delta
// table, the comdex flag, the trade route deletion and the ICQ purge are exercised on the
// exact validators, routes and queries they will meet. It is the release gate for the
// constants; the synthetic UpgradeTestSuite covers the logic.
type MainnetExportTestSuite struct {
	apptesting.AppTestHelper
}

func (s *MainnetExportTestSuite) SetupTest() {
	s.Setup()
}

func TestMainnetExportTestSuite(t *testing.T) {
	if _, err := os.Stat(mainnetExportPath); errors.Is(err, os.ErrNotExist) {
		t.Skipf("skipping: mainnet export fixture not present at %s — see testdata/README.md", mainnetExportPath)
	}
	suite.Run(t, new(MainnetExportTestSuite))
}

type strideExport struct {
	AppState struct {
		Stakeibc struct {
			HostZoneList []json.RawMessage `json:"host_zone_list"`
			TradeRoutes  []json.RawMessage `json:"trade_routes"`
		} `json:"stakeibc"`
		Interchainquery struct {
			Queries []json.RawMessage `json:"queries"`
		} `json:"interchainquery"`
	} `json:"app_state"`
}

func (s *MainnetExportTestSuite) loadExport() strideExport {
	file, err := os.Open(mainnetExportPath)
	s.Require().NoError(err)
	defer file.Close()
	reader, err := gzip.NewReader(file)
	s.Require().NoError(err)
	var export strideExport
	s.Require().NoError(json.NewDecoder(reader).Decode(&export))
	return export
}

func (s *MainnetExportTestSuite) TestUpgradeFromMainnetExport() {
	export := s.loadExport()

	hostZones := map[string]stakeibctypes.HostZone{}
	for _, raw := range export.AppState.Stakeibc.HostZoneList {
		var hostZone stakeibctypes.HostZone
		s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &hostZone))
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
		hostZones[hostZone.ChainId] = hostZone
	}
	for _, raw := range export.AppState.Stakeibc.TradeRoutes {
		var route stakeibctypes.TradeRoute
		s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &route))
		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)
	}
	queriesBefore := map[string]icqtypes.Query{}
	for _, raw := range export.AppState.Interchainquery.Queries {
		var query icqtypes.Query
		s.Require().NoError(s.App.AppCodec().UnmarshalJSON(raw, &query))
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
		queriesBefore[query.Id] = query
	}
	haqqBefore, found := hostZones[v35.HaqqChainId]
	s.Require().True(found, "fixture must contain haqq")
	s.Require().Len(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), 1, "fixture must contain the dYdX route")

	s.ConfirmUpgradeSucceeded(v35.UpgradeName)

	// Haqq: every validator in the table moved by exactly its delta, nothing else moved
	haqqAfter, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	expectedTotal := haqqBefore.TotalDelegations
	deltas := map[string]sdkmath.Int{}
	for _, delta := range v35.HaqqDelegationDeltas {
		deltas[delta.Address] = delta.Delta
		expectedTotal = expectedTotal.Add(delta.Delta)
	}
	for _, before := range haqqBefore.Validators {
		after, _, found := stakeibckeeper.GetValidatorFromAddress(haqqAfter.Validators, before.Address)
		s.Require().True(found, before.Name)
		expected := before.Delegation
		if delta, inTable := deltas[before.Address]; inTable {
			expected = expected.Add(delta)
		}
		s.Require().Equal(expected.String(), after.Delegation.String(), before.Name)
		s.Require().False(after.SlashQueryInProgress, "slash flag cleared on %s", before.Name)
	}
	s.Require().Equal(expectedTotal.String(), haqqAfter.TotalDelegations.String(), "haqq TotalDelegations")
	s.Require().NotEqual(haqqBefore.TotalDelegations.String(), haqqAfter.TotalDelegations.String(),
		"the table must have applied (a skip means the constants no longer match the export)")

	// Comdex: deprecated, not halted
	comdex, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(comdex.Deprecated)
	s.Require().False(comdex.Halted)

	// Trade route gone
	s.Require().Len(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), 0)

	// ICQs: haqq slash-path queries gone, everything else intact
	slashPath := map[string]bool{
		stakeibckeeper.ICQCallbackID_Validator: true, stakeibckeeper.ICQCallbackID_Delegation: true, stakeibckeeper.ICQCallbackID_Calibrate: true,
	}
	remaining := map[string]bool{}
	for _, query := range s.App.InterchainqueryKeeper.AllQueries(s.Ctx) {
		remaining[query.Id] = true
	}
	for id, query := range queriesBefore {
		shouldBeDeleted := query.ChainId == v35.HaqqChainId && slashPath[query.CallbackId]
		s.Require().Equal(!shouldBeDeleted, remaining[id], "query %s (%s/%s)", id, query.ChainId, query.CallbackId)
	}
}
```

Run: `go test ./app/upgrades/v35/... 2>&1 | tail -3`
Expected: `ok` including `TestMainnetExportTestSuite` (not skipped).

- [ ] **Step 4: Changelog entry**

Add at the top of the versions in `CHANGELOG.md`, matching the existing section format:

```markdown
## [v35.0.0] - 2026-XX-XX

### Wind-down: close the doors

- Removed the liquid stake, LSM liquid stake and redeem tx handlers (stakeibc, staketia, stakedym). The stakeibc keeper paths remain for the reward collector, community pool and in-flight flows; the staketia and stakedym redeem keeper paths go with their handlers. Historical txs still decode (types stay registered); new submissions are rejected by the router.
- Removed the register-host-zone, trade route, community-pool rebate and trade-controller handlers (stakeibc), oracle registration (icaoracle), token-price registration (icqoracle), and every auction, airdrop and legacy claim message.
- Autopilot stakeibc actions disabled; `MsgLiquidStake` and `MsgRedeemStake` dropped from the ICA host allow-list.
- Wasm code upload restricted to the gov module; admin of the four Stride-administered Hyperlane contracts transferred to gov.
- comdex-1 marked deprecated; dYdX trade route deleted.
- haqq_11235-1 validator delegations trued up to the chain (undetected downtime slashes); open haqq slash-path ICQs purged.
```

- [ ] **Step 5: Full build and the module test sweep**

Run: `go build ./... && go test ./app/... ./x/... 2>&1 | grep -v "^ok\|no test files" | head -20`
Expected: no output (every package `ok` or has no tests). Note the pre-existing `TestCreateModuleAccount` failure in `utils` if it appears; it is on main and unrelated.

- [ ] **Step 6: Commit**

```bash
git add app/upgrades/v35 CHANGELOG.md
git commit -m "test(upgrade): v35 removed-message guard and mainnet export suite; changelog"
```

### Task 9: Localstride dry run (manual, documents the release checklist)

**Files:**
- Modify: `app/upgrades/v35/testdata/README.md` (append the dry-run notes and results)

**Interfaces:**
- Depends on: Tasks 1-9
- Review: no

- [ ] **Step 1: Start localstride from a mainnet export on the v34 binary and upgrade to v35**

Follow the repo's existing localstride upgrade procedure (memory note: use the old-release binary plus the uncached-context patch, then swap; `pkill -x strided` between runs). Submit the v35 upgrade plan and let it execute.

- [ ] **Step 2: Verify on the running chain**

```bash
strided q stakeibc host-zone comdex-1 2>/dev/null | grep -E "deprecated|halted"
strided q stakeibc trade-routes 2>/dev/null
strided q autopilot params 2>/dev/null
strided q interchain-accounts host params 2>/dev/null
strided q wasm params 2>/dev/null
GOV=$(strided q auth module-account gov -o json 2>/dev/null | jq -r '.account.value.address // .account.base_account.address'); echo "gov: $GOV"
for c in stride1vqk4huclshlfp0up9f0wdckv3nmfnzngwyju40kazgyc8jugmf2qe6k7mv stride1ytdpedv8tkt364mzjyyrdw3nqcj3nlw7axvtfpp9wwcyjnfeclxqe2dg8v stride1907l3m8649c7dma9a7lpqau06yqykka4xmukxlguu7j9ddlu2k6skpm57w stride1j50chhlj7g9prlzfh9guhpxlp6h0wnp9y3g499ry7x3cyeg0r6pspuw8tz; do strided q wasm contract $c 2>/dev/null | grep admin; done
strided tx stakeibc liquid-stake 1000 uatom --from val 2>&1 | tail -1
```

Expected: `deprecated: true`, `halted: false`; no trade routes; `stakeibc_active: false`; allow list without the two stakeibc messages; upload access `AnyOfAddresses` with only the gov address; every contract `admin:` equal to `$GOV`; the liquid-stake tx rejected with an unknown-message error.

- [ ] **Step 3: Let two stride epochs pass and confirm reinvest, unbonding and claim still run**

Watch the logs for `Reinvesting tokens`, `Initiating unbondings`, and a successful `ClaimUndelegatedTokens` on a claimable record. Record the results in the README and commit.

```bash
git add app/upgrades/v35/testdata/README.md
git commit -m "docs(upgrade): v35 localstride dry-run notes"
```
