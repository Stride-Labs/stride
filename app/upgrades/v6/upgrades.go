package v6

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimkeeper "github.com/Stride-Labs/stride/v9/x/claim/keeper"
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
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Reset Claims
		airdropClaimTypes := []string{"stride", "gaia", "osmosis", "juno", "stars"}
		for _, claimType := range airdropClaimTypes {
			if err := claimKeeper.ResetClaimStatus(ctx, claimType); err != nil {
				return vm, errorsmod.Wrapf(err, fmt.Sprintf("unable to reset %s claim status", claimType))
			}
		}
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
