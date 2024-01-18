package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a host zone to the store
func (k Keeper) SetHostZone(ctx sdk.Context, hostZone types.HostZone) {
	// TODO [sttia]
}

// Reads a host zone from the store
// There should always be a host zone, so this should error if it is not found
func (k Keeper) GetHostZone(ctx sdk.Context) (hostZone types.HostZone, err error) {
	// TODO [sttia]
	return hostZone, nil
}
