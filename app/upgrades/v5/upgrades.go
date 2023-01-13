package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	interchainquerykeeper "github.com/Stride-Labs/stride/v4/x/interchainquery/keeper"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v5"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	interchainqueryKeeper interchainquerykeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Remove authz from store as it causes an issue with state sync
		delete(vm, authz.ModuleName)

		// Remove a stale query from the interchainquery store
		// This query used an old query ID format and got stuck after the format was updated
		staleQueryId := "60b8e09dc7a65938cd6e6e5728b8aa0ca3726ffbe5511946a4f08ced316174ab"
		interchainqueryKeeper.DeleteQuery(ctx, staleQueryId)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
