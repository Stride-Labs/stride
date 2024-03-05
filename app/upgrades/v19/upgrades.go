package v19

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v18/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

const (
	UpgradeName = "v19"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	stakeibcKeeper stakeibckeeper.Keeper,
	wasmKeeper wasmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v19...")

		// Run module migrations first to add wasm to the store
		ctx.Logger().Info("Running module migrations...")
		newVm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVm, err
		}

		// Update wasm params so that contracts can only be uploaded through governance
		wasmParams := wasmKeeper.GetParams(ctx)
		wasmParams.CodeUploadAccess = wasmtypes.AccessConfig{
			Permission: wasmtypes.AccessTypeNobody,
			Addresses:  []string{},
		}
		wasmParams.InstantiateDefaultPermission = wasmtypes.AccessTypeNobody
		if err := wasmKeeper.SetParams(ctx, wasmParams); err != nil {
			return newVm, err
		}

		// Migate the stakeibc params to add the MaxICAMessagesPerTx parameter
		MigrateStakeibcParams(ctx, stakeibcKeeper)

		return newVm, nil
	}
}

// Migrate the stakeibc params to add the MaxICAMessagesPerTx parameter
//
// NOTE: If a parameter is added, the old params cannot be unmarshalled
// to the new schema. To get around this, we have to set each parameter explicitly
// Considering all mainnet stakeibc params are set to the default, we can just use that
func MigrateStakeibcParams(ctx sdk.Context, k stakeibckeeper.Keeper) {
	params := stakeibctypes.DefaultParams()
	k.SetParams(ctx, params)
}
