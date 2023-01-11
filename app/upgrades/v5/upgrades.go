package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var (
	UpgradeName = "v5"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// The following modules need state migrations as a result of a change from uints to sdk.Ints
		vm[claimtypes.ModuleName] = 2
		vm[icacallbacktypes.ModuleName] = 2
		vm[recordtypes.ModuleName] = 2
		vm[stakeibctypes.ModuleName] = 2
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
