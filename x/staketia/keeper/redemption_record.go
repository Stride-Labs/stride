package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a redemption record to the store
func (k Keeper) SetRedemptionRecord(ctx sdk.Context, redemptionRecord types.RedemptionRecord) {
	// TODO [sttia]
}

// Reads a redemption record from the store
func (k Keeper) GetRedemptionRecord(ctx sdk.Context, unbondingRecordId uint64, address string) (redemptionRecord types.RedemptionRecord, found bool) {
	// TODO [sttia]
	return redemptionRecord, found
}

// Removes a redemption record from the store
func (k Keeper) RemoveRedemptionRecord(ctx sdk.Context, unbondingRecordId uint64, address string) {
	// TODO [sttia]
}

// Returns all redemption records
func (k Keeper) GetAllRedemptionRecords(ctx sdk.Context) (redemptionRecords []types.RedemptionRecord) {
	// TODO [sttia]
	return redemptionRecords
}

// Returns all redemption records for a given unbonding record
func (k Keeper) GetAllRedemptionRecordsFromUnbondingId(ctx sdk.Context, unbondingRecordId uint64) (redemptionRecords []types.RedemptionRecord) {
	// TODO [sttia]
	return redemptionRecords
}
