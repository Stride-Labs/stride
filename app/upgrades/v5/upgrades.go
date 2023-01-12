package v5

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimmigration "github.com/Stride-Labs/stride/v4/x/claim/migrations/v2"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	icacallbacksmigration "github.com/Stride-Labs/stride/v4/x/icacallbacks/migrations/v2"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordsmigration "github.com/Stride-Labs/stride/v4/x/records/migrations/v2"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibcmigration "github.com/Stride-Labs/stride/v4/x/stakeibc/migrations/v2"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var (
	UpgradeName = "v5"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	claimStoreKey storetypes.StoreKey,
	icacallbackStorekey storetypes.StoreKey,
	recordStoreKey storetypes.StoreKey,
	stakeibcStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// The following modules need state migrations as a result of a change from uints to sdk.Ints
		claimmigration.MigrateStore(ctx, claimStoreKey, claimtypes.ModuleCdc)
		icacallbacksmigration.MigrateStore(ctx, icacallbackStorekey, icacallbackstypes.ModuleCdc)
		recordsmigration.MigrateStore(ctx, recordStoreKey, recordstypes.ModuleCdc)
		stakeibcmigration.MigrateStore(ctx, stakeibcStoreKey, stakeibctypes.ModuleCdc)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
