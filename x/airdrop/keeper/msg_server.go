package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
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

// Admin transaction to create a new airdrop
func (ms msgServer) CreateAirdrop(goCtx context.Context, msg *types.MsgCreateAirdrop) (*types.MsgCreateAirdropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId); found {
		return nil, types.ErrAirdropAlreadyExists.Wrapf("airdrop %s", msg.AirdropId)
	}

	airdrop := types.Airdrop{
		Id:                    msg.AirdropId,
		DistributionStartDate: msg.DistributionStartDate,
		DistributionEndDate:   msg.DistributionEndDate,
		ClawbackDate:          msg.ClawbackDate,
		ClaimTypeDeadlineDate: msg.ClaimTypeDeadlineDate,
		EarlyClaimPenalty:     msg.EarlyClaimPenalty,
		ClaimAndStakeBonus:    msg.ClaimAndStakeBonus,
		DistributionAddress:   msg.DistributionAddress,
	}
	ms.Keeper.SetAirdrop(ctx, airdrop)

	return &types.MsgCreateAirdropResponse{}, nil
}

// Admin transaction to update an existing airdrop
func (ms msgServer) UpdateAirdrop(goCtx context.Context, msg *types.MsgUpdateAirdrop) (*types.MsgUpdateAirdropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId); !found {
		return nil, types.ErrAirdropNotFound.Wrapf("airdrop %s", msg.AirdropId)
	}

	airdrop := types.Airdrop{
		Id:                    msg.AirdropId,
		DistributionStartDate: msg.DistributionStartDate,
		DistributionEndDate:   msg.DistributionEndDate,
		ClawbackDate:          msg.ClawbackDate,
		ClaimTypeDeadlineDate: msg.ClaimTypeDeadlineDate,
		EarlyClaimPenalty:     msg.EarlyClaimPenalty,
		ClaimAndStakeBonus:    msg.ClaimAndStakeBonus,
		DistributionAddress:   msg.DistributionAddress,
	}
	ms.Keeper.SetAirdrop(ctx, airdrop)

	return &types.MsgUpdateAirdropResponse{}, nil
}

// Admin transaction to add user allocations
func (ms msgServer) AddAllocations(goCtx context.Context, msg *types.MsgAddAllocations) (*types.MsgAddAllocationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId); !found {
		return nil, types.ErrAirdropNotFound.Wrapf("airdrop %s", msg.AirdropId)
	}

	for _, rawAllocation := range msg.Allocations {
		if _, found := ms.Keeper.GetUserAllocation(ctx, msg.AirdropId, rawAllocation.UserAddress); found {
			return nil, types.ErrUserAllocationAlreadyExists.Wrapf("user %s", rawAllocation.UserAddress)
		}

		userAllocation := types.UserAllocation{
			AirdropId:        msg.AirdropId,
			Address:          rawAllocation.UserAddress,
			Claimed:          sdkmath.ZeroInt(),
			Allocations:      rawAllocation.Allocations,
			ClaimType:        types.UNSPECIFIED,
			ValidatorAddress: "",
		}
		ms.Keeper.SetUserAllocation(ctx, userAllocation)
	}

	return &types.MsgAddAllocationsResponse{}, nil
}

// Admin transaction to update a user's allocations
func (ms msgServer) UpdateUserAllocation(goCtx context.Context, msg *types.MsgUpdateUserAllocation) (*types.MsgUpdateUserAllocationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId); !found {
		return nil, types.ErrAirdropNotFound.Wrapf("airdrop %s", msg.AirdropId)
	}

	userAllocation, found := ms.Keeper.GetUserAllocation(ctx, msg.AirdropId, msg.UserAddress)
	if !found {
		return nil, types.ErrUserAllocationNotFound.Wrapf("user %s", msg.UserAddress)
	}
	userAllocation.Allocations = msg.Allocations
	ms.Keeper.SetUserAllocation(ctx, userAllocation)

	return &types.MsgUpdateUserAllocationResponse{}, nil
}

// Admin address to link a stride and non-stride address, merging their allocations
func (ms msgServer) LinkAddresses(goCtx context.Context, msg *types.MsgLinkAddresses) (*types.MsgLinkAddressesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	return &types.MsgLinkAddressesResponse{}, nil
}
