package apptesting

import (
	"github.com/Stride-Labs/stride/app"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

func (s *AppTestHelper) mintAndSend(acc sdk.AccAddress, amount sdk.Coin) {
	// This mints to the bank module (even though stakeibc module is specified)
	err := s.App.BankKeeper.MintCoins(s.Ctx, stakeibc.ModuleName, sdk.NewCoins(amount))
	s.Require().NoError(err)
	// SendFromModule fails so we need to get the actual address of the bank
	bankAddress := s.App.AccountKeeper.GetModuleAddress(banktypes.ModuleName)
	err = s.App.BankKeeper.SendCoins(s.Ctx, bankAddress, acc, sdk.NewCoins(amount))
	s.Require().NoError(err)
}

func (s *AppTestHelper) FundModuleAccount(moduleName string, amount sdk.Coin) {
	// SendToModule fails so we need to get the actual module address
	moduleAddress := s.App.AccountKeeper.GetModuleAddress(moduleName)
	s.mintAndSend(moduleAddress, amount)
}

func (s *AppTestHelper) FundAccount(acc sdk.AccAddress, amount sdk.Coin) {
	s.mintAndSend(acc, amount)
}

func (s *AppTestHelper) GetModuleBalance(moduleName string, denom string) sdk.Coin {
	moduleAddress := s.App.AccountKeeper.GetModuleAddress(moduleName)
	return s.App.BankKeeper.GetBalance(s.Ctx, moduleAddress, denom)
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
