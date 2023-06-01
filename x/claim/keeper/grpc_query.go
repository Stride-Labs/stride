package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v9/x/claim/types"
)

var _ types.QueryServer = Keeper{}

// Params returns balances of the distributor account
func (k Keeper) DistributorAccountBalance(c context.Context, req *types.QueryDistributorAccountBalanceRequest) (*types.QueryDistributorAccountBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	bal, err := k.GetDistributorAccountBalance(ctx, req.AirdropIdentifier)
	if err != nil {
		return nil, err
	}
	return &types.QueryDistributorAccountBalanceResponse{DistributorAccountBalance: sdk.NewCoins(bal)}, nil
}

// Params returns params of the claim module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// ClaimRecord returns user claim record by address and airdrop identifier
func (k Keeper) ClaimRecord(
	goCtx context.Context,
	req *types.QueryClaimRecordRequest,
) (*types.QueryClaimRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	claimRecord, err := k.GetClaimRecord(ctx, addr, req.AirdropIdentifier)
	if err != nil {
		return nil, err
	}
	return &types.QueryClaimRecordResponse{ClaimRecord: claimRecord}, nil
}

// ClaimableForAction returns claimable amount per action
func (k Keeper) ClaimableForAction(
	goCtx context.Context,
	req *types.QueryClaimableForActionRequest,
) (*types.QueryClaimableForActionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	coins, err := k.GetClaimableAmountForAction(ctx, addr, req.Action, req.AirdropIdentifier, false)
	if err != nil {
		return nil, err
	}

	return &types.QueryClaimableForActionResponse{Coins: coins}, nil
}

// TotalClaimable returns total claimable amount for user
func (k Keeper) TotalClaimable(
	goCtx context.Context,
	req *types.QueryTotalClaimableRequest,
) (*types.QueryTotalClaimableResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	coins, err := k.GetUserTotalClaimable(ctx, addr, req.AirdropIdentifier, req.IncludeClaimed)
	if err != nil {
		return nil, err
	}

	return &types.QueryTotalClaimableResponse{Coins: coins}, nil
}

// UserVestings returns all vestings for user
func (k Keeper) UserVestings(
	goCtx context.Context,
	req *types.QueryUserVestingsRequest,
) (*types.QueryUserVestingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	vestings, spendableCoins := k.GetUserVestings(ctx, addr)

	return &types.QueryUserVestingsResponse{
		SpendableCoins: spendableCoins,
		Periods:        vestings,
	}, nil
}

// ClaimStatus returns all vestings for user
func (k Keeper) ClaimStatus(
	goCtx context.Context,
	req *types.QueryClaimStatusRequest,
) (*types.QueryClaimStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	claimStatus, err := k.GetClaimStatus(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &types.QueryClaimStatusResponse{ClaimStatus: claimStatus}, nil
}

// ClaimMetadata returns all vestings for user
func (k Keeper) ClaimMetadata(
	goCtx context.Context,
	req *types.QueryClaimMetadataRequest,
) (*types.QueryClaimMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	claimMetadata := k.GetClaimMetadata(ctx)

	return &types.QueryClaimMetadataResponse{ClaimMetadata: claimMetadata}, nil
}
