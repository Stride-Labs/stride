package v29

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consumerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/consumer/keeper"
)

var (
	UpgradeName = "v29"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v29
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	consumerKeeper consumerkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Updating CCV params...")
		DisableCcvRewards(ctx, consumerKeeper)

		return versionMap, nil
	}
}

func DisableCcvRewards(ctx sdk.Context, ck consumerkeeper.Keeper) {
	params := ck.GetConsumerParams(ctx)
	params.ConsumerRedistributionFraction = "1.0"
	ck.SetParams(ctx, params)
}
