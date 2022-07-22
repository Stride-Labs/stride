package v2

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const UpgradeName = "v2"

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	gk govkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		UpdateGovParams(ctx, gk)
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func UpdateGovParams(ctx sdk.Context, gk govkeeper.Keeper) {
	params := gk.GetDepositParams(ctx)
	// TODO: Set all params of interest
	params.MaxDepositPeriod = time.Duration(123456789)
	gk.SetDepositParams(ctx, params)
}
