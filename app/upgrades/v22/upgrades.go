package v22

import (
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	stakeibckeeper "github.com/Stride-Labs/stride/v26/x/stakeibc/keeper"
)

var (
	UpgradeName = "v22"

	OsmosisChainId = "osmosis-1"
	DydxChainId    = "dydx-mainnet-1"

	MaxMessagesPerIcaByHost = map[string]uint64{
		OsmosisChainId: 150,
		DydxChainId:    60,
	}
)

// CreateUpgradeHandler creates an SDK upgrade handler for v22
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v22...")

		ctx.Logger().Info("Migrating host zones...")
		MigrateHostZones(ctx, stakeibcKeeper)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Add the MaxMessagesPerIcaTx parameter to each host zone
func MigrateHostZones(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		maxMessages, ok := MaxMessagesPerIcaByHost[hostZone.ChainId]
		if !ok {
			maxMessages = stakeibckeeper.DefaultMaxMessagesPerIcaTx
		}
		hostZone.MaxMessagesPerIcaTx = maxMessages
		k.SetHostZone(ctx, hostZone)
	}
}
