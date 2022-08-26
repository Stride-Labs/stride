package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
<<<<<<< HEAD

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
)

const (
	UpgradeName = "Upgrade to Resolve Consensus Bug"
=======
)

const (
	UpgradeName = "v2"
>>>>>>> 446e60f60974385a66496536070980b2901563a4
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
<<<<<<< HEAD
		vm[recordstypes.ModuleName] = 2
=======
>>>>>>> 446e60f60974385a66496536070980b2901563a4
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
