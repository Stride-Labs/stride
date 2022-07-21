package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterInterchainAccount(goCtx context.Context, msg *types.MsgRegisterInterchainAccount) (*types.MsgRegisterInterchainAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRegisterInterchainAccountResponse{}, nil
}
