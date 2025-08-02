package v27

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	stakeibckeeper "github.com/Stride-Labs/stride/v28/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

var (
	UpgradeName = "v27"

	EvmosChainId   = "evmos_9001-2"
	GaiaChainId    = "cosmoshub-4"
	OsmosisChainId = "osmosis-1"

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdkmath.LegacyMustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdkmath.LegacyMustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisRedemptionRateBuffer = sdkmath.LegacyMustNewDecFromStr("0.02")

	// Inner redemption rate adjustment variables
	RedemptionRateInnerAdjustment = sdkmath.LegacyMustNewDecFromStr("0.001")
)

// CreateUpgradeHandler creates an SDK upgrade handler for v27
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
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

		ctx.Logger().Info("Update redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

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

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Updating redemption rate outer bounds...")

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
