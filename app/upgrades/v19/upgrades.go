package v19

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	ratelimitkeeper "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/keeper"
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v4/x/ccv/consumer/keeper"
)

const (
	UpgradeName = "v19"

	WasmAdmin = "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	consumerKeeper ccvconsumerkeeper.Keeper,
	ratelimitKeeper ratelimitkeeper.Keeper,
	wasmKeeper wasmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v19...")

		// Run module migrations first to add wasm to the store
		ctx.Logger().Info("Running module migrations...")
		newVm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return newVm, errorsmod.Wrapf(err, "unable to run module migrations")
		}

		// Migrate ICS to v4
		if err := MigrateICSOutstandingDowntime(ctx, consumerKeeper); err != nil {
			return newVm, errorsmod.Wrapf(err, "unable to migrate ICS to v14")
		}

		// Migrate to open sourced rate limiter
		MigrateRateLimitModule(ctx, ratelimitKeeper)

		// Update wasm upload permissions
		if err := SetWasmPermissions(ctx, wasmKeeper); err != nil {
			return newVm, errorsmod.Wrapf(err, "unable to set wasm permissions")
		}

		return newVm, nil
	}
}

// Migrates Outstanding Downtime for upgrade to ICS v4
// https://github.com/cosmos/interchain-security/blob/release/v4.0.x/UPGRADING.md#v40x
func MigrateICSOutstandingDowntime(ctx sdk.Context, ck ccvconsumerkeeper.Keeper) error {
	ctx.Logger().Info("Migrating ICS outstanding downtime...")

	downtimes := ck.GetAllOutstandingDowntimes(ctx)
	for _, od := range downtimes {
		consAddr, err := sdk.ConsAddressFromBech32(od.ValidatorConsensusAddress)
		if err != nil {
			return err
		}
		ck.DeleteOutstandingDowntime(ctx, consAddr)
	}

	ctx.Logger().Info("Finished ICS outstanding downtime")
	return nil
}

// Migrate the rate limit module to the open sourced version
// The module has the same store key so all the rate limit types
// can remain unchanged
// The only required change is to create the new epoch type
// that's used instead of the epochs module
func MigrateRateLimitModule(ctx sdk.Context, k ratelimitkeeper.Keeper) {
	// Initialize the hour epoch so that the epoch number matches
	// the current hour and the start time is precisely on the hour
	genesisState := ratelimittypes.DefaultGenesis()
	hourEpoch := genesisState.HourEpoch
	hourEpoch.EpochNumber = uint64(ctx.BlockTime().Hour())
	hourEpoch.EpochStartTime = ctx.BlockTime().Truncate(time.Hour)
	hourEpoch.EpochStartHeight = ctx.BlockHeight()
	k.SetHourEpoch(ctx, hourEpoch)
}

// Update wasm params so that contracts can only be uploaded through governance
func SetWasmPermissions(ctx sdk.Context, wk wasmkeeper.Keeper) error {
	wasmParams := wk.GetParams(ctx)
	wasmParams.CodeUploadAccess = wasmtypes.AccessConfig{
		Permission: wasmtypes.AccessTypeAnyOfAddresses,
		Addresses:  []string{WasmAdmin},
	}
	wasmParams.InstantiateDefaultPermission = wasmtypes.AccessTypeNobody
	if err := wk.SetParams(ctx, wasmParams); err != nil {
		return err
	}
	return nil
}
