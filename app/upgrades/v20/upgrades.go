package v20

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v19/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

const (
	UpgradeName           = "v20"
	dydxCPTreasuryAddress = "dydx15ztc7xy42tn2ukkc0qjthkucw9ac63pgp70urn"
	dydxChainId           = "dydx-mainnet-1"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	stakeIbcKeeper stakeibckeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting upgrade v20...")

		ctx.Logger().Info("Adding DYDX Community Pool Treasury Address...")
		if err := SetDydxCommunityPoolTreasuryAddress(ctx, stakeIbcKeeper); err != nil {
			return nil, errorsmod.Wrapf(err, "unable to set dydx community pool treasury address")
		}

		return nil, nil
	}
}

// Write the Community Pool Treasury Address to the DYDX host_zone struct
func SetDydxCommunityPoolTreasuryAddress(ctx sdk.Context, stakeIbcKeeper stakeibckeeper.Keeper) error {

	// Get the dydx host_zone
	hostZone, found := stakeIbcKeeper.GetHostZone(ctx, dydxChainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "user redemption record not found")
	}

	// Set the treasury address
	// TODO replace with writing to store (types don't work with v19/20)
	hostZone.communityPoolTreasuryAddress = dydxCPTreasuryAddress

	// Save the dydx host_zone
	stakeIbcKeeper.SetHostZone(ctx, hostZone)

	return nil
}
