package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) ClaimUndelegatedTokens(goCtx context.Context, msg *types.MsgClaimUndelegatedTokens) (*types.MsgClaimUndelegatedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		errMsg := fmt.Sprintf("Host zone %s not found", msg.HostZone)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrInvalidHostZone, errMsg)
	}
	// claimableRecords :=

	return &types.MsgClaimUndelegatedTokensResponse{}, nil
}
