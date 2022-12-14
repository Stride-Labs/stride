package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v5"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v4
// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	claimStoreKey storetypes.StoreKey,
	recordStoreKey storetypes.StoreKey,
	stakeibcStoreKey storetypes.StoreKey,
	cdc codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		newVm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVm, err
		}

		// migrate store
		err = MigrateStore(ctx, claimStoreKey, recordStoreKey, stakeibcStoreKey, cdc)
		if err == nil {
			return nil, err
		}
		return newVm, nil
	}
}
