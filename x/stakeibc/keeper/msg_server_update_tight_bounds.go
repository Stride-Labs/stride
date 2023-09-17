package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (k msgServer) UpdateTightBounds(goCtx context.Context, msg *types.MsgUpdateTightBounds) (*types.MsgUpdateTightBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Confirm host zone exists
	_, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	// TODO: set the bounds

	return &types.MsgUpdateTightBoundsResponse{}, nil
}
