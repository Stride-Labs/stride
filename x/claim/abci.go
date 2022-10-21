package claim

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/claim/keeper"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {

	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	// End Airdrop
	for _, airdrop := range params.Airdrops {
		goneTime := ctx.BlockTime().Sub(airdrop.AirdropStartTime)
		if goneTime > airdrop.AirdropDuration {
			// airdrop time has passed
			err := k.EndAirdrop(ctx, airdrop.AirdropIdentifier)
			if err != nil {
				panic(err)
			}
		}
	}
}
