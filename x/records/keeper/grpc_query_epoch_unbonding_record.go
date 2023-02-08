package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/records/types"
)

func (k Keeper) EpochUnbondingRecordAll(c context.Context, req *types.QueryAllEpochUnbondingRecordRequest) (*types.QueryAllEpochUnbondingRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var epochUnbondingRecords []types.EpochUnbondingRecord
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	epochUnbondingRecordStore := prefix.NewStore(store, types.KeyPrefix(types.EpochUnbondingRecordKey))

	pageRes, err := query.Paginate(epochUnbondingRecordStore, req.Pagination, func(key []byte, value []byte) error {
		var epochUnbondingRecord types.EpochUnbondingRecord
		if err := k.Cdc.Unmarshal(value, &epochUnbondingRecord); err != nil {
			return err
		}

		epochUnbondingRecords = append(epochUnbondingRecords, epochUnbondingRecord)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllEpochUnbondingRecordResponse{EpochUnbondingRecord: epochUnbondingRecords, Pagination: pageRes}, nil
}

func (k Keeper) EpochUnbondingRecord(c context.Context, req *types.QueryGetEpochUnbondingRecordRequest) (*types.QueryGetEpochUnbondingRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	epochUnbondingRecord, found := k.GetEpochUnbondingRecord(ctx, req.EpochNumber)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetEpochUnbondingRecordResponse{EpochUnbondingRecord: epochUnbondingRecord}, nil
}
