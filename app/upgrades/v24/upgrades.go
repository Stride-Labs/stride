package v24

import (
	"context"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	recordskeeper "github.com/Stride-Labs/stride/v28/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v28/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v28/x/stakeibc/keeper"
)

var (
	UpgradeName = "v24"

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdkmath.LegacyMustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdkmath.LegacyMustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdkmath.LegacyMustNewDecFromStr("0.02")
)

// CreateUpgradeHandler creates an SDK upgrade handler for v24
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bankKeeper bankkeeper.Keeper,
	recordsKeeper recordskeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info("Starting upgrade v24...")

		// Migrate data structures
		MigrateHostZones(ctx, stakeibcKeeper)
		MigrateDepositRecords(ctx, recordsKeeper)
		MigrateEpochUnbondingRecords(ctx, recordsKeeper)
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Migrate host zones to accomodate the staketia migration changes, adding a
// redemptions enabled field to each host zone
func MigrateHostZones(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Migrating host zones...")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		hostZone.RedemptionsEnabled = true
		k.SetHostZone(ctx, hostZone)
	}
}

// Migrates the deposit records to set the DelegationTxsInProgress field
// which should be 1 if the status was DELEGATION_IN_PROGRESS, and 0 otherwise
func MigrateDepositRecords(ctx sdk.Context, k recordskeeper.Keeper) {
	ctx.Logger().Info("Migrating deposit records...")

	for _, depositRecord := range k.GetAllDepositRecord(ctx) {
		if depositRecord.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS {
			depositRecord.DelegationTxsInProgress = 1
		} else {
			depositRecord.DelegationTxsInProgress = 0
		}
		k.SetDepositRecord(ctx, depositRecord)
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
//
// If the record is in status UNBONDING_IN_PROGRESS, we need to also set UndelegationTxsInProgress to 1;
// otherwise, it should be set to 0
func MigrateHostZoneUnbondingRecords(hostZoneUnbonding *recordstypes.HostZoneUnbonding) *recordstypes.HostZoneUnbonding {
	if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_QUEUE ||
		hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE ||
		hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS {

		hostZoneUnbonding.StTokensToBurn = sdkmath.ZeroInt()
		hostZoneUnbonding.NativeTokensToUnbond = sdkmath.ZeroInt()
		hostZoneUnbonding.ClaimableNativeTokens = sdkmath.ZeroInt()
		hostZoneUnbonding.UndelegationTxsInProgress = 0

	} else if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
		hostZoneUnbonding.StTokensToBurn = hostZoneUnbonding.StTokenAmount
		hostZoneUnbonding.NativeTokensToUnbond = hostZoneUnbonding.NativeTokenAmount
		hostZoneUnbonding.ClaimableNativeTokens = sdkmath.ZeroInt()
		hostZoneUnbonding.UndelegationTxsInProgress = 1

	} else if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_CLAIMABLE {
		hostZoneUnbonding.StTokensToBurn = sdkmath.ZeroInt()
		hostZoneUnbonding.NativeTokensToUnbond = sdkmath.ZeroInt()
		hostZoneUnbonding.ClaimableNativeTokens = hostZoneUnbonding.NativeTokenAmount
		hostZoneUnbonding.UndelegationTxsInProgress = 0
	}

	return hostZoneUnbonding
}

// Migrate epoch unbonding records to accomodate the batched undelegations code changes,
// adding the new accounting fields to the host zone unbonding records
func MigrateEpochUnbondingRecords(ctx sdk.Context, k recordskeeper.Keeper) {
	ctx.Logger().Info("Migrating epoch unbonding records...")

	for _, epochUnbondingRecord := range k.GetAllEpochUnbondingRecord(ctx) {
		for i, oldHostZoneUnbondingRecord := range epochUnbondingRecord.HostZoneUnbondings {
			updatedHostZoneUnbondingRecord := MigrateHostZoneUnbondingRecords(oldHostZoneUnbondingRecord)
			epochUnbondingRecord.HostZoneUnbondings[i] = updatedHostZoneUnbondingRecord
		}
		k.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	}
}

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Updating redemption rate bounds...")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		// Give osmosis a bit more slack since OSMO stakers collect real yield
		outerAdjustment := RedemptionRateOuterMaxAdjustment
		if hostZone.ChainId == OsmosisChainId {
			outerAdjustment = outerAdjustment.Add(OsmosisRedemptionRateBuffer)
		}

		outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
		outerMaxDelta := hostZone.RedemptionRate.Mul(outerAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
		hostZone.MaxRedemptionRate = outerMax

		k.SetHostZone(ctx, hostZone)
	}
}
