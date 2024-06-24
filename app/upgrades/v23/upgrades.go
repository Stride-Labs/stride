package v23

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	clientkeeper "github.com/cosmos/ibc-go/v7/modules/core/02-client/keeper"

	stakeibckeeper "github.com/Stride-Labs/stride/v22/x/stakeibc/keeper"
)

var (
	UpgradeName = "v23"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	clientKeeper clientkeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v23...")

		ctx.Logger().Info("Adding wasm client...")
		AddWasmAllowedClient(ctx, clientKeeper)

		ctx.Logger().Info("Migrating trade routes...")
		MigrateTradeRoutes(ctx, stakeibcKeeper)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Add the wasm client to the IBC client's allowed clients
func AddWasmAllowedClient(ctx sdk.Context, k clientkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.AllowedClients = append(params.AllowedClients, ibcwasmtypes.Wasm)
	k.SetParams(ctx, params)
}

// Migration to deprecate the trade config
// The min transfer amount can be set from the min swap amount
func MigrateTradeRoutes(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, tradeRoute := range k.GetAllTradeRoutes(ctx) {
		tradeRoute.MinTransferAmount = tradeRoute.TradeConfig.MinSwapAmount
		k.SetTradeRoute(ctx, tradeRoute)
	}
}
