package keeper

import (
	"testing"
	"time"

	strideapp "github.com/Stride-Labs/stride/v10/app"
	"github.com/Stride-Labs/stride/v10/x/claim/keeper"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ClaimKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	app := strideapp.InitStrideTestApp(true)
	claimKeeper := app.ClaimKeeper
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	return &claimKeeper, ctx
}
