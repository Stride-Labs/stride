package v19

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"

	"github.com/Stride-Labs/stride/v26/utils"
)

const (
	UpgradeName = "v19"

	StTiaDenom                = "stutia"
	RateLimitDurationHours    = 24
	CelestiaTransferChannelId = "channel-162"
	OsmosisTransferChannelId  = "channel-5"
	NeutronTransferChannelId  = "channel-123"

	WasmAdmin = "stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	ratelimitKeeper ratelimitkeeper.Keeper,
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

		// Update wasm upload permissions
		if err := SetWasmPermissions(ctx, wasmKeeper); err != nil {
			return newVm, errorsmod.Wrapf(err, "unable to set wasm permissions")
		}

		// Migrate to open sourced rate limiter
		MigrateRateLimitModule(ctx, ratelimitKeeper)

		// Add stTIA rate limits to Celestia and Osmosis
		if err := AddStTiaRateLimit(ctx, ratelimitKeeper); err != nil {
			return newVm, err
		}

		return newVm, nil
	}
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
	hourEpoch.EpochNumber = utils.IntToUint(int64(ctx.BlockTime().Hour()))
	hourEpoch.EpochStartTime = ctx.BlockTime().Truncate(time.Hour)
	hourEpoch.EpochStartHeight = ctx.BlockHeight()
	k.SetHourEpoch(ctx, hourEpoch)
}

// Add a 10% rate limit for stTIA to Celestia and Osmosis
func AddStTiaRateLimit(ctx sdk.Context, k ratelimitkeeper.Keeper) error {
	addRateLimitMsgTemplate := ratelimittypes.MsgAddRateLimit{
		Denom:          StTiaDenom,
		MaxPercentSend: sdk.NewInt(10),
		MaxPercentRecv: sdk.NewInt(10),
		DurationHours:  RateLimitDurationHours,
	}

	for _, channelId := range []string{CelestiaTransferChannelId, OsmosisTransferChannelId, NeutronTransferChannelId} {
		addMsg := addRateLimitMsgTemplate
		addMsg.ChannelId = channelId

		if err := k.AddRateLimit(ctx, &addMsg); err != nil {
			return errorsmod.Wrapf(err, "unable to add stTIA rate limit to %s", channelId)
		}
	}

	return nil
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
