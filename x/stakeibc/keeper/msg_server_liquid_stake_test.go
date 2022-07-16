package keeper_test

import (
	"fmt"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	_ "github.com/stretchr/testify/suite"
)

type Account struct {
	acc           sdk.AccAddress
	atomBalance   sdk.Coin
	stAtomBalance sdk.Coin
}

type State struct {
	depositRecordAmount int64
	hostZone            types.HostZone
}

type TestCase struct {
	user         Account
	module       Account
	initialState State
	validMsg     stakeibc.MsgLiquidStake
}

func (suite *KeeperTestSuite) PrintBalances() {
	stakeIbcModule := suite.App.AccountKeeper.GetModuleAddress(stakeibc.ModuleName)
	mintModule := suite.App.AccountKeeper.GetModuleAddress(minttypes.ModuleName)
	bankModule := suite.App.AccountKeeper.GetModuleAddress(banktypes.ModuleName)
	fmt.Println("Stakeibc Module Balance:", suite.App.BankKeeper.GetBalance(suite.Ctx, stakeIbcModule, "ustrd"))
	fmt.Println("Mint Module Balance:    ", suite.App.BankKeeper.GetBalance(suite.Ctx, mintModule, "ustrd"))
	fmt.Println("Bank Module Balance:    ", suite.App.BankKeeper.GetBalance(suite.Ctx, bankModule, "ustrd"))
	fmt.Println("User Balance:           ", suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], "ustrd"))
	fmt.Println("Stride Supply:          ", suite.App.BankKeeper.GetSupply(suite.Ctx, "ustrd"))
}

func (suite *KeeperTestSuite) SetupLiquidStake() TestCase {
	stakeAmount := int64(1_000_000)
	user := Account{
		acc:           suite.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 5_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 0),
	}
	suite.FundAccount(user.acc, user.atomBalance)

	module := Account{
		acc:           suite.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	suite.FundAccount(module.acc, module.atomBalance)
	suite.FundAccount(module.acc, module.stAtomBalance)

	hostZone := stakeibc.HostZone{
		ChainId:        "GAIA",
		HostDenom:      "uatom",
		IBCDenom:       "ibc/uatom",
		RedemptionRate: sdk.NewDec(1.0),
	}

	epochTracker := stakeibc.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     1,
	}

	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             0,
	}

	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, hostZone)
	suite.App.StakeibcKeeper.SetEpochTracker(suite.Ctx, epochTracker)
	suite.App.RecordsKeeper.SetDepositRecord(suite.Ctx, initialDepositRecord)

	return TestCase{
		user:   user,
		module: module,
		initialState: State{
			depositRecordAmount: 0,
			hostZone:            hostZone,
		},
		validMsg: stakeibc.MsgLiquidStake{
			Creator:   user.acc.String(),
			HostDenom: "uatom",
			Amount:    stakeAmount,
		},
	}
}

func (suite *KeeperTestSuite) TestLiquidStakeSuccessful() {
	tc := suite.SetupLiquidStake()
	user := tc.user
	module := tc.module
	msg := tc.validMsg
	stakeAmount := sdk.NewInt(msg.Amount)

	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// User IBC/UATOM balance should have DECREASED by the size of the stake
	expectedUserAtomBalance := user.atomBalance.SubAmount(stakeAmount)
	actualUserAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, "ibc/uatom")
	// Module IBC/UATOM balance should have INCREASED by the size of the stake
	expectedModuleAtomBalance := module.atomBalance.AddAmount(stakeAmount)
	actualModuleAtomBalance := suite.GetModuleBalance(stakeibc.ModuleName, "ibc/uatom")
	// User STUATOM balance should have INCREASED by the size of the stake
	expectedUserStAtomBalance := user.stAtomBalance.AddAmount(stakeAmount)
	actualUserStAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, "stuatom")

	suite.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")
	suite.CompareCoins(expectedUserAtomBalance, actualUserAtomBalance, "user ibc/uatom balance")
	suite.CompareCoins(expectedModuleAtomBalance, actualModuleAtomBalance, "module ibc/uatom balance")

	// Confirm deposit record created
}

func (suite *KeeperTestSuite) TestLiquidStakeDifferentExchangeRates() {
	// tc := suite.SetupLiquidStake()
	// msg := tc.validMsg
	// type cases struct {
	// 	exchangeRate sdk.NewDec

	// }
	// for exchangeRate := range []float64{0.25, 0.5, 0.75, 1.0, 1.25, 1.5} {
	// 	hz := tc.initialState.hostZone
	// 	hz.RedemptionRate = sdk.NewDecWithPrec(exchangeRate)
	// 	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, hz)

	// 	suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)

	// 	expectedStAtomBalance := exchangeRate * tc.user.atomBalance
	// 	suite.Require().Equal()
	// }
	// confirm balances are good
}

func (suite *KeeperTestSuite) TestLiquidStakeHostZoneNotFound() {
	tc := suite.SetupLiquidStake()
	invalidMsg := tc.validMsg
	invalidMsg.HostDenom = "ufakedenom"
	suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	// confirm host zone not found error
}

func (suite *KeeperTestSuite) TestLiquidStakeIbcCoinParseError() {
	tc := suite.SetupLiquidStake()
	// Update hostzone with denom that can't be parsed
	badHostZone := tc.initialState.hostZone
	badHostZone.IBCDenom = "ibc/u0000atom"
	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, badHostZone)
	suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)
	// confirm coin parse error
}

func (suite *KeeperTestSuite) TestLiquidStakeNotIbcDenom() {
	tc := suite.SetupLiquidStake()
	// Update hostzone with non ibc denom
	badHostZone := tc.initialState.hostZone
	badHostZone.IBCDenom = "i/uatom"
	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, badHostZone)
	suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)
	// confirm not ibc denom error
}

func (suite *KeeperTestSuite) TestLiquidStakeInsufficientBalance() {
	tc := suite.SetupLiquidStake()
	// Set liquid stake amount to value greater than account balance
	invalidMsg := tc.validMsg
	balance := tc.user.atomBalance.Amount.Int64()
	invalidMsg.Amount = balance + 1000
	suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	// confirm insufficient balance error
}

func (suite *KeeperTestSuite) TestLiquidStakeModuleSendFailure() {
	// not sure what to do here
}

func (suite *KeeperTestSuite) TestLiquidStakeMintError() {
	// not sure what to do here
}

func (suite *KeeperTestSuite) TestLiquidStakeNoEpochTracker() {
	tc := suite.SetupLiquidStake()
	// Remove epoch tracker
	suite.App.StakeibcKeeper.RemoveEpochTracker(suite.Ctx, epochtypes.STRIDE_EPOCH)
	suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)
	// confirm no epoch tracker error
}
