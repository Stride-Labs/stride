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

		// Entry points that bypass the msg service router (spec §5)
		DisableAutopilotStakeibc(ctx, autopilotKeeper)
		RemoveStakeibcFromICAHostAllowList(ctx, icaHostKeeper)

		// Wasm control to gov: the upload-access write is the one step that may fail the upgrade
		if err := SetWasmUploadAccessToGov(ctx, wasmKeeper); err != nil {
			return vm, err
		}
		MoveDeployKeyContractAdminsToGov(ctx, wasmKeeper)

		// Helpers are added here by the later tasks, in the order fixed by the plan

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}
