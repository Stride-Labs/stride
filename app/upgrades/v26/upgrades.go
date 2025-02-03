package v26

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqoraclekeeper "github.com/Stride-Labs/stride/v25/x/icqoracle/keeper"
	icqoracletypes "github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

var UpgradeName = "v26"

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	icqoracleKeeper icqoraclekeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v26...")

		err := icqoracleKeeper.SetParams(ctx, icqoracletypes.Params{
			OsmosisChainId:            "osmosis-1",
			OsmosisConnectionId:       "connection-2",
			UpdateIntervalSec:         5 * 60,  // 5 min
			PriceExpirationTimeoutSec: 15 * 60, // 15 min
		})
		if err != nil {
			panic(fmt.Errorf("unable to set icqoracle params: %w", err))
		}

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
