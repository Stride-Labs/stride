package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) QueryExchangerate(goCtx context.Context, msg *types.MsgQueryExchangerate) (*types.MsgQueryExchangerateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgQueryExchangerateResponse{}, nil
}
