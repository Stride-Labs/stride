package v26

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icqoraclekeeper "github.com/Stride-Labs/stride/v27/x/icqoracle/keeper"
	icqoracletypes "github.com/Stride-Labs/stride/v27/x/icqoracle/types"
)

var (
	UpgradeName = "v26"

	OsmosisChainId      = "osmosis-1"
	OsmosisConnectionId = "connection-2"

	ICQOracleUpdateIntervalSec         = uint64(5 * 60)  // 5 min
	ICQOraclePriceExpirationTimeoutSec = uint64(15 * 60) // 15 min
)

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	icqoracleKeeper icqoraclekeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v26...")

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		// Set params after migrations
		icqoracleKeeper.SetParams(ctx, icqoracletypes.Params{
			OsmosisChainId:            OsmosisChainId,
			OsmosisConnectionId:       OsmosisConnectionId,
			UpdateIntervalSec:         ICQOracleUpdateIntervalSec,
			PriceExpirationTimeoutSec: ICQOraclePriceExpirationTimeoutSec,
		})

		return versionMap, nil
	}
}
