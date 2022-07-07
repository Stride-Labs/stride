package keeper

import (
	"testing"
	"time"

	strideapp "github.com/Stride-Labs/stride/app"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func StakeibcKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	isCheckTx := false
	app := strideapp.Setup(isCheckTx)
	stakeibcKeeper := app.StakeibcKeeper
	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &stakeibcKeeper, ctx
}
