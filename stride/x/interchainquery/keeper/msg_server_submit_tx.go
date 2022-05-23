package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SubmitTx(goCtx context.Context, msg *types.MsgSubmitTx) (*types.MsgSubmitTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO(TEST-50) Handling the message
	_ = ctx

	return &types.MsgSubmitTxResponse{}, nil
}
