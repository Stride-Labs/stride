package v5

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimmigration "github.com/Stride-Labs/stride/v9/x/claim/migrations/v2"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
	icacallbacksmigration "github.com/Stride-Labs/stride/v9/x/icacallbacks/migrations/v2"
	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	interchainquerykeeper "github.com/Stride-Labs/stride/v9/x/interchainquery/keeper"
	recordsmigration "github.com/Stride-Labs/stride/v9/x/records/migrations/v2"
	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	stakeibcmigration "github.com/Stride-Labs/stride/v9/x/stakeibc/migrations/v2"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v5"

	// This query used an old query ID format and got stuck after the format was updated
	StaleQueryId = "60b8e09dc7a65938cd6e6e5728b8aa0ca3726ffbe5511946a4f08ced316174ab"
)

// Helper function to log the migrated modules consensus versions
func logModuleMigration(ctx sdk.Context, versionMap module.VersionMap, moduleName string) {
	currentVersion := versionMap[moduleName]
	ctx.Logger().Info(fmt.Sprintf("migrating module %s from version %d to version %d", moduleName, currentVersion-1, currentVersion))
}

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
		currentVersions := mm.GetVersionMap()

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
		logModuleMigration(ctx, currentVersions, claimtypes.ModuleName)
		if err := claimmigration.MigrateStore(ctx, claimStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate claim store")
		}

		logModuleMigration(ctx, currentVersions, icacallbacktypes.ModuleName)
		if err := icacallbacksmigration.MigrateStore(ctx, icacallbackStorekey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate icacallbacks store")
		}

		logModuleMigration(ctx, currentVersions, recordtypes.ModuleName)
		if err := recordsmigration.MigrateStore(ctx, recordStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate records store")
		}

		logModuleMigration(ctx, currentVersions, stakeibctypes.ModuleName)
		if err := stakeibcmigration.MigrateStore(ctx, stakeibcStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate stakeibc store")
		}

		// `RunMigrations` (below) checks the old consensus version of each module (found in
		// the store) and compares it against the updated consensus version in the binary
		// If the old and new consensus versions are not the same, it attempts to call that
		// module's migration function that must be registered ahead of time
		//
		// Since the migrations above were executed directly (instead of being registered
		// and invoked through a Migrator), we need to set the module versions in the versionMap
		// to the new version, to prevent RunMigrations from attempting to re-run each migrations
		vm[claimtypes.ModuleName] = currentVersions[claimtypes.ModuleName]
		vm[icacallbacktypes.ModuleName] = currentVersions[icacallbacktypes.ModuleName]
		vm[recordtypes.ModuleName] = currentVersions[recordtypes.ModuleName]
		vm[stakeibctypes.ModuleName] = currentVersions[stakeibctypes.ModuleName]

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
