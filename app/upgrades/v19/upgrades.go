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
)

const (
	UpgradeName = "v19"

	StTiaDenom                = "stutia"
	CelestiaTransferChannelId = "channel-162"
	OsmosisTransferChannelId  = "channel-5"
	RateLimitDurationHours    = 24
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
	hourEpoch.EpochNumber = uint64(ctx.BlockTime().Hour())
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

	addCelestiaMsg := addRateLimitMsgTemplate
	addCelestiaMsg.ChannelId = CelestiaTransferChannelId

	addOsmosisMsg := addRateLimitMsgTemplate
	addOsmosisMsg.ChannelId = OsmosisTransferChannelId

	if err := k.AddRateLimit(ctx, &addCelestiaMsg); err != nil {
		return errorsmod.Wrapf(err, "unable to add stTIA rate limit to celestia")
	}
	if err := k.AddRateLimit(ctx, &addOsmosisMsg); err != nil {
		return errorsmod.Wrapf(err, "unable to add stTIA rate limit to celestia")
	}

	return nil
}
