package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

// Submits an ICQ to get the validator's delegated shares
func (k msgServer) CalibrateDelegation(goCtx context.Context, msg *types.MsgCalibrateDelegation) (*types.MsgCalibrateDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound
	}

	if err := k.SubmitCalibrationICQ(ctx, hostZone, msg.Valoper); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return nil, err
	}

	return &types.MsgCalibrateDelegationResponse{}, nil
}
