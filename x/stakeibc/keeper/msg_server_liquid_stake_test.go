package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
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

type LiquidStakeState struct {
	depositRecordAmount int64
	hostZone            types.HostZone
}

type LiquidStakeTestCase struct {
	user         Account
	module       Account
	initialState LiquidStakeState
	validMsg     types.MsgLiquidStake
}

func (s *KeeperTestSuite) SetupLiquidStake() LiquidStakeTestCase {
	stakeAmount := uint64(1_000_000)
	initialDepositAmount := int64(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(ibcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(stAtom, 0),
	}
	s.FundAccount(user.acc, user.atomBalance)

	module := Account{
		acc:           s.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		atomBalance:   sdk.NewInt64Coin(ibcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(stAtom, 10_000_000),
	}
	s.FundModuleAccount(types.ModuleName, module.atomBalance)
	s.FundModuleAccount(types.ModuleName, module.stAtomBalance)

	hostZone := types.HostZone{
		ChainId:        "GAIA",
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
		RedemptionRate: sdk.NewDec(1.0),
	}

	epochTracker := types.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     1,
	}

	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             initialDepositAmount,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, initialDepositRecord)

	return LiquidStakeTestCase{
		user:   user,
		module: module,
		initialState: LiquidStakeState{
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
		},
		validMsg: types.MsgLiquidStake{
			Creator:   user.acc.String(),
			HostDenom: atom,
			Amount:    stakeAmount,
		},
	}
}

func (s *KeeperTestSuite) TestLiquidStakeSuccessful() {
	tc := s.SetupLiquidStake()
	user := tc.user
	module := tc.module
	msg := tc.validMsg
	stakeAmount := sdk.NewInt(int64(msg.Amount))
	initialStAtomSupply := s.App.BankKeeper.GetSupply(s.Ctx, stAtom)

	_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

	// Confirm balances
	// User IBC/UATOM balance should have DECREASED by the size of the stake
	expectedUserAtomBalance := user.atomBalance.SubAmount(stakeAmount)
	actualUserAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, ibcAtom)
	// Module IBC/UATOM balance should have INCREASED by the size of the stake
	expectedModuleAtomBalance := module.atomBalance.AddAmount(stakeAmount)
	actualModuleAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, module.acc, ibcAtom)
	// User STUATOM balance should have INCREASED by the size of the stake
	expectedUserStAtomBalance := user.stAtomBalance.AddAmount(stakeAmount)
	actualUserStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, stAtom)
	// Bank supply of STUATOM should have INCREASED by the size of the stake
	expectedBankSupply := initialStAtomSupply.AddAmount(stakeAmount)
	actualBankSupply := s.App.BankKeeper.GetSupply(s.Ctx, stAtom)

	s.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")
	s.CompareCoins(expectedUserAtomBalance, actualUserAtomBalance, "user ibc/uatom balance")
	s.CompareCoins(expectedModuleAtomBalance, actualModuleAtomBalance, "module ibc/uatom balance")
	s.CompareCoins(expectedBankSupply, actualBankSupply, "bank stuatom supply")

	// Confirm deposit record adjustment
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")

	expectedDepositRecordAmount := tc.initialState.depositRecordAmount + stakeAmount.Int64()
	actualDepositRecordAmount := records[0].Amount
	s.Require().Equal(expectedDepositRecordAmount, actualDepositRecordAmount, "deposit record amount")
}

func (s *KeeperTestSuite) TestLiquidStakeDifferentRedemptionRates() {
	tc := s.SetupLiquidStake()
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
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

		// Liquid stake for each balance and confirm stAtom minted
		startingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, stAtom).Amount.Int64()
		_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
		s.Require().NoError(err)
		endingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, stAtom).Amount.Int64()
		actualStAtomMinted := endingStAtomBalance - startingStAtomBalance

		expectedStAtomMinted := int64(float64(msg.Amount) / redemptionRateFloat)
		testDescription := fmt.Sprintf("st atom balance for redemption rate: %v", redemptionRateFloat)
		s.Require().Equal(expectedStAtomMinted, actualStAtomMinted, testDescription)
	}
}

func (s *KeeperTestSuite) TestLiquidStakeHostZoneNotFound() {
	tc := s.SetupLiquidStake()
	// Update message with invalid denom
	invalidMsg := tc.validMsg
	invalidMsg.HostDenom = "ufakedenom"
	_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "no host zone found for denom (ufakedenom): host zone not registered")
}

func (s *KeeperTestSuite) TestLiquidStakeIbcCoinParseError() {
	tc := s.SetupLiquidStake()
	// Update hostzone with denom that can't be parsed
	badHostZone := tc.initialState.hostZone
	badHostZone.IBCDenom = "ibc.0atom"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)
	_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	badCoin := fmt.Sprintf("%d%s", tc.validMsg.Amount, badHostZone.IBCDenom)
	s.Require().EqualError(err, fmt.Sprintf("failed to parse coin (%s): invalid decimal coin expression: %s", badCoin, badCoin))
}

func (s *KeeperTestSuite) TestLiquidStakeNotIbcDenom() {
	tc := s.SetupLiquidStake()
	// Update hostzone with non-ibc denom
	badDenom := "i/uatom"
	badHostZone := tc.initialState.hostZone
	badHostZone.IBCDenom = badDenom
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)
	// Fund the user with the non-ibc denom
	s.FundAccount(tc.user.acc, sdk.NewInt64Coin(badDenom, 1000000000))
	_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("denom is not an IBC token (%s): invalid token denom", badHostZone.IBCDenom))
}

func (s *KeeperTestSuite) TestLiquidStakeInsufficientBalance() {
	tc := s.SetupLiquidStake()
	// Set liquid stake amount to value greater than account balance
	invalidMsg := tc.validMsg
	balance := tc.user.atomBalance.Amount.Int64()
	invalidMsg.Amount = uint64(balance + 1000)
	_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	expectedErr := fmt.Sprintf("balance is lower than staking amount. staking amount: %d, balance: %d: insufficient funds", balance+1000, balance)
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestLiquidStakeNoEpochTracker() {
	tc := s.SetupLiquidStake()
	// Remove epoch tracker
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)
	_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("no epoch number for epoch (%s): not found", epochtypes.STRIDE_EPOCH))
}

func (s *KeeperTestSuite) TestLiquidStakeNoDepositRecord() {
	tc := s.SetupLiquidStake()
	// Remove deposit record
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx, 1)
	_, err := s.msgServer.LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("no deposit record for epoch (%d): not found", 1))
}
