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
// The steps run in this order:
//  1. RunMigrations.
//  2. Turn off autopilot stakeibc and drop liquid stake / redeem stake from the ICA host allow-list.
//  3. Restrict wasm code upload to gov, then move the deploy key's contract admins to gov.
//  4. Mark comdex-1 deprecated and delete the dYdX trade route.
//  5. Deactivate the ICA oracles and empty the rate limiter.
//  6. Reset stale DelegationChangesInProgress flags on zones with no ICA in flight.
//  7. Purge haqq's pending slash-path ICQs, then all withdrawal-balance and calibration ICQs.
//  8. Apply the haqq delegation delta table (after its ICQs are gone).
//  9. Requeue the one failed LSM detokenization with its amount reduced by one.
//
// icaHostKeeper and ratelimitKeeper are pointers because their methods have pointer
// receivers. The ICA controller and channel keepers used by the stale-flag reset are read
// through the stakeibc keeper's exported fields, as is the records keeper used by the LSM reset.
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
		PurgeCalibrationQueries(ctx, icqKeeper)

		// Haqq delegation reconciliation, after its slash-path ICQs are gone (spec §5)
		ReconcileHaqqDelegations(ctx, stakeibcKeeper)

		// The failed LSM detokenization, retried by the EndBlocker with one token less (spec §5)
		ResetFailedLSMDeposit(ctx, stakeibcKeeper.RecordsKeeper)

		ctx.Logger().Info(fmt.Sprintf("Upgrade %s complete", UpgradeName))
		return vm, nil
	}
}
