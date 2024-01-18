package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a slash record to the store
func (k Keeper) SetSlashRecord(ctx sdk.Context, slashRecord types.SlashRecord) {
	// TODO [sttia]
}

// Returns all slash records
func (k Keeper) GetAllSlashRecords(ctx sdk.Context) (slashRecords []types.SlashRecord) {
	// TODO [sttia]
	return slashRecords
}
