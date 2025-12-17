package v31

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	staketiakeeper "github.com/Stride-Labs/stride/v31/x/staketia/keeper"
	staketiatypes "github.com/Stride-Labs/stride/v31/x/staketia/types"
)

var (
	UpgradeName = "v31"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v31
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	staketiaKeeper staketiakeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade %s...", UpgradeName))

		ctx.Logger().Info("Running module migrations...")
		versionMap, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info("Updating celestia unbonding period...")
		hostZone, err := staketiaKeeper.GetHostZone(ctx)
		if err != nil {
			return nil, err
		}
		hostZone.UnbondingPeriodSeconds = staketiatypes.CelestiaUnbondingPeriodSeconds
		staketiaKeeper.SetHostZone(ctx, hostZone)

		return versionMap, nil
	}
}
