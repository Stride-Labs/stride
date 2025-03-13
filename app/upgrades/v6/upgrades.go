package v6

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	claimkeeper "github.com/Stride-Labs/stride/v26/x/claim/keeper"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v6"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v6
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	claimKeeper claimkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)

		// Reset Claims
		airdropClaimTypes := []string{"stride", "gaia", "osmosis", "juno", "stars"}
		for _, claimType := range airdropClaimTypes {
			if err := claimKeeper.ResetClaimStatus(ctx, claimType); err != nil {
				return vm, errorsmod.Wrapf(err, "unable to reset %s claim status", claimType)
			}
		}
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
