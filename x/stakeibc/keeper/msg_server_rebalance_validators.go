package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (k msgServer) RebalanceValidators(goCtx context.Context, msg *types.MsgRebalanceValidators) (*types.MsgRebalanceValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("RebalanceValidators executing %v", msg))

	if err := k.RebalanceDelegationsForHostZone(ctx, msg.HostZone); err != nil {
		return nil, err
	}
	return &types.MsgRebalanceValidatorsResponse{}, nil
}
