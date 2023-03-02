package v7

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	epochsmigration "github.com/Stride-Labs/stride/v7/x/epochs/migrations/v7"
	epochstypes "github.com/Stride-Labs/stride/v7/x/epochs/types"

	mintmigration "github.com/Stride-Labs/stride/v7/x/mint/migrations/v7"
	minttypes "github.com/Stride-Labs/stride/v7/x/mint/types"

	stakeibcmigration "github.com/Stride-Labs/stride/v7/x/stakeibc/migrations/v7"
	stakeibctypes "github.com/Stride-Labs/stride/v7/x/stakeibc/types"

	interchainaccountsmigration "github.com/Stride-Labs/stride/v7/x/interchainaccounts/migrations/v7"
	interchainaccountstypes "github.com/Stride-Labs/stride/v7/x/interchainaccounts/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v7"
)

// Helper function to log the migrated modules consensus versions
func logModuleMigration(ctx sdk.Context, versionMap module.VersionMap, moduleName string) {
	currentVersion := versionMap[moduleName]
	ctx.Logger().Info(fmt.Sprintf("migrating module %s from version %d to version %d", moduleName, currentVersion-1, currentVersion))
}

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	icacallbackStoreKey storetypes.StoreKey,
	stakeibcStoreKey storetypes.StoreKey,
	interchainaccountsStoreKey storetypes.StoreKey,
	epochsStoreKey storetypes.StoreKey,
	mintStoreKey storetypes.StoreKey,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		currentVersions := mm.GetVersionMap()

		// TODO: diversification

		// The following modules need state migrations as a result of a change from uints to sdk.Ints
		//    - stakeibc
		//    - epochs
		//    - mint
		//    - interchainaccounts

		// StakeIbc migrations
		//  - juno unbonding freq update: set the stakeibc module's host_zone juno freq to 5
		//  - add min/max redemption rates to existing host zones
		logModuleMigration(ctx, currentVersions, stakeibctypes.ModuleName)
		if err := stakeibcmigration.MigrateStore(ctx, stakeibcStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate stakeibc store")
		}

		// For the rate limit module, add the hourly epoch to x/epochs
		logModuleMigration(ctx, currentVersions, epochstypes.ModuleName)
		if err := epochsmigration.MigrateStore(ctx, epochsStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate epochs store")
		}
		// TODO: for the rate limit module, add store key

		// For the mint module, update the inflation
		logModuleMigration(ctx, currentVersions, minttypes.ModuleName)
		if err := mintmigration.MigrateStore(ctx, mintStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate mint store")
		}

		logModuleMigration(ctx, currentVersions, interchainaccountstypes.ModuleName)
		if err := interchainaccountsmigration.MigrateStore(ctx, interchainaccountsStoreKey, cdc); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to migrate records store")
		}

		// `RunMigrations` (below) checks the old consensus version of each module (found in
		// the store) and compares it against the updated consensus version in the binary
		// If the old and new consensus versions are not the same, it attempts to call that
		// module's migration function that must be registered ahead of time
		//
		// Since the migrations above were executed directly (instead of being registered
		// and invoked through a Migrator), we need to set the module versions in the versionMap
		// to the new version, to prevent RunMigrations from attempting to re-run each migrations
		vm[stakeibctypes.ModuleName] = currentVersions[stakeibctypes.ModuleName]
		vm[epochstypes.ModuleName] = currentVersions[epochstypes.ModuleName]
		vm[minttypes.ModuleName] = currentVersions[minttypes.ModuleName]
		vm[interchainaccountstypes.ModuleName] = currentVersions[interchainaccountstypes.ModuleName]

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
