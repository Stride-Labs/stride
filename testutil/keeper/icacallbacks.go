package keeper

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	strideapp "github.com/Stride-Labs/stride/v9/app"
	"github.com/Stride-Labs/stride/v9/x/icacallbacks/keeper"
)

func IcacallbacksKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	app := strideapp.InitStrideTestApp(true)
	icacallbackskeeper := app.IcacallbacksKeeper
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &icacallbackskeeper, ctx
}
