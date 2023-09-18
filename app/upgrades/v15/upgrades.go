package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
)

var (
	UpgradeName = "v15"

	EvmosChainId                = "evmos_9001-2"
	EvmosOuterMinRedemptionRate = sdk.MustNewDecFromStr("1.290")
	EvmosInnerMinRedemptionRate = sdk.MustNewDecFromStr("1.318")
	EvmosMaxRedemptionRate      = sdk.MustNewDecFromStr("1.500")

	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateInnerMinAdjustment = sdk.MustNewDecFromStr("0.03")
	RedemptionRateInnerMaxAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("0.10")
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v15...")

		// Set host zone redemption rate bounds based on a percentage of their current rate
		for _, hostZone := range stakeibcKeeper.GetAllHostZone(ctx) {
			if hostZone.ChainId == EvmosChainId {
				hostZone.MinRedemptionRate = EvmosOuterMinRedemptionRate
				hostZone.MinInnerRedemptionRate = EvmosInnerMinRedemptionRate
				hostZone.MaxInnerRedemptionRate = EvmosMaxRedemptionRate
				hostZone.MaxRedemptionRate = EvmosMaxRedemptionRate

				stakeibcKeeper.SetHostZone(ctx, hostZone)
			} else {
				outerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMinAdjustment)
				innerMinDelta := hostZone.RedemptionRate.Mul(RedemptionRateInnerMinAdjustment)
				innerMaxDelta := hostZone.RedemptionRate.Mul(RedemptionRateInnerMaxAdjustment)
				outerMaxDelta := hostZone.RedemptionRate.Mul(RedemptionRateOuterMaxAdjustment)

				outerMin := hostZone.RedemptionRate.Sub(outerMinDelta)
				innerMin := hostZone.RedemptionRate.Sub(innerMinDelta)
				innerMax := hostZone.RedemptionRate.Add(innerMaxDelta)
				outerMax := hostZone.RedemptionRate.Add(outerMaxDelta)

				hostZone.MinRedemptionRate = outerMin
				hostZone.MinInnerRedemptionRate = innerMin
				hostZone.MaxInnerRedemptionRate = innerMax
				hostZone.MaxRedemptionRate = outerMax

				stakeibcKeeper.SetHostZone(ctx, hostZone)
			}
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
