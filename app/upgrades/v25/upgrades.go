package v25

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordskeeper "github.com/Stride-Labs/stride/v24/x/records/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v24/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v24/x/stakeibc/types"
	staketiakeeper "github.com/Stride-Labs/stride/v24/x/staketia/keeper"
	staketiatypes "github.com/Stride-Labs/stride/v24/x/staketia/types"
)

var (
	UpgradeName = "v25"

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdk.MustNewDecFromStr("0.02")
)

// CreateUpgradeHandler creates an SDK upgrade handler for v25
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bankKeeper bankkeeper.Keeper,
	recordsKeeper recordskeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
	staketiaKeeper staketiakeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v25...")

		// Migrate staketia to stakeibc
		if err := staketiakeeper.InitiateMigration(ctx, staketiaKeeper, bankKeeper, recordsKeeper, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate staketia to stakeibc")
		}

		// Add celestia validators
		if err := AddCelestiaValidators(ctx, stakeibcKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to add celestia validators")
		}

		// Update redemption rate bounds
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Adds the full celestia validator set, with a 0 delegation for each
func AddCelestiaValidators(ctx sdk.Context, k stakeibckeeper.Keeper) error {
	for _, validatorConfig := range Validators {
		validator := stakeibctypes.Validator{
			Name:    validatorConfig.name,
			Address: validatorConfig.address,
			Weight:  validatorConfig.weight,
		}

		if err := k.AddValidatorToHostZone(ctx, staketiatypes.CelestiaChainId, validator, false); err != nil {
			return err
		}

		// Query and store the validator's sharesToTokens rate
		if err := k.QueryValidatorSharesToTokensRate(ctx, staketiatypes.CelestiaChainId, validatorConfig.address); err != nil {
			return err
		}
	}
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
