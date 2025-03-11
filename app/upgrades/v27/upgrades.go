package v27

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	consumerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/consumer/keeper"
)

var UpgradeName = "v27"

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	consumerKeeper consumerkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		// https://github.com/cosmos/interchain-security/blob/v6.4.1/UPGRADING.md#consumer
		InitializeConsumerId(ctx, consumerKeeper)

		return versionMap, nil
	}
}

// InitializeConsumerId sets the consumer Id parameter in the consumer module,
// to the consumer id for which the consumer is registered on the provider chain.
// The consumer id can be obtained in by querying the provider, e.g. by using the
// QueryConsumerIdFromClientId query.
func InitializeConsumerId(ctx sdk.Context, consumerKeeper consumerkeeper.Keeper) error {
	params, err := consumerKeeper.GetParams(ctx)
	if err != nil {
		return err
	}
	params.ConsumerId = ConsumerId
	return consumerKeeper.SetParams(ctx, params)
}
