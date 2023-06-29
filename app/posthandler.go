package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v11/x/stakeibc/keeper"
)

func NewPostHandler(stakeIbcKeeper *stakeibckeeper.Keeper) sdk.PostHandler {
	postDecorators := []sdk.PostDecorator{
		stakeibckeeper.NewStakeIbcPostDecorator(*stakeIbcKeeper),
	}
	return sdk.ChainPostDecorators(postDecorators...)
}