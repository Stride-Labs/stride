package keeper

import (
	"testing"
	"time"

	strideapp "github.com/Stride-Labs/stride/v10/app"
	"github.com/Stride-Labs/stride/v10/x/epochs/keeper"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func EpochsKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	app := strideapp.InitStrideTestApp(true)
	epochsKeeper := app.EpochsKeeper
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &epochsKeeper, ctx
}
