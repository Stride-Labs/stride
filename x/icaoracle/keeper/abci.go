package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker of icaoracle module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// TODO: Submit ICAs for each metric
}
