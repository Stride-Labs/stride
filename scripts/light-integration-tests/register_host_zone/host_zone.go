package lighttest

import (
	strideapp "github.com/Stride-Labs/stride/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BaseAppState(stride *strideapp.StrideApp, ctx sdk.Context) (*strideapp.StrideApp, sdk.Context) {
	stride.StakeibcKeeper.Logger(ctx).Info("Made it to host zone")
	return stride, ctx
}
