package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

var _ types.QueryServer = Keeper{}

// Queries the host zone struct
func (k Keeper) HostZone(c context.Context, req *types.QueryHostZoneRequest) (*types.QueryHostZoneResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return &types.QueryHostZoneResponse{}, err
	}

	return &types.QueryHostZoneResponse{HostZone: &hostZone}, nil
}

// Queries the delegation records with an optional to include archived records
func (k Keeper) DelegationRecords(c context.Context, req *types.QueryDelegationRecordsRequest) (*types.QueryDelegationRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	delegationRecords := k.GetAllActiveDelegationRecords(ctx)
	if req.IncludeArchived {
		delegationRecords = append(delegationRecords, k.GetAllArchivedDelegationRecords(ctx)...)
	}

	return &types.QueryDelegationRecordsResponse{DelegationRecords: delegationRecords}, nil
}

// Queries the unbonding records with an optional to include archived records
func (k Keeper) UnbondingRecords(c context.Context, req *types.QueryUnbondingRecordsRequest) (*types.QueryUnbondingRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	unbondingRecords := k.GetAllActiveUnbondingRecords(ctx)
	if req.IncludeArchived {
		unbondingRecords = append(unbondingRecords, k.GetAllArchivedUnbondingRecords(ctx)...)
	}

	return &types.QueryUnbondingRecordsResponse{UnbondingRecords: unbondingRecords}, nil
}

// Queries a single user redemption record
func (k Keeper) RedemptionRecord(c context.Context, req *types.QueryRedemptionRecordRequest) (*types.QueryRedemptionRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	redemptionRecord, found := k.GetRedemptionRecord(ctx, req.UnbondingRecordId, req.Address)
	if !found {
		return &types.QueryRedemptionRecordResponse{}, types.ErrRedemptionRecordNotFound.Wrapf(
			"no redemption record found for unbonding ID %d and address %s", req.UnbondingRecordId, req.Address)
	}

	return &types.QueryRedemptionRecordResponse{RedemptionRecord: &redemptionRecord}, nil
}

// Queries all redemption records with an optional filter by address
func (k Keeper) RedemptionRecords(c context.Context, req *types.QueryRedemptionRecordsRequest) (*types.QueryRedemptionRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	redemptionRecords := []types.RedemptionRecord{}

	// If they specify an address, search for that address and only return the matches
	if req.Address != "" {
		redemptionRecords := k.GetRedemptionRecordsFromAddress(ctx, req.Address)
		return &types.QueryRedemptionRecordsResponse{
			RedemptionRecords: redemptionRecords,
			Pagination:        nil,
		}, nil
	}

	// If they specify an unbonding record ID, grab just the records for that ID
	if req.UnbondingRecordId != 0 {
		redemptionRecords := k.GetRedemptionRecordsFromUnbondingId(ctx, req.UnbondingRecordId)
		return &types.QueryRedemptionRecordsResponse{
			RedemptionRecords: redemptionRecords,
			Pagination:        nil,
		}, nil
	}

	// Otherwise, return a paginated list of all redemption records
	store := ctx.KVStore(k.storeKey)
	redemptionRecordStore := prefix.NewStore(store, types.RedemptionRecordsKeyPrefix)

	pageRes, err := query.Paginate(redemptionRecordStore, req.Pagination, func(key []byte, value []byte) error {
		var redemptionRecord types.RedemptionRecord
		if err := k.cdc.Unmarshal(value, &redemptionRecord); err != nil {
			return err
		}

		redemptionRecords = append(redemptionRecords, redemptionRecord)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryRedemptionRecordsResponse{
		RedemptionRecords: redemptionRecords,
		Pagination:        pageRes,
	}, nil
}

// Queries all slash records
func (k Keeper) SlashRecords(c context.Context, req *types.QuerySlashRecordsRequest) (*types.QuerySlashRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	slashRecords := k.GetAllSlashRecords(ctx)

	return &types.QuerySlashRecordsResponse{SlashRecords: slashRecords}, nil
}
