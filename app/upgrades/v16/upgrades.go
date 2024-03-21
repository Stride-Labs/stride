package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ratelimitkeeper "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/keeper"

	stakeibckeeper "github.com/Stride-Labs/stride/v20/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v20/x/stakeibc/types"
)

var (
	UpgradeName = "v16"

	CosmosHubChainId = "cosmoshub-4"
	CosmosHubStToken = "stuatom"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
	ratelimitKeeper ratelimitkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v16...")

		// unhalt Cosmos Hub host zone
		ctx.Logger().Info("Unhalting Cosmos Hub...")
		hostZone, found := stakeibcKeeper.GetHostZone(ctx, CosmosHubChainId)
		if !found {
			return vm, stakeibctypes.ErrHostZoneNotFound.Wrap(CosmosHubChainId)
		}

		hostZone.Halted = false
		stakeibcKeeper.SetHostZone(ctx, hostZone)

		// remove stuatom from rate limits
		ctx.Logger().Info("Removing stuatom as a blacklisted asset...")
		ratelimitKeeper.RemoveDenomFromBlacklist(ctx, CosmosHubStToken)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
