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

	ctx := sdk.UnwrapSDKContext(c)

	deposit, found := k.GetLSMTokenDeposit(ctx, req.GetChainId(), req.GetDenom())
	if !found {
		return nil, status.Error(codes.NotFound, "LSM deposit not found")
	}

	return &types.QueryLSMDepositResponse{Deposit: deposit}, nil
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

	// Filter for matches by hand if validator_address or status optional filters are given
	filtered := []types.LSMTokenDeposit{}
	filterByValidator := req.GetValidatorAddress() != ""
	filterByStatus := req.GetStatus() != ""
	for _, deposit := range deposits {
		validatorMatch := !filterByValidator || (deposit.ValidatorAddress == req.GetValidatorAddress())
		statusMatch := !filterByStatus || (deposit.Status.String() == req.GetStatus())
		if validatorMatch && statusMatch {
			filtered = append(filtered, deposit)
		}
	}
	deposits = filtered

	// Be aware this could be an empty array, there may have been no deposits matching given filters
	return &types.QueryLSMDepositsResponse{Deposits: deposits}, nil
}
