package apptesting

import (
	"github.com/Stride-Labs/stride/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(Bech32Prefix, Bech32Prefix+sdk.PrefixPublic)
	config.Seal()

	checkTx := false
	s.App = app.InitTestApp(checkTx)
	s.Ctx = s.App.BaseApp.NewContext(checkTx, tmtypes.Header{Height: 1, ChainID: "stride-1"})
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = CreateRandomAccounts(3)
}

func (s *AppTestHelper) FundModuleAccount(moduleName string, amount sdk.Coin) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, moduleName, sdk.NewCoins(amount))
	s.Require().NoError(err)
}

func (s *AppTestHelper) FundAccount(acc sdk.AccAddress, amount sdk.Coin) {
	amountCoins := sdk.NewCoins(amount)
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, amountCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, amountCoins)
	s.Require().NoError(err)
}

func (s *AppTestHelper) CompareCoins(expectedCoin sdk.Coin, actualCoin sdk.Coin, msg string) {
	s.Require().Equal(expectedCoin.Amount.Int64(), actualCoin.Amount.Int64(), msg)
}

func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}
