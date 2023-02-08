package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	strideapp "github.com/Stride-Labs/stride/v4/app"
	"github.com/Stride-Labs/stride/v4/x/epochs/keeper"
)

func EpochsKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	app := strideapp.InitStrideTestApp(true)
	epochsKeeper := app.EpochsKeeper
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &epochsKeeper, ctx
}
