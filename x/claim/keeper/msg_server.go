package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
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

func (server msgServer) SetAirdropAllocations(goCtx context.Context, msg *types.MsgSetAirdropAllocations) (*types.MsgSetAirdropAllocationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	records := []types.ClaimRecord{}
	totalWeight, err := server.keeper.GetTotalWeight(ctx, msg.AirdropIdentifier)
	if err != nil {
		return nil, err
	}

	airdropDistributor, err := server.keeper.GetAirdropDistributor(ctx, msg.AirdropIdentifier)
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

	users, weights := server.keeper.GetUnallocatedUsers(ctx, msg.AirdropIdentifier, msg.Users, msg.Weights)
	for idx, user := range users {
		record := types.ClaimRecord{
			Address:           user,
			Weight:            weights[idx],
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: msg.AirdropIdentifier,
		}

		records = append(records, record)
		totalWeight = totalWeight.Add(weights[idx])
	}

	server.keeper.SetTotalWeight(ctx, totalWeight, msg.AirdropIdentifier)
	err = server.keeper.SetClaimRecords(ctx, records)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetAirdropAllocationsResponse{}, nil
}

func (server msgServer) ClaimFreeAmount(goCtx context.Context, msg *types.MsgClaimFreeAmount) (*types.MsgClaimFreeAmountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return nil, err
	}

	coins, err := server.keeper.ClaimAllCoinsForAction(ctx, addr, types.ACTION_FREE)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimFreeAmountResponse{ClaimedAmount: coins}, nil
}

func (server msgServer) CreateAirdrop(goCtx context.Context, msg *types.MsgCreateAirdrop) (*types.MsgCreateAirdropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		return nil, err
	}

	airdrop := server.keeper.GetAirdropByDistributor(ctx, msg.Distributor)
	if airdrop != nil {
		return nil, types.ErrDistributorAlreadyExists
	}

	err = server.keeper.CreateAirdropAndEpoch(ctx, msg.Distributor, msg.Denom, msg.StartTime, msg.Duration, msg.Identifier)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateAirdropResponse{}, nil
}

func (server msgServer) DeleteAirdrop(goCtx context.Context, msg *types.MsgDeleteAirdrop) (*types.MsgDeleteAirdropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		return nil, err
	}

	distributor, err := server.keeper.GetAirdropDistributor(ctx, msg.Identifier)
	if err != nil {
		return nil, err
	}

	if !addr.Equals(distributor) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distributor address")
	}

	err = server.keeper.DeleteAirdropAndEpoch(ctx, msg.Identifier)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeleteAirdropResponse{}, nil
}
