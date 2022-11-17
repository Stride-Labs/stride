package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple staking hooks, all hook functions are run in array sequence
type MultiStakeIBCHooks []StakeIBCHooks

func NewMultiStakeIBCHooks(hooks ...StakeIBCHooks) MultiStakeIBCHooks {
	return hooks
}

func (h MultiStakeIBCHooks) AfterLiquidStake(ctx sdk.Context, addr sdk.AccAddress) {
	for i := range h {
		h[i].AfterLiquidStake(ctx, addr)
	}
}
