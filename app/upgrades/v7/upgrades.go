package v7

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	epochskeeper "github.com/Stride-Labs/stride/v6/x/epochs/keeper"
	epochstypes "github.com/Stride-Labs/stride/v6/x/epochs/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v6/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v7"
)

// Helper function to log the migrated modules consensus versions
func logModuleMigration(ctx sdk.Context, versionMap module.VersionMap, moduleName string) {
	currentVersion := versionMap[moduleName]
	ctx.Logger().Info(fmt.Sprintf("migrating module %s from version %d to version %d", moduleName, currentVersion-1, currentVersion))
}

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	epochskeeper epochskeeper.Keeper,
	stakeibckeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// TODO:
		// ICA
		// 		add allow messages
		// misc
		// 		incentive diversification
		// 		inflation
		//  	BaseAccount issue

		// Add an hourly epoch which will be used by the rate limit store
		AddHourEpoch(ctx, epochskeeper)

		// Change the juno unbonding frequency to 5
		ModifyJunoUnbondingFrequency(ctx, stakeibckeeper)

		// Add min/max redemption rate threshold for each host zone
		AddMinMaxRedemptionRate(ctx, stakeibckeeper)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Add a new hourly epoch that will be used by the rate limit module
func AddHourEpoch(ctx sdk.Context, k epochskeeper.Keeper) {
	ctx.Logger().Info("Adding hour epoch")

	hourEpoch := epochstypes.EpochInfo{
		Identifier:              epochstypes.HOUR_EPOCH,
		StartTime:               time.Time{},
		Duration:                time.Hour,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}

	k.SetEpochInfo(ctx, hourEpoch)
}

// Update the unbonding frequency of juno to 5 to align with the 28 day unbonding period
func ModifyJunoUnbondingFrequency(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	ctx.Logger().Info("Updating juno unbonding frequency")

	junoChainId := "juno-1"
	unbondingFrequency := uint64(5)

	junoHostZone, found := k.GetHostZone(ctx, junoChainId)
	if !found {
		return stakeibctypes.ErrHostZoneNotFound
	}
	junoHostZone.UnbondingFrequency = unbondingFrequency
	k.SetHostZone(ctx, junoHostZone)

	return nil
}

// Add the min/max redemption rate to each host zone as safety bounds
// Use the default min/max for each
func AddMinMaxRedemptionRate(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	// TODO
	return nil
}
