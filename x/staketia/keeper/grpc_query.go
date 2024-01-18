package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

var _ types.QueryServer = Keeper{}

// Queries the host zone struct
func (k Keeper) HostZone(c context.Context, req *types.QueryHostZoneRequest) (*types.QueryHostZoneResponse, error) {
	// TODO [sttia]
	return &types.QueryHostZoneResponse{}, nil
}

// Queries the delegation records with an optional to include archived records
func (k Keeper) DelegationRecords(c context.Context, req *types.QueryDelegationRecordsRequest) (*types.QueryDelegationRecordsResponse, error) {
	// TODO [sttia]
	return &types.QueryDelegationRecordsResponse{}, nil
}

// Queries the unbonding records with an optional to include archived records
func (k Keeper) UnbondingRecords(c context.Context, req *types.QueryUnbondingRecordsRequest) (*types.QueryUnbondingRecordsResponse, error) {
	// TODO [sttia]
	return &types.QueryUnbondingRecordsResponse{}, nil
}

// Queries a single user redemption record
func (k Keeper) RedemptionRecord(c context.Context, req *types.QueryRedemptionRecord) (*types.QueryRedemptionRecordResponse, error) {
	// TODO [sttia]
	return &types.QueryRedemptionRecordResponse{}, nil
}

// Queries all redemption records with an optional filter by address
func (k Keeper) AllRedemptionRecords(c context.Context, req *types.QueryAllRedemptionRecords) (*types.QueryAllRedemptionRecordsResponse, error) {
	// TODO [sttia]
	return &types.QueryAllRedemptionRecordsResponse{}, nil
}

// Queries all slash records
func (k Keeper) SlashRecords(c context.Context, req *types.QuerySlashRecords) (*types.QuerySlashRecordsResponse, error) {
	// TODO [sttia]
	return &types.QuerySlashRecordsResponse{}, nil
}
