package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

	airdropDistributor, err := server.keeper.GetAirdropDistributor(ctx)
	if err != nil {
		return nil, err
	}

	if !addr.Equals(airdropDistributor) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distributor address")
	}

	airdropDenom, err := server.keeper.GetAirdropClaimDenom(ctx)
	if err != nil {
		return nil, err
	}

	for _, airdropCoin := range msg.AirdropAmount {
		if airdropCoin.Denom != airdropDenom {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid airdrop denom")
		}
	}

	err = server.keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, msg.AirdropAmount)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositAirdropResponse{}, nil
}

func (server msgServer) SetAirdropAllocations(goCtx context.Context, msg *types.MsgSetAirdropAllocations) (*types.MsgSetAirdropAllocationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	totalWeight := sdk.NewDec(0)
	records := []types.ClaimRecord{}

	airdropDistributor, err := server.keeper.GetAirdropDistributor(ctx)
	if err != nil {
		return nil, err
	}

	addr, err := sdk.AccAddressFromBech32(msg.Allocator)
	if err != nil {
		return nil, err
	}

	if !addr.Equals(airdropDistributor) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distributor address")
	}

	for idx, user := range msg.Users {
		record := types.ClaimRecord{
			Address:         user,
			Weight:          msg.Weights[idx],
			ActionCompleted: []bool{false, false, false},
		}

		records = append(records, record)
		totalWeight = totalWeight.Add(msg.Weights[idx])
	}

	server.keeper.SetTotalWeight(ctx, totalWeight)
	server.keeper.SetClaimRecords(ctx, records)

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
