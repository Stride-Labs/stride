package v20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v19/x/stakeibc/keeper"
)

const (
	UpgradeName           = "v20"
	dydxCPTreasuryAddress = "dydx15ztc7xy42tn2ukkc0qjthkucw9ac63pgp70urn"
	dydxChainId           = "dydx-mainnet-1"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakeIbcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v20...")

		ctx.Logger().Info("Adding DYDX Community Pool Treasury Address...")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Write the Community Pool Treasury Address to the DYDX host_zone struct
func SetDydxCommunityPoolTreasuryAddress(ctx sdk.Context, stakeIbcKeeper stakeibckeeper.Keeper) error {

	// Get the dydx host_zone

	// Set the treasury address

	// Save the dydx host_zone

	return nil
}
