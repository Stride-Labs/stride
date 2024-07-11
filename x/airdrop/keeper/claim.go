package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Claim(ctx sdk.Context, claimer string) error {
	// TODO[airdrop] implement logic

	return nil
}

func (k Keeper) ClaimAndStake(ctx sdk.Context, claimer string) error {
	// TODO[airdrop] implement logic

	return nil
}

func (k Keeper) ClaimEarly(ctx sdk.Context, claimer string) error {
	// TODO[airdrop] implement logic

	return nil
}
