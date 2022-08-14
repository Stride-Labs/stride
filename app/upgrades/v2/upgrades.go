package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
)

const (
	UpgradeName = "Upgrade to Resolve Consensus Bug"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm[recordstypes.ModuleName] = 2
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
