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

func (k Keeper) DepositRecordAll(c context.Context, req *types.QueryAllDepositRecordRequest) (*types.QueryAllDepositRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var depositRecords []types.DepositRecord
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	depositRecordStore := prefix.NewStore(store, types.KeyPrefix(types.DepositRecordKey))

	pageRes, err := query.Paginate(depositRecordStore, req.Pagination, func(key []byte, value []byte) error {
		var depositRecord types.DepositRecord
		if err := k.Cdc.Unmarshal(value, &depositRecord); err != nil {
			return err
		}

		depositRecords = append(depositRecords, depositRecord)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDepositRecordResponse{DepositRecord: depositRecords, Pagination: pageRes}, nil
}

func (k Keeper) DepositRecord(c context.Context, req *types.QueryGetDepositRecordRequest) (*types.QueryGetDepositRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	depositRecord, found := k.GetDepositRecord(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetDepositRecordResponse{DepositRecord: depositRecord}, nil
}
