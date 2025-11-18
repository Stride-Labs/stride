package v16

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v30/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v30/x/stakeibc/types"
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
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
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
