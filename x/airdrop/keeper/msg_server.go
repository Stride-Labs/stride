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

// User transaction to claim all the pending airdrop rewards up to the current day
func (ms msgServer) ClaimDaily(goCtx context.Context, msg *types.MsgClaimDaily) (*types.MsgClaimDailyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.ClaimDaily(ctx, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimDailyResponse{}, nil
}

// User transaction to claim half of their total amount now, and forfeit the other half to be clawed back
func (ms msgServer) ClaimEarly(goCtx context.Context, msg *types.MsgClaimEarly) (*types.MsgClaimEarlyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.ClaimEarly(ctx, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimEarlyResponse{}, nil
}

// User transaction to claim and stake the full airdrop amount
// The rewards will be locked until the end of the distribution period, but will recieve rewards throughout this time
func (ms msgServer) ClaimAndStake(goCtx context.Context, msg *types.MsgClaimAndStake) (*types.MsgClaimAndStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.ClaimAndStake(ctx, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimAndStakeResponse{}, nil
}
