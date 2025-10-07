package v9

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	claimkeeper "github.com/Stride-Labs/stride/v29/x/claim/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v29
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	claimKeeper claimkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)
		ctx.Logger().Info("Starting upgrade v9...")

		if err := AddFieldsToAirdropType(ctx, claimKeeper); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to update airdrop schema")
		}

		ctx.Logger().Info("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func AddFieldsToAirdropType(ctx sdk.Context, claimKeeper claimkeeper.Keeper) error {
	ctx.Logger().Info("Adding additional fields to airdrop struct...")

	// Get list of airdrops from claim parameters
	claimParams, err := claimKeeper.GetParams(ctx)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get claim parameters")
	}

	for _, airdrop := range claimParams.Airdrops {
		// Add the chain ID to each airdrop
		chainId, ok := AirdropChainIds[airdrop.AirdropIdentifier]
		if !ok {
			ctx.Logger().Error(fmt.Sprintf("Chain ID not specified for %s airdrop", chainId))
			continue
		}
		airdrop.ChainId = chainId

		// Enable autopilot for evmos only
		if airdrop.AirdropIdentifier == EvmosAirdropId {
			airdrop.AutopilotEnabled = true
		} else {
			airdrop.AutopilotEnabled = false
		}
	}

	// Update list of airdrops
	if err := claimKeeper.SetParams(ctx, claimParams); err != nil {
		return errorsmod.Wrapf(err, "unable to set claim parameters")
	}

	return nil
}
