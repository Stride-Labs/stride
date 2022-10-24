package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Endblocker handler
func (k Keeper) EndBlocker(ctx sdk.Context) {
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
