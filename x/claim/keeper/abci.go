package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Endblocker handler
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// Check airdrop elapsed time every 1000 blocks
	if ctx.BlockHeight()%1000 == 0 {
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
}
