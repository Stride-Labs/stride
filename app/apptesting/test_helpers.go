package apptesting

import (
	"github.com/Stride-Labs/stride/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

const Bech32Prefix = "stride"

type AppTestHelper struct {
	suite.Suite

	App         *app.StrideApp
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
}

func (s *AppTestHelper) Setup() {
	checkTx := false
	s.App = app.InitTestApp(checkTx)
	s.Ctx = s.App.BaseApp.NewContext(checkTx, tmtypes.Header{Height: 1, ChainID: "stride-1"})
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = CreateRandomAccounts(3)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(Bech32Prefix, Bech32Prefix+sdk.PrefixPublic)
	config.Seal()
}

func (s *AppTestHelper) FundModuleAccount(moduleName string, amount string) {
	coins := s.StringToCoin(amount)
	simapp.FundModuleAccount(s.App.BankKeeper, s.Ctx, moduleName, coins)
}

func (s *AppTestHelper) FundAcc(acc sdk.AccAddress, moduleName, amount string) {
	coins := s.StringToCoin(amount)
	err := s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, moduleName, acc, coins)
	// err := simapp.FundAccount(s.App.BankKeeper, s.Ctx, acc, coins)
	s.Require().NoError(err)
}

func (s *AppTestHelper) StringToCoin(amount string) sdk.Coins {
	coins, err := sdk.ParseCoinNormalized(amount)
	s.Require().NoError(err)
	return sdk.NewCoins(coins)
}

func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}
