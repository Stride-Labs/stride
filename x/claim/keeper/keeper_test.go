package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/Stride-Labs/stride/app"
	"github.com/Stride-Labs/stride/x/claim/types"
	minttypes "github.com/Stride-Labs/stride/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	// querier sdk.Querier
	app *app.StrideApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.InitStrideTestApp(true)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	// Mint coins to airdrop module
	suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000))))
	suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, types.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000))))

	airdropStartTime := time.Now()

	err := suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
		AirdropStartTime: airdropStartTime,
		AirdropDuration:  types.DefaultAirdropDuration,
		ClaimDenom:       sdk.DefaultBondDenom,
	})
	if err != nil {
		panic(err)
	}

	suite.ctx = suite.ctx.WithBlockTime(airdropStartTime)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
