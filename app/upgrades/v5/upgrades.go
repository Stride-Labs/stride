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
	icacallbacksmigration "github.com/Stride-Labs/stride/v4/x/icacallbacks/migrations/v2"
	interchainquerykeeper "github.com/Stride-Labs/stride/v4/x/interchainquery/keeper"
	recordsmigration "github.com/Stride-Labs/stride/v4/x/records/migrations/v2"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	stakeibcmigration "github.com/Stride-Labs/stride/v4/x/stakeibc/migrations/v2"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v5"
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
		// This query used an old query ID format and got stuck after the format was updated
		staleQueryId := "60b8e09dc7a65938cd6e6e5728b8aa0ca3726ffbe5511946a4f08ced316174ab"
		interchainqueryKeeper.DeleteQuery(ctx, staleQueryId)

		// Add the SafetyMaxSlashPercent param to the stakeibc param store
		stakeibcParams := stakeibcKeeper.GetParams(ctx)
		stakeibcParams.SafetyMaxSlashPercent = stakeibctypes.DefaultSafetyMaxSlashPercent
		stakeibcKeeper.SetParams(ctx, stakeibcParams)

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

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
