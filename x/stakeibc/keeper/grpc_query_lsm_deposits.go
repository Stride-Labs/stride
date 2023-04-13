package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k Keeper) LSMDeposits(c context.Context, req *types.QueryLSMDepositsRequest) (*types.QueryLSMDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var allDeposits []types.LSMTokenDeposit
	ctx := sdk.UnwrapSDKContext(c)

	// chainId must be included, but denom and status are optional felds

	// Case 1: They are asking for a specific chainId + denom --> GetLSMTokenDeposit
	// 	 if they included a status and a denom.... ignore the status and search for denom
	if req.GetDenom() != "" {
		deposit, found := k.GetLSMTokenDeposit(ctx, req.GetChainId(), req.GetDenom())
		if found {
			allDeposits = append(allDeposits, deposit)
		}
	}

	// Case 2: They are asking for all deposits with chainId + status --> Get
	//   if they included a status and denom, already handled above, denom is missing for this branch
	if req.GetXStatus() != nil && req.GetDenom() == "" {
		deposits := k.GetLSMDepositsForHostZoneWithStatus(ctx, req.GetChainId(), req.GetStatus())
		if len(deposits) > 0 {
			allDeposits = append(allDeposits, deposits...)
		}
	}

	// Case 3: They are looking for all deposits with chainId -->
	//   both status and denom optional arguments have to be left out for this branch
	if req.GetXStatus() == nil && req.GetDenom() == "" {
		deposits := k.GetLSMDepositsForHostZone(ctx, req.GetChainId())
		if len(deposits) > 0 {
			allDeposits = append(allDeposits, deposits...)
		}
	}

	// Be aware this could be an empty array, there might simply have been no matching deposits
	return &types.QueryLSMDepositsResponse{Deposits: allDeposits}, nil
}
