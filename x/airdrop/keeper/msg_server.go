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

	err := ms.Keeper.ClaimDaily(ctx, msg.AirdropId, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimDailyResponse{}, nil
}

// User transaction to claim half of their total amount now, and forfeit the other half to be clawed back
func (ms msgServer) ClaimEarly(goCtx context.Context, msg *types.MsgClaimEarly) (*types.MsgClaimEarlyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.ClaimEarly(ctx, msg.AirdropId, msg.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimEarlyResponse{}, nil
}

// Admin transaction to create a new airdrop
func (ms msgServer) CreateAirdrop(goCtx context.Context, msg *types.MsgCreateAirdrop) (*types.MsgCreateAirdropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId); found {
		return nil, types.ErrAirdropAlreadyExists.Wrapf("airdrop %s", msg.AirdropId)
	}

	airdrop := types.Airdrop{
		Id:                    msg.AirdropId,
		RewardDenom:           msg.RewardDenom,
		DistributionStartDate: msg.DistributionStartDate,
		DistributionEndDate:   msg.DistributionEndDate,
		ClawbackDate:          msg.ClawbackDate,
		ClaimTypeDeadlineDate: msg.ClaimTypeDeadlineDate,
		EarlyClaimPenalty:     msg.EarlyClaimPenalty,
		DistributorAddress:    msg.DistributorAddress,
		AllocatorAddress:      msg.AllocatorAddress,
		LinkerAddress:         msg.LinkerAddress,
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
		RewardDenom:           msg.RewardDenom,
		DistributionStartDate: msg.DistributionStartDate,
		DistributionEndDate:   msg.DistributionEndDate,
		ClawbackDate:          msg.ClawbackDate,
		ClaimTypeDeadlineDate: msg.ClaimTypeDeadlineDate,
		EarlyClaimPenalty:     msg.EarlyClaimPenalty,
		DistributorAddress:    msg.DistributorAddress,
		AllocatorAddress:      msg.AllocatorAddress,
		LinkerAddress:         msg.LinkerAddress,
	}
	ms.Keeper.SetAirdrop(ctx, airdrop)

	return &types.MsgUpdateAirdropResponse{}, nil
}

// Admin transaction to add user allocations
func (ms msgServer) AddAllocations(goCtx context.Context, msg *types.MsgAddAllocations) (*types.MsgAddAllocationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	airdrop, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId)
	if !found {
		return nil, types.ErrAirdropNotFound.Wrapf("airdrop %s", msg.AirdropId)
	}
	if msg.Admin != airdrop.AllocatorAddress {
		return nil, types.ErrInvalidAdminAddress.Wrapf("user allocations can only be added by the allocator admin")
	}

	periodLengthSeconds := ms.Keeper.GetParams(ctx).PeriodLengthSeconds
	expectedDays := airdrop.GetAirdropPeriods(periodLengthSeconds)

	for _, rawAllocation := range msg.Allocations {
		if _, found := ms.Keeper.GetUserAllocation(ctx, msg.AirdropId, rawAllocation.UserAddress); found {
			return nil, types.ErrUserAllocationAlreadyExists.Wrapf("user %s", rawAllocation.UserAddress)
		}

		if len(rawAllocation.Allocations) != int(expectedDays) {
			return nil, types.ErrInvalidAllocationListLength.Wrapf("expected %d, provided %d",
				expectedDays, len(rawAllocation.Allocations))
		}

		userAllocation := types.UserAllocation{
			AirdropId:   msg.AirdropId,
			Address:     rawAllocation.UserAddress,
			Claimed:     sdkmath.ZeroInt(),
			Forfeited:   sdkmath.ZeroInt(),
			Allocations: rawAllocation.Allocations,
		}
		ms.Keeper.SetUserAllocation(ctx, userAllocation)
	}

	return &types.MsgAddAllocationsResponse{}, nil
}

// Admin transaction to update a user's allocations
func (ms msgServer) UpdateUserAllocation(goCtx context.Context, msg *types.MsgUpdateUserAllocation) (*types.MsgUpdateUserAllocationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	airdrop, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId)
	if !found {
		return nil, types.ErrAirdropNotFound.Wrapf("airdrop %s", msg.AirdropId)
	}
	if msg.Admin != airdrop.AllocatorAddress {
		return nil, types.ErrInvalidAdminAddress.Wrapf("user allocation updates can only be performed by the allocator admin")
	}

	userAllocation, found := ms.Keeper.GetUserAllocation(ctx, msg.AirdropId, msg.UserAddress)
	if !found {
		return nil, types.ErrUserAllocationNotFound.Wrapf("user %s", msg.UserAddress)
	}

	if len(msg.Allocations) != len(userAllocation.Allocations) {
		return nil, types.ErrInvalidAllocationListLength.Wrapf("current allocations length: %d, provided length: %d",
			len(userAllocation.Allocations), len(msg.Allocations))
	}

	userAllocation.Allocations = msg.Allocations
	ms.Keeper.SetUserAllocation(ctx, userAllocation)

	return &types.MsgUpdateUserAllocationResponse{}, nil
}

// Admin address to link a stride and non-stride address, merging their allocations
func (ms msgServer) LinkAddresses(goCtx context.Context, msg *types.MsgLinkAddresses) (*types.MsgLinkAddressesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	airdrop, found := ms.Keeper.GetAirdrop(ctx, msg.AirdropId)
	if !found {
		return nil, types.ErrAirdropNotFound.Wrapf("airdrop %s", msg.AirdropId)
	}
	if msg.Admin != airdrop.LinkerAddress {
		return nil, types.ErrInvalidAdminAddress.Wrapf("linking can only be performed by the linkor admin")
	}

	if err := ms.Keeper.LinkAddresses(ctx, msg.AirdropId, msg.StrideAddress, msg.HostAddress); err != nil {
		return nil, err
	}

	return &types.MsgLinkAddressesResponse{}, nil
}
