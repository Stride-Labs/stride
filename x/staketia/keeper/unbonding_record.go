package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a unbonding record to the store
func (k Keeper) SetUnbondingRecord(ctx sdk.Context, unbondingRecord types.UnbondingRecord) {
	// TODO [sttia]
}

// Reads a unbonding record from the store
func (k Keeper) GetUnbondingRecord(ctx sdk.Context, recordId uint64) (unbondingRecord types.UnbondingRecord, found bool) {
	// TODO [sttia]
	return unbondingRecord, found
}

// Removes a unbonding record from the store
func (k Keeper) RemoveUnbondingRecord(ctx sdk.Context, recordId uint64) {
	// TODO [sttia]
}

// Returns all unbonding records
func (k Keeper) GetAllUnbondingRecords(ctx sdk.Context) (unbondingRecords []types.UnbondingRecord) {
	// TODO [sttia]
	return unbondingRecords
}

// Returns all unbonding records with a specific status
func (k Keeper) GetAllUnbondingRecordsByStatus(ctx sdk.Context, status types.UnbondingRecordStatus) (unbondingRecords []types.UnbondingRecord) {
	// TODO [sttia]
	return unbondingRecords
}

// Updates the status on a unbonding record
func (k Keeper) UpdateUnbondingRecordStatus(ctx sdk.Context, recordId uint64, status types.UnbondingRecordStatus) {
	// TODO [sttia]
}

// Gets the queue unbonding record (there should only be one)
func (k Keeper) GetQueueUnbondingRecord(ctx sdk.Context) (unbondingRecord types.UnbondingRecord, err error) {
	// TODO [sttia]
	return unbondingRecord, nil
}
