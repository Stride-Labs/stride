package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/records/keeper"
)

func RecordsKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	// app := strideapp.InitStrideTestApp(true)
	// recordKeeper := app.RecordsKeeper
	// ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	// return &recordKeeper, ctx
	return nil, sdk.Context{}
}
