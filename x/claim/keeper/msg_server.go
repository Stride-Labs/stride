package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/claim/types"
)

type msgServer struct {
	keeper Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) DepositAirdrop(goCtx context.Context, msg *types.MsgDepositAirdrop) (*types.MsgDepositAirdropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		return nil, err
	}

	err = server.keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, msg.AirdropAmount)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositAirdropResponse{}, nil
}

func (server msgServer) SetAirdropAllocations(goCtx context.Context, msg *types.MsgSetAirdropAllocations) (*types.MsgSetAirdropAllocationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	for idx, user := range msg.Users {
		record := types.ClaimRecord{
			Address:         user,
			Weight:          msg.Weights[idx],
			ActionCompleted: []bool{false, false},
		}
		server.keeper.SetClaimRecord(ctx, record)
	}

	return &types.MsgSetAirdropAllocationsResponse{}, nil
}

func (server msgServer) ClaimFreeAmount(goCtx context.Context, msg *types.MsgClaimFreeAmount) (*types.MsgClaimFreeAmountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return nil, err
	}

	coins, err := server.keeper.ClaimCoinsForAction(ctx, addr, types.ActionFree)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimFreeAmountResponse{ClaimedAmount: coins}, nil
}
