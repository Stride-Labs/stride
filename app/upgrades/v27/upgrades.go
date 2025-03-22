package v27

import (
	"context"
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
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
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
