package v5

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimmigration "github.com/Stride-Labs/stride/v4/x/claim/migrations/v2"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	icacallbacksmigration "github.com/Stride-Labs/stride/v4/x/icacallbacks/migrations/v2"
	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	interchainquerykeeper "github.com/Stride-Labs/stride/v4/x/interchainquery/keeper"
	recordsmigration "github.com/Stride-Labs/stride/v4/x/records/migrations/v2"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	stakeibcmigration "github.com/Stride-Labs/stride/v4/x/stakeibc/migrations/v2"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v5"

	// This query used an old query ID format and got stuck after the format was updated
	StaleQueryId = "60b8e09dc7a65938cd6e6e5728b8aa0ca3726ffbe5511946a4f08ced316174ab"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	interchainqueryKeeper interchainquerykeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
	claimStoreKey storetypes.StoreKey,
	icacallbackStorekey storetypes.StoreKey,
	recordStoreKey storetypes.StoreKey,
	stakeibcStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Remove authz from store as it causes an issue with state sync
		delete(vm, authz.ModuleName)

		// Remove a stale query from the interchainquery store
		interchainqueryKeeper.DeleteQuery(ctx, StaleQueryId)

		// To add the SafetyMaxSlashPercent param to the stakeibc param store,
		// we just re-initialize all the params with their default value
		// The alternative would be to migrate the entire paramstore, but since each param is still
		// set to it's default value, this is a safer/less error-prone approach
		stakeibcKeeper.SetParams(ctx, stakeibctypes.DefaultParams())

		// The following modules need state migrations as a result of a change from uints to sdk.Ints
		//    - claim
		//    - icacallbacks
		//    - records
		//    - stakeibc
		if err := claimmigration.MigrateStore(ctx, claimStoreKey, cdc); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to migrate claim store")
		}
		if err := icacallbacksmigration.MigrateStore(ctx, icacallbackStorekey, cdc); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to migrate icacallbacks store")
		}
		if err := recordsmigration.MigrateStore(ctx, recordStoreKey, cdc); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to migrate records store")
		}
		if err := stakeibcmigration.MigrateStore(ctx, stakeibcStoreKey, cdc); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to migrate stakeibc store")
		}

		// Update "from" module version in map to current version to prevent migrator from
		// re-running the above migrations
		currentVersions := mm.GetVersionMap()
		vm[claimtypes.ModuleName] = currentVersions[claimtypes.ModuleName]
		vm[icacallbacktypes.ModuleName] = currentVersions[icacallbacktypes.ModuleName]
		vm[recordtypes.ModuleName] = currentVersions[recordtypes.ModuleName]
		vm[stakeibctypes.ModuleName] = currentVersions[stakeibctypes.ModuleName]

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
