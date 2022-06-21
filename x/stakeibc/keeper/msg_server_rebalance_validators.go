package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RebalanceValidators(goCtx context.Context, msg *types.MsgRebalanceValidators) (*types.MsgRebalanceValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRebalanceValidatorsResponse{}, nil
}
