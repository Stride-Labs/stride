package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k Keeper) LSMDeposit(c context.Context, req *types.QueryLSMDepositRequest) (*types.QueryLSMDepositResponse, error) {
	if req == nil || req.GetChainId() == "" || req.GetDenom() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var foundDeposit types.LSMTokenDeposit
	ctx := sdk.UnwrapSDKContext(c)

	deposit, found := k.GetLSMTokenDeposit(ctx, req.GetChainId(), req.GetDenom())
	if found {
		foundDeposit = deposit
	}

	return &types.QueryLSMDepositResponse{Deposit: foundDeposit}, nil
}

func (k Keeper) LSMDeposits(c context.Context, req *types.QueryLSMDepositsRequest) (*types.QueryLSMDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var deposits []types.LSMTokenDeposit
	ctx := sdk.UnwrapSDKContext(c)

	// Case 1: no chain_id was given, so we should load all deposits across all chains
	if req.GetChainId() == "" {
		deposits = k.GetAllLSMTokenDeposit(ctx)
	}

	// Case 2: chain_id is given, load all for that chain
	if req.GetChainId() != "" {
		deposits = k.GetLSMDepositsForHostZone(ctx, req.GetChainId())
	}

	// Filter for matches by hand if validator_address is given
	filtered := []types.LSMTokenDeposit{}
	if req.GetValidatorAddress() != "" {
		for _, deposit := range deposits {
			if deposit.ValidatorAddress == req.GetValidatorAddress() {
				filtered = append(filtered, deposit)
			}
		}
		deposits = filtered
	}

	// Filter for matches by hand if status is given
	filtered = []types.LSMTokenDeposit{}
	if req.GetStatus() != "" {
		for _, deposit := range deposits {
			if deposit.Status.String() == req.GetStatus() {
				filtered = append(filtered, deposit)
			}
		}
		deposits = filtered
	}

	// Be aware this could be an empty array, there might simply have been no matching deposits
	return &types.QueryLSMDepositsResponse{Deposits: deposits}, nil
}
