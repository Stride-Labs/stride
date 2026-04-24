package v32

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	stakeibckeeper "github.com/Stride-Labs/stride/v32/x/stakeibc/keeper"
)

var (
	UpgradeName = "v32"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v32
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	govKeeper govkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}
