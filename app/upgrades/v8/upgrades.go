package v8

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/Stride-Labs/stride/v9/utils"
	autopilotkeeper "github.com/Stride-Labs/stride/v9/x/autopilot/keeper"
	autopilottypes "github.com/Stride-Labs/stride/v9/x/autopilot/types"
	claimkeeper "github.com/Stride-Labs/stride/v9/x/claim/keeper"
	"github.com/Stride-Labs/stride/v9/x/claim/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
)

var (
	UpgradeName             = "v8"
	EvmosAirdropDistributor = "stride10dy5pmc2fq7fnmufjfschkfrxaqnpykl6ezy5j"
	EvmosAirdropIdentifier  = "evmos"
	EvmosChainId            = "evmos_9001-2"
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

		// Reset Claims
		ctx.Logger().Info("Resetting airdrop claims...")
		for _, claimType := range ResetAirdropIdentifiers {
			if err := claimKeeper.ResetClaimStatus(ctx, claimType); err != nil {
				return vm, errorsmod.Wrapf(err, fmt.Sprintf("unable to reset %s claim status", claimType))
			}
		}

		// Delete all unofficial airdrops from the store so they don't conflict with the evmos addition
		claimParams, err := claimKeeper.GetParams(ctx)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "unable to get claim parameters")
		}

		updatedAirdropList := []*types.Airdrop{}
		for _, existingAirdrop := range claimParams.Airdrops {
			if utils.ContainsString(ResetAirdropIdentifiers, existingAirdrop.AirdropIdentifier) {
				updatedAirdropList = append(updatedAirdropList, existingAirdrop)
			}
		}
		updatedClaimsParams := claimtypes.Params{
			Airdrops: updatedAirdropList,
		}

		if err := claimKeeper.SetParams(ctx, updatedClaimsParams); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to remove unofficial airdrops")
		}

		// Add the evmos airdrop
		ctx.Logger().Info("Adding evmos airdrop...")
		duration := uint64(AirdropDuration.Seconds())
		if err := claimKeeper.CreateAirdropAndEpoch(ctx, types.MsgCreateAirdrop{
			Distributor:      EvmosAirdropDistributor,
			Identifier:       EvmosAirdropIdentifier,
			ChainId:          EvmosChainId,
			Denom:            claimtypes.DefaultClaimDenom,
			StartTime:        uint64(AirdropStartTime.Unix()),
			Duration:         duration,
			AutopilotEnabled: true,
		}); err != nil {
			return vm, err
		}

		ctx.Logger().Info("Loading airdrop allocations...")
		claimKeeper.LoadAllocationData(ctx, allocations)

		// Update autopilot params
		autopilotParams := autopilottypes.DefaultParams()
		autopilotKeeper.SetParams(ctx, autopilotParams)

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
