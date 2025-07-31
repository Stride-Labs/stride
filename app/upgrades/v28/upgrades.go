package v28

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqkeeper "github.com/Stride-Labs/stride/v27/x/interchainquery/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
)

var (
	UpgradeName = "v28"

	EvmosChainId          = "evmos_9001-2"
	QueryId               = "2c39af4c3d2ecb96d8bbf7f3386468c5909e51fe3364b8d1f9d6fce173dd1f7a"
	QueryValidatorAddress = "evmosvaloper1tdss4m3x7jy9mlepm2dwy8820l7uv6m2vx6z88"
	EvmosDelegationIca    = "evmos1d67tx0zekagfhw6chhgza6qmhyad5qprru0nwazpx5s85ld0wh2sdhhznd"

	// Redemption rate bounds updated to give slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.50")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("1.00")
)

// CreateUpgradeHandler creates an SDK upgrade handler for v27
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	icqKeeper icqkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v28...")

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Update redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		ctx.Logger().Info("Processing stale ICQ...")
		ClearStuckEvmosQuery(ctx, stakeibcKeeper, icqKeeper)

		return versionMap, nil
	}
}

// Updates the outer redemption rate bounds
func UpdateRedemptionRateBounds(ctx sdk.Context, k stakeibckeeper.Keeper) {
	ctx.Logger().Info("Updating redemption rate outer bounds...")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
		outerMaxDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMaxAdjustment)

		outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
		outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

		hostZone.MinRedemptionRate = outerMin
		hostZone.MaxRedemptionRate = outerMax

		k.SetHostZone(ctx, hostZone)
	}
}

// Cleans up the stale ICQ
func ClearStuckEvmosQuery(ctx sdk.Context, k stakeibckeeper.Keeper, icqKeeper icqkeeper.Keeper) {
	ctx.Logger().Info("Deleting stale ICQ...")
	icqKeeper.DeleteQuery(ctx, QueryId)

	ctx.Logger().Info("Setting validator slash_query_in_progress to false...")
	hostZone, found := k.GetHostZone(ctx, EvmosChainId)
	if !found {
		ctx.Logger().Error("host zone not found")
		return
	}

	// find the right validator and set slash_query_in_progress to false
	for i, validator := range hostZone.Validators {
		if validator.Address == QueryValidatorAddress {
			validator.SlashQueryInProgress = false
			hostZone.Validators[i] = validator
			k.SetHostZone(ctx, hostZone)

			ctx.Logger().Info("Set validator slash_query_in_progress to false")
			return
		}
	}
}
