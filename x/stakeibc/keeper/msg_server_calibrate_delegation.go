package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Submits an ICQ to get the validator's delegated shares
func (k msgServer) CalibrateDelegation(goCtx context.Context, msg *types.MsgCalibrateDelegation) (*types.MsgCalibrateDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx
	return &types.MsgCalibrateDelegationResponse{}, nil
}
