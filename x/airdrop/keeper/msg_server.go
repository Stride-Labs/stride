package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// User transaction to claim all the pending daily airdrop rewards
func (ms msgServer) Claim(goCtx context.Context, msg *types.MsgClaim) (*types.MsgClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.Claim(ctx, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimResponse{}, nil
}

// User transaction to claim half of their total amount now, and forfeit the
// other half to be clawed back
func (ms msgServer) ClaimEarly(goCtx context.Context, msg *types.MsgClaimEarly) (*types.MsgClaimEarlyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.ClaimEarly(ctx, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimEarlyResponse{}, nil
}

// User transaction to claim and automatically lock stake their whole airdrop
// amount now. The funds can be unstaked by the user once the airdrop is over
func (ms msgServer) ClaimAndStake(goCtx context.Context, msg *types.MsgClaimAndStake) (*types.MsgClaimAndStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.ClaimAndStake(ctx, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimAndStakeResponse{}, nil
}
