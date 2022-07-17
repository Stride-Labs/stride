package keeper_test

import (
	"fmt"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"
)

const (
	atom    = "uatom"
	stAtom  = "stuatom"
	ibcAtom = "ibc/uatom"
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

type LiquidStakeTestCase struct {
	user         Account
	module       Account
	initialState State
	validMsg     stakeibc.MsgLiquidStake
}

func (suite *KeeperTestSuite) SetupLiquidStake() LiquidStakeTestCase {
	stakeAmount := int64(1_000_000)
	initialDepositAmount := int64(1_000_000)
	user := Account{
		acc:           suite.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(ibcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(stAtom, 0),
	}
	suite.FundAccount(user.acc, user.atomBalance)

	module := Account{
		acc:           suite.App.AccountKeeper.GetModuleAddress(stakeibc.ModuleName),
		atomBalance:   sdk.NewInt64Coin(ibcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(stAtom, 10_000_000),
	}
	suite.FundModuleAccount(stakeibc.ModuleName, module.atomBalance)
	suite.FundModuleAccount(stakeibc.ModuleName, module.stAtomBalance)

	hostZone := stakeibc.HostZone{
		ChainId:        "GAIA",
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
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
		Amount:             initialDepositAmount,
	}

	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, hostZone)
	suite.App.StakeibcKeeper.SetEpochTracker(suite.Ctx, epochTracker)
	suite.App.RecordsKeeper.SetDepositRecord(suite.Ctx, initialDepositRecord)

	return LiquidStakeTestCase{
		user:   user,
		module: module,
		initialState: State{
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
		},
		validMsg: stakeibc.MsgLiquidStake{
			Creator:   user.acc.String(),
			HostDenom: atom,
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
	initialStAtomSupply := suite.App.BankKeeper.GetSupply(suite.Ctx, stAtom)

	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// Confirm balances
	// User IBC/UATOM balance should have DECREASED by the size of the stake
	expectedUserAtomBalance := user.atomBalance.SubAmount(stakeAmount)
	actualUserAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, ibcAtom)
	// Module IBC/UATOM balance should have INCREASED by the size of the stake
	expectedModuleAtomBalance := module.atomBalance.AddAmount(stakeAmount)
	actualModuleAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, module.acc, ibcAtom)
	// User STUATOM balance should have INCREASED by the size of the stake
	expectedUserStAtomBalance := user.stAtomBalance.AddAmount(stakeAmount)
	actualUserStAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, stAtom)
	// Bank supply of STUATOM should have INCREASED by the size of the stake
	expectedBankSupply := initialStAtomSupply.AddAmount(stakeAmount)
	actualBankSupply := suite.App.BankKeeper.GetSupply(suite.Ctx, stAtom)

	suite.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")
	suite.CompareCoins(expectedUserAtomBalance, actualUserAtomBalance, "user ibc/uatom balance")
	suite.CompareCoins(expectedModuleAtomBalance, actualModuleAtomBalance, "module ibc/uatom balance")
	suite.CompareCoins(expectedBankSupply, actualBankSupply, "bank stuatom supply")

	// Confirm deposit record adjustment
	records := suite.App.RecordsKeeper.GetAllDepositRecord(suite.Ctx)
	suite.Require().Len(records, 1, "number of deposit records")

	expectedDepositRecordAmount := tc.initialState.depositRecordAmount + stakeAmount.Int64()
	actualDepositRecordAmount := records[0].Amount
	suite.Require().Equal(expectedDepositRecordAmount, actualDepositRecordAmount, "deposit record amount")
}

func (suite *KeeperTestSuite) TestLiquidStakeDifferentRedemptionRates() {
	tc := suite.SetupLiquidStake()
	user := tc.user
	msg := tc.validMsg

	// Loop over exchange rates: {0.2, 0.4, 0.6, ..., 2.0}
	for i := -8; i <= 10; i += 2 {
		redemptionDelta := sdk.NewDecWithPrec(1.0, 1).Mul(sdk.NewDec(int64(i))) // i = 2 => delta = 0.2
		newRedemptionRate := sdk.NewDec(1.0).Add(redemptionDelta)
		redemptionRateFloat := newRedemptionRate.MustFloat64()

		// Update rate in host zone
		hz := tc.initialState.hostZone
		hz.RedemptionRate = newRedemptionRate
		suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, hz)

		// Liquid stake for each balance and confirm stAtom minted
		startingStAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, stAtom).Amount.Int64()
		_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &msg)
		suite.Require().NoError(err)
		endingStAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, stAtom).Amount.Int64()
		actualStAtomMinted := endingStAtomBalance - startingStAtomBalance

		expectedStAtomMinted := int64(float64(msg.Amount) / redemptionRateFloat)
		testDescription := fmt.Sprintf("st atom balance for redemption rate: %v", redemptionRateFloat)
		suite.Require().Equal(expectedStAtomMinted, actualStAtomMinted, testDescription)
	}
}

