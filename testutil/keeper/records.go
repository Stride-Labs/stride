package keeper

import (
	"testing"
	"time"

	strideapp "github.com/Stride-Labs/stride/app"
	"github.com/Stride-Labs/stride/x/records/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func RecordsKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	checkTx := false
	app := strideapp.InitTestApp(checkTx)
	recordKeeper := app.RecordsKeeper
	ctx := app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &recordKeeper, ctx
}
