package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterAccount registers an ICA account on behalf of the stakeibc module
// NOTE: this is not a standard message; only the stakeibc module should call this function. However,
// this is temporarily in the message server to facilitate easy testing and development.
// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
func (k msgServer) RegisterAccount(goCtx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.Owner); err != nil {
		return nil, err
	}
	// Return MsgRegisterAccountResponse, err
	return &types.MsgRegisterAccountResponse{}, nil
}