func (suite *KeeperTestSuite) TestLiquidStakeHostZoneNotFound() {
	tc := suite.SetupLiquidStake()
	// Update message with invalid denom
	invalidMsg := tc.validMsg
	invalidMsg.HostDenom = "ufakedenom"
	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, "no host zone found for denom (ufakedenom): host zone not registered")
}

func (suite *KeeperTestSuite) TestLiquidStakeIbcCoinParseError() {
	tc := suite.SetupLiquidStake()
	// Update hostzone with denom that can't be parsed
	badHostZone := tc.initialState.hostZone
	badHostZone.IBCDenom = "ibc.0atom"
	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, badHostZone)
	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)

	badCoin := fmt.Sprintf("%d%s", tc.validMsg.Amount, badHostZone.IBCDenom)
	suite.Require().EqualError(err, fmt.Sprintf("failed to parse coin (%s): invalid decimal coin expression: %s", badCoin, badCoin))
}

func (suite *KeeperTestSuite) TestLiquidStakeNotIbcDenom() {
	tc := suite.SetupLiquidStake()
	// Update hostzone with non-ibc denom
	badDenom := "i/uatom"
	badHostZone := tc.initialState.hostZone
	badHostZone.IBCDenom = badDenom
	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, badHostZone)
	// Fund the user with the non-ibc denom
	suite.FundAccount(tc.user.acc, sdk.NewInt64Coin(badDenom, 1000000000))
	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)

	suite.Require().EqualError(err, fmt.Sprintf("denom is not an IBC token (%s): invalid token denom", badHostZone.IBCDenom))
}

func (suite *KeeperTestSuite) TestLiquidStakeInsufficientBalance() {
	tc := suite.SetupLiquidStake()
	// Set liquid stake amount to value greater than account balance
	invalidMsg := tc.validMsg
	balance := tc.user.atomBalance.Amount.Int64()
	invalidMsg.Amount = balance + 1000
	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	expectedErr := fmt.Sprintf("balance is lower than staking amount. staking amount: %d, balance: %d: insufficient funds", balance+1000, balance)
	suite.Require().EqualError(err, expectedErr)
}

func (suite *KeeperTestSuite) TestLiquidStakeNoEpochTracker() {
	tc := suite.SetupLiquidStake()
	// Remove epoch tracker
	suite.App.StakeibcKeeper.RemoveEpochTracker(suite.Ctx, epochtypes.STRIDE_EPOCH)
	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)

	suite.Require().EqualError(err, fmt.Sprintf("no epoch number for epoch (%s): not found", epochtypes.STRIDE_EPOCH))
}

func (suite *KeeperTestSuite) TestLiquidStakeNoDepositRecord() {
	tc := suite.SetupLiquidStake()
	// Remove deposit record
	suite.App.RecordsKeeper.RemoveDepositRecord(suite.Ctx, 1)
	_, err := suite.msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), &tc.validMsg)

	suite.Require().EqualError(err, fmt.Sprintf("no deposit record for epoch (%d): not found", 1))
}
