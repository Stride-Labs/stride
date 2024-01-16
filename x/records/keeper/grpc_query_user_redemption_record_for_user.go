package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v17/x/records/types"
)

func (k Keeper) UserRedemptionRecordForUser(c context.Context, req *types.QueryAllUserRedemptionRecordForUserRequest) (*types.QueryAllUserRedemptionRecordForUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var userRedemptionRecords []types.UserRedemptionRecord

	ctx := sdk.UnwrapSDKContext(c)

	// limit loop to 50 records for performance
	var loopback uint64
	loopback = req.Limit
	if loopback > 50 {
		loopback = 50
	}
	var i uint64
	for i = 0; i < loopback; i++ {
		if i > req.Day {
			// we have reached the end of the records
			break
		}
		currentDay := req.Day - i
		// query the user redemption record for the current day
		userRedemptionRecord, found := k.GetUserRedemptionRecord(ctx, types.UserRedemptionRecordKeyFormatter(req.ChainId, currentDay, req.Address))
		if !found {
			continue
		}
		userRedemptionRecords = append(userRedemptionRecords, userRedemptionRecord)
	}

	return &types.QueryAllUserRedemptionRecordForUserResponse{UserRedemptionRecord: userRedemptionRecords}, nil
}
