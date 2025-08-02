package v23

import (
	"context"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"

	recordskeeper "github.com/Stride-Labs/stride/v27/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
)

var (
	UpgradeName = "v23"

	CosmosChainId         = "cosmoshub-4"
	FailedLSMDepositDenom = "cosmosvaloper1yh089p0cre4nhpdqw35uzde5amg3qzexkeggdn/37467"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v23
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	clientKeeper clientkeeper.Keeper,
	recordsKeeper recordskeeper.Keeper,
	stakeibcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info("Starting upgrade v23...")

		ctx.Logger().Info("Adding wasm client...")
		AddWasmAllowedClient(ctx, clientKeeper)

		ctx.Logger().Info("Migrating trade routes...")
		MigrateTradeRoutes(ctx, stakeibcKeeper)

		ctx.Logger().Info("Resetting failed LSM detokenization record...")
		ResetLSMRecord(ctx, recordsKeeper)

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

// Reset the failed LSM detokenization record status and decrement the amount by 1
// so that it will succeed on the retry
func ResetLSMRecord(ctx sdk.Context, k recordskeeper.Keeper) {
	lsmDeposit, found := k.GetLSMTokenDeposit(ctx, CosmosChainId, FailedLSMDepositDenom)
	if !found {
		// No need to panic in this case since the difference is immaterial
		ctx.Logger().Error("Failed LSM deposit record not found")
		return
	}
	lsmDeposit.Status = recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE
	lsmDeposit.Amount = lsmDeposit.Amount.Sub(sdkmath.OneInt())
	k.SetLSMTokenDeposit(ctx, lsmDeposit)
}
