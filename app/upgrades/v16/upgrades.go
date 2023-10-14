package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v15/x/stakeibc/keeper"
)

var (
	UpgradeName = "v16"

	CosmosHubChainId = "cosmoshub-4"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v15...")

		// unhalt Cosmos Hub host zone
		ctx.Logger().Info("Unhalting Cosmos Hub...")
		hostZone, found := stakeibcKeeper.GetHostZone(ctx, CosmosHubChainId)
		if !found {
			ctx.Logger().Error("Cosmos Hub host zone not found!")
		} else {
			hostZone.Halted = false
			stakeibcKeeper.SetHostZone(ctx, hostZone)
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
