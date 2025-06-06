package v27

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

var (
	UpgradeName = "v27"

	EvmosChainId = "evmos_9001-2"
	GaiaChainId  = "cosmoshub-4"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v27...")

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Resetting delegation changes in progress...")
		if err := DecrementEvmosDelegationChangesInProgress(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to reset delegation changes in progress")
		}

		ctx.Logger().Info("Enabling LSM...")
		if err := EnableLSMForGaia(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to enable LSM")
		}

		return versionMap, nil
	}
}

// Decrement DelegationChangesInProgress on Evmos vals by 1
func DecrementEvmosDelegationChangesInProgress(
	ctx sdk.Context,
	sk stakeibckeeper.Keeper,
) error {
	hostZone, found := sk.GetHostZone(ctx, EvmosChainId)
	if !found {
		return stakeibctypes.ErrHostZoneNotFound.Wrapf("failed to fetch %s", EvmosChainId)
	}

	for _, val := range hostZone.Validators {
		if val.DelegationChangesInProgress > 0 {
			val.DelegationChangesInProgress = val.DelegationChangesInProgress - 1
		}
	}

	sk.SetHostZone(ctx, hostZone)

	return nil
}

// Enable LSM liquid stakes for Gaia
// LSM Liquid Stakes will have been disabled via governance before this upgrade passes
func EnableLSMForGaia(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	hostZone, found := k.GetHostZone(ctx, GaiaChainId)
	if !found {
		return stakeibctypes.ErrHostZoneNotFound.Wrap(GaiaChainId)
	}

	hostZone.LsmLiquidStakeEnabled = true
	k.SetHostZone(ctx, hostZone)

	return nil
}
