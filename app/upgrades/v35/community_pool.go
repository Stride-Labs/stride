package v35

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	"github.com/Stride-Labs/stride/v35/utils"
)

// TransferCommunityPoolToAuthority sends every whole coin in the community pool, in every denom,
// to the team multisig. Gov can no longer pass a community pool spend once submission is closed
// and stake is gone, so the handler moves the funds while it still can. The pool is tracked in
// decimals, so the sub-unit remainder of each denom stays behind. Each denom is sent on its own
// in a cache context, so one that fails is logged and skipped with its pool balance untouched
// while the rest still transfer; the multisig can spend whatever is left later through
// MsgCommunityPoolSpend as the chain authority.
func TransferCommunityPoolToAuthority(ctx sdk.Context, k distrkeeper.Keeper) {
	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("v35: unable to read the community pool, skipping its transfer: %s", err))
		return
	}

	amount, _ := feePool.CommunityPool.TruncateDecimal()
	if amount.IsZero() {
		ctx.Logger().Info("v35: community pool holds no whole coins, nothing to transfer")
		return
	}

	// Coins are sorted by denom, so every node sends in the same sequence
	recipient := sdk.MustAccAddressFromBech32(UpgradeAuthority)
	transferred := sdk.NewCoins()
	for _, coin := range amount {
		err := utils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
			return k.DistributeFromFeePool(cacheCtx, sdk.NewCoins(coin), recipient)
		})
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: skipping community pool transfer of %s: %s", coin, err))
			continue
		}
		transferred = transferred.Add(coin)
	}

	ctx.Logger().Info(fmt.Sprintf("v35: transferred community pool (%s) to %s", transferred, UpgradeAuthority))
}
