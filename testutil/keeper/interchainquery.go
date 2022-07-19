package keeper

import (
	"testing"
	"time"

	strideapp "github.com/Stride-Labs/stride/app"
	"github.com/Stride-Labs/stride/x/interchainquery/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func InterchainqueryKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	checkTx := false
	app := strideapp.InitTestApp(checkTx)
	icqKeeper := app.InterchainqueryKeeper
	ctx := app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &icqKeeper, ctx
}
