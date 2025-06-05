package v27

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v26/x/stakeibc/keeper"
)

var (
	UpgradeName = "v27"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v27...")

		// Run migrations first
		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		return versionMap, nil
	}
}
