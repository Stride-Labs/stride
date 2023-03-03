package keeper

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

const nanosecondsInDay = 86400000000000

func (k Keeper) AddressUnbondings(c context.Context, req *types.QueryAddressUnbondings) (*types.QueryAddressUnbondingsResponse, error) {
	/*
		The function queries all the unbondings associated with a Stride address.
		This should provide more visiblity into the unbonding process for a user.
	*/

	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	var addressUnbondings []types.AddressUnbonding

	epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		for _, hostZoneUnbonding := range epochUnbondingRecord.GetHostZoneUnbondings() {
			for _, userRedemptionRecordId := range hostZoneUnbonding.GetUserRedemptionRecords() {
				userRedemptionRecordComponents := strings.Split(userRedemptionRecordId, ".")
				userRedemptionRecordAddress := userRedemptionRecordComponents[2]
				if userRedemptionRecordAddress == req.Address {
					userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, userRedemptionRecordId)
					if !found {
						continue // the record has already been claimed
					}

					// get the anticipated unbonding time
					unbondingTime := hostZoneUnbonding.UnbondingTime
					if unbondingTime == 0 {
						hostZone, found := k.GetHostZone(ctx, hostZoneUnbonding.HostZoneId)
						if !found {
							return nil, sdkerrors.ErrKeyNotFound
						}
						dayEpochTracker, found := k.GetEpochTracker(ctx, "day")
						if !found {
							return nil, sdkerrors.ErrKeyNotFound
						}
						currentDay := dayEpochTracker.EpochNumber
						daysUntilUnbonding := hostZone.UnbondingFrequency - (currentDay % hostZone.UnbondingFrequency)
						unbondingStartTime := dayEpochTracker.NextEpochStartTime + daysUntilUnbonding - 1
						unbondingDurationEstimate := hostZone.UnbondingFrequency * 7
						unbondingTime = unbondingStartTime + unbondingDurationEstimate*nanosecondsInDay
					}
					unbondingTime = unbondingTime + nanosecondsInDay
					unbondingTimeStr := time.Unix(0, int64(unbondingTime)).String()

					addressUnbonding := types.AddressUnbonding{
						Address:                req.Address,
						Receiver:               userRedemptionRecord.Receiver,
						UnbondingEstimatedTime: unbondingTimeStr,
						Amount:                 userRedemptionRecord.Amount,
						Denom:                  userRedemptionRecord.Denom,
						ClaimIsPending:         userRedemptionRecord.ClaimIsPending,
					}
					addressUnbondings = append(addressUnbondings, addressUnbonding)
				}
			}
		}
	}

	return &types.QueryAddressUnbondingsResponse{AddressUnbondings: addressUnbondings}, nil
}
