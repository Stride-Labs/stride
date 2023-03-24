package v8

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	autopilotkeeper "github.com/Stride-Labs/stride/v7/x/autopilot/keeper"
	autopilottypes "github.com/Stride-Labs/stride/v7/x/autopilot/types"
	claimkeeper "github.com/Stride-Labs/stride/v7/x/claim/keeper"
	claimtypes "github.com/Stride-Labs/stride/v7/x/claim/types"
)

var (
	UpgradeName             = "v8"
	EvmosAirdropDistributor = "stride10dy5pmc2fq7fnmufjfschkfrxaqnpykl6ezy5j"
	EvmosAirdropIdentifier  = "evmos"
	AirdropDuration         = time.Hour * 24 * 30 * 12 * 3 // 3 years
	ResetAirdropIdentifiers = []string{"stride", "gaia", "osmosis", "juno", "stars"}
	AirdropStartTime        = time.Date(2023, 4, 3, 16, 0, 0, 0, time.UTC).Unix() // April 3, 2023 @ 16:00 UTC (12:00 EST)
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

		// Reset Claims
		ctx.Logger().Info("Resetting airdrop claims...")
		for _, claimType := range ResetAirdropIdentifiers {
			if err := claimKeeper.ResetClaimStatus(ctx, claimType); err != nil {
				return vm, errorsmod.Wrapf(err, fmt.Sprintf("unable to reset %s claim status", claimType))
			}
		}

		// Add the evmos airdrop
		ctx.Logger().Info("Adding evmos airdrop...")
		duration := uint64(AirdropDuration.Seconds())
		if err := claimKeeper.CreateAirdropAndEpoch(ctx, EvmosAirdropDistributor, claimtypes.DefaultClaimDenom, uint64(AirdropStartTime), duration, EvmosAirdropIdentifier); err != nil {
			return vm, err
		}

		ctx.Logger().Info("Loading airdrop allocations...")
		claimKeeper.LoadAllocationData(ctx, allocations)

		// Update autopilot params
		autopilotParams := autopilottypes.DefaultParams()
		autopilotKeeper.SetParams(ctx, autopilotParams)

		ctx.Logger().Info("Running module mogrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
