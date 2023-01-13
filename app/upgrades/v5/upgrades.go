package v5

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimmigration "github.com/Stride-Labs/stride/v4/x/claim/migrations/v2"
	icacallbacksmigration "github.com/Stride-Labs/stride/v4/x/icacallbacks/migrations/v2"
	recordsmigration "github.com/Stride-Labs/stride/v4/x/records/migrations/v2"
	stakeibcmigration "github.com/Stride-Labs/stride/v4/x/stakeibc/migrations/v2"
)

var (
	UpgradeName = "v5"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	claimStoreKey storetypes.StoreKey,
	icacallbackStorekey storetypes.StoreKey,
	recordStoreKey storetypes.StoreKey,
	stakeibcStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// The following modules need state migrations as a result of a change from uints to sdk.Ints
		claimmigration.MigrateStore(ctx, claimStoreKey, cdc)
		icacallbacksmigration.MigrateStore(ctx, icacallbackStorekey, cdc)
		recordsmigration.MigrateStore(ctx, recordStoreKey, cdc)
		stakeibcmigration.MigrateStore(ctx, stakeibcStoreKey, cdc)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
