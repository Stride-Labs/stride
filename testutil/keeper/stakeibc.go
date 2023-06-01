package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/keeper"
)

func StakeibcKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	// app := strideapp.InitStrideTestApp(true)
	// stakeibcKeeper := app.StakeibcKeeper
	// ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	// return &stakeibcKeeper, ctx
	return nil, sdk.Context{}
}
