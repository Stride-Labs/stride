package v8

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v8/x/autopilot/keeper"
	autopilottypes "github.com/Stride-Labs/stride/v8/x/autopilot/types"
	claimkeeper "github.com/Stride-Labs/stride/v8/x/claim/keeper"
)

var (
	UpgradeName             = "v8"
	EvmosAirdropDistributor = "stride10dy5pmc2fq7fnmufjfschkfrxaqnpykl6ezy5j"
	EvmosAirdropIdentifier  = "evmos"
	AirdropDuration         = time.Hour * 24 * 30 * 12 * 3 // 3 years
	ResetAirdropIdentifiers = []string{"stride", "gaia", "osmosis", "juno", "stars"}
	AirdropStartTime        = time.Date(2023, 4, 3, 16, 0, 0, 0, time.UTC) // April 3, 2023 @ 16:00 UTC (12:00 EST)
)

// CreateUpgradeHandler creates an SDK upgrade handler for v8
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	claimKeeper claimkeeper.Keeper,
	autopilotKeeper autopilotkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v8...")

		// Update autopilot params
		autopilotParams := autopilottypes.DefaultParams()
		autopilotKeeper.SetParams(ctx, autopilotParams)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
