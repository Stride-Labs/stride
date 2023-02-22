package v6

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimkeeper "github.com/Stride-Labs/stride/v5/x/claim/keeper"
)

// Note: ensure these values are properly set before running upgrade
var (
	UpgradeName = "v6"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	claimKeeper claimkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Reset Claims
		if err := claimKeeper.ResetClaimStatus(ctx, "stride"); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to reset stride claim status")
		}
		if err := claimKeeper.ResetClaimStatus(ctx, "gaia"); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to reset gaia claim status")
		}
		if err := claimKeeper.ResetClaimStatus(ctx, "osmo"); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to reset osmo claim status")
		}
		if err := claimKeeper.ResetClaimStatus(ctx, "juno"); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to reset juno claim status")
		}
		if err := claimKeeper.ResetClaimStatus(ctx, "stars"); err != nil {
			return vm, sdkerrors.Wrapf(err, "unable to reset stars claim status")
		}
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
