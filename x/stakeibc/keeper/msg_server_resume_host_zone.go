package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ResumeHostZone(goCtx context.Context, msg *types.MsgResumeHostZone) (*types.MsgResumeHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get Host Zone
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		errMsg := fmt.Sprintf("invalid chain id, zone for %s not found", msg.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}

	// Check the zone is halted
	if !hostZone.Halted {
		errMsg := fmt.Sprintf("invalid chain id, zone for %s not halted", msg.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotHalted, errMsg)
	}

	// remove from blacklist
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	k.RatelimitKeeper.RemoveDenomFromBlacklist(ctx, stDenom)

	// Resume zone
	hostZone.Halted = false
	k.SetHostZone(ctx, hostZone)

	return &types.MsgResumeHostZoneResponse{}, nil
}
