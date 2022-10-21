package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

var distributors map[string]sdk.AccAddress

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.InitStrideTestApp(true)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "stride-1", Time: time.Now().UTC()})

	// Initiate a distributor account
	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
	distributors = make(map[string]sdk.AccAddress)
	distributors[types.DefaultAirdropIdentifier] = addr1

	// Mint coins to airdrop module
	suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000))))
	suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr1, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000000))))

	airdropStartTime := time.Now()
	err := suite.app.ClaimKeeper.SetParams(suite.ctx, types.Params{
		Airdrops: []*types.Airdrop{
			{
				AirdropIdentifier:  types.DefaultAirdropIdentifier,
				AirdropStartTime:   airdropStartTime,
				AirdropDuration:    types.DefaultAirdropDuration,
				ClaimDenom:         sdk.DefaultBondDenom,
				DistributorAddress: distributors[types.DefaultAirdropIdentifier].String(),
			},
		},
	})
	if err != nil {
		panic(err)
	}

	suite.ctx = suite.ctx.WithBlockTime(airdropStartTime)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
