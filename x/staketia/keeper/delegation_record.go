package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a delegation record to the store
func (k Keeper) SetDelegationRecord(ctx sdk.Context, delegationRecord types.DelegationRecord) {
	// TODO [sttia]
}

// Reads a delegation record from the store
func (k Keeper) GetDelegationRecord(ctx sdk.Context, recordId uint64) (delegationRecord types.DelegationRecord, found bool) {
	// TODO [sttia]
	return delegationRecord, found
}

// Removes a delegation record from the store
func (k Keeper) RemoveDelegationRecord(ctx sdk.Context, recordId uint64) {
	// TODO [sttia]
}

// Returns all delegation records
func (k Keeper) GetAllDelegationRecords(ctx sdk.Context) (delegationRecords []types.DelegationRecord) {
	// TODO [sttia]
	return delegationRecords
}

// Updates the status on a delegation record
func (k Keeper) UpdateDelegationRecordStatus(ctx sdk.Context, recordId uint64, status types.DelegationRecordStatus) {
	// TODO [sttia]
}
