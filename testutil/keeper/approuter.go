package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	strideapp "github.com/Stride-Labs/stride/v3/app"
	"github.com/Stride-Labs/stride/v3/x/app-router/keeper"
)

func AppRouterKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	app := strideapp.InitStrideTestApp(true)
	routerKeeper := app.RouterKeeper
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &routerKeeper, ctx
}
