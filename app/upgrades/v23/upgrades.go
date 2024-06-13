package v23

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordskeeper "github.com/Stride-Labs/stride/v22/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v22/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v22/x/stakeibc/keeper"
)

var (
	UpgradeName = "v23"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	recordsKeeper recordskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v23...")

		ctx.Logger().Info("Migrating trade routes...")
		MigrateTradeRoutes(ctx, stakeibcKeeper)

		ctx.Logger().Info("Migrating epoch unbonding records...")
		MigrateEpochUnbondingRecords(ctx, recordsKeeper)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Migration to deprecate the trade config
// The min transfer amount can be set from the min swap amount
func MigrateTradeRoutes(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, tradeRoute := range k.GetAllTradeRoutes(ctx) {
		tradeRoute.MinTransferAmount = tradeRoute.TradeConfig.MinSwapAmount
		k.SetTradeRoute(ctx, tradeRoute)
	}
}

// Migrates a single host zone unbonding record to add the new fields: StTokensToBurn,
// NativeTokensToUnbond, and ClaimableNativeTokens
//
// If the record is in status: UNBONDING_QUEUE, EXIT_TRANSFER_QUEUE, or EXIT_TRANSFER_IN_PROGRESS,
// set stTokensToBurn, NativeTokensToUnbond, and ClaimableNativeTokens all to 0
//
// If the record is in status: UNBONDING_IN_PROGRESS
// set StTokensToBurn to the value of StTokenAmount, NativeTokensToUnbond to the value of NativeTokenAmount,
// and ClaimableNativeTokens to 0
//
// If the record is in status CLAIMABLE,
// set StTokensToBurn and NativeTokensToUnbond to 0, and set ClaimableNativeTokens to the value of NativeTokenAmount
func MigrateHostZoneUnbondingRecords(hostZoneUnbonding *recordstypes.HostZoneUnbonding) *recordstypes.HostZoneUnbonding {
	if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_QUEUE ||
		hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE ||
		hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS {

		hostZoneUnbonding.StTokensToBurn = sdkmath.ZeroInt()
		hostZoneUnbonding.NativeTokensToUnbond = sdkmath.ZeroInt()
		hostZoneUnbonding.ClaimableNativeTokens = sdkmath.ZeroInt()

	} else if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
		hostZoneUnbonding.StTokensToBurn = hostZoneUnbonding.StTokenAmount
		hostZoneUnbonding.NativeTokensToUnbond = hostZoneUnbonding.NativeTokenAmount
		hostZoneUnbonding.ClaimableNativeTokens = sdkmath.ZeroInt()

	} else if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_CLAIMABLE {
		hostZoneUnbonding.StTokensToBurn = sdkmath.ZeroInt()
		hostZoneUnbonding.NativeTokensToUnbond = sdkmath.ZeroInt()
		hostZoneUnbonding.ClaimableNativeTokens = hostZoneUnbonding.NativeTokenAmount
	}

	return hostZoneUnbonding
}

// Migrate epoch unbonding records to add the new fields from the batched undelegations code
func MigrateEpochUnbondingRecords(ctx sdk.Context, k recordskeeper.Keeper) {
	for _, epochUnbondingRecord := range k.GetAllEpochUnbondingRecord(ctx) {
		for i, oldHostZoneUnbondingRecord := range epochUnbondingRecord.HostZoneUnbondings {
			updatedHostZoneUnbondingRecord := MigrateHostZoneUnbondingRecords(oldHostZoneUnbondingRecord)
			epochUnbondingRecord.HostZoneUnbondings[i] = updatedHostZoneUnbondingRecord
		}
		k.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	}
}
