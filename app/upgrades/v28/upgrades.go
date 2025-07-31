package v28

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	consumerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/consumer/keeper"

	stakeibckeeper "github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
)

var (
	UpgradeName = "v28"

	// Redemption rate bounds updated to give slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdkmath.LegacyMustNewDecFromStr("0.50")
	RedemptionRateOuterMaxAdjustment = sdkmath.LegacyMustNewDecFromStr("1.00")
)

// CreateUpgradeHandler creates an SDK upgrade handler for v27
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	consumerKeeper consumerkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		// Initialize consumer ID
		// https://github.com/cosmos/interchain-security/blob/v6.4.1/UPGRADING.md#consumer
		ctx.Logger().Info("Setting consumer ID parameter...")
		InitializeConsumerId(ctx, consumerKeeper)

		// Apply distribution fix
		ctx.Logger().Info("Applying distribution module fix...")
		if err := ApplyDistributionFix(ctx, distrKeeper); err != nil {
			// Log warning but continue with upgrade (non-critical)
			ctx.Logger().Warn("Failed to apply distribution fix, continuing...", "warning", err.Error())
		} else {
			ctx.Logger().Info("Distribution fix successfully applied")
		}

		// Loosen slack from redemption rate bounds
		ctx.Logger().Info("Update redemption rate bounds...")
		UpdateRedemptionRateBounds(ctx, stakeibcKeeper)

		return versionMap, nil
	}
}

// InitializeConsumerId sets the consumer Id parameter in the consumer module,
// to the consumer id for which the consumer is registered on the provider chain.
// The consumer id can be obtained in by querying the provider, e.g. by using the
// QueryConsumerIdFromClientId query.
//
// Steps to retrieve the Stride consumer chain ID from Cosmos Hub provider:
//  1. First, obtain the client ID from Stride using the command:
//     `strided q ccvconsumer provider-info` which returns "07-tendermint-1154"
//  2. Then, use the Provider's QueryConsumerIdFromClientId endpoint to get the corresponding consumer ID:
//     - API endpoint: https://rest.cosmos.directory/cosmoshub/interchain_security/ccv/provider/consumer_id/07-tendermint-1154
//     - This endpoint implements the query defined in the Interchain Security repository at:
//     https://github.com/cosmos/interchain-security/blob/307b1446/proto/interchain_security/ccv/provider/v1/query.proto#L132-L138
func InitializeConsumerId(ctx sdk.Context, consumerKeeper consumerkeeper.Keeper) {
	params := consumerKeeper.GetConsumerParams(ctx)
	params.ConsumerId = "1"
	consumerKeeper.SetParams(ctx, params)
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
