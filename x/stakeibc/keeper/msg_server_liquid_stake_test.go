package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type Account struct {
	acc           sdk.AccAddress
	atomBalance   sdk.Coin
	stAtomBalance sdk.Coin
}

type LiquidStakeState struct {
	depositRecordAmount sdk.Int
	hostZone            types.HostZone
}

type LiquidStakeTestCase struct {
	user         Account
	zoneAccount  Account
	initialState LiquidStakeState
	validMsg     stakeibctypes.MsgLiquidStake
}

func (s *KeeperTestSuite) SetupLiquidStake() LiquidStakeTestCase {
	stakeAmount := sdk.NewInt(1_000_000)
	initialDepositAmount := sdk.NewInt(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(IbcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(StAtom, 0),
	}
	s.FundAccount(user.acc, user.atomBalance)

	zoneAddress := types.NewZoneAddress(HostChainId)

	zoneAccount := Account{
		acc:           zoneAddress,
		atomBalance:   sdk.NewInt64Coin(IbcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(StAtom, 10_000_000),
	}
	s.FundAccount(zoneAccount.acc, zoneAccount.atomBalance)
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Address:        zoneAddress.String(),
	}

	epochTracker := stakeibctypes.EpochTracker{
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
		user:        user,
		zoneAccount: zoneAccount,
		initialState: LiquidStakeState{
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
		},
		validMsg: stakeibctypes.MsgLiquidStake{
			Creator:   user.acc.String(),
			HostDenom: Atom,
			Amount:    stakeAmount,
		},
	}
}

func (s *KeeperTestSuite) TestLiquidStake_Successful() {
	tc := s.SetupLiquidStake()
	user := tc.user
	zoneAccount := tc.zoneAccount
	msg := tc.validMsg
	initialStAtomSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

	// Confirm balances
	// User IBC/UATOM balance should have DECREASED by the size of the stake
	expectedUserAtomBalance := user.atomBalance.SubAmount(msg.Amount)
	actualUserAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, IbcAtom)
	// zoneAccount IBC/UATOM balance should have INCREASED by the size of the stake
	expectedzoneAccountAtomBalance := zoneAccount.atomBalance.AddAmount(msg.Amount)
	actualzoneAccountAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount.acc, IbcAtom)
	// User STUATOM balance should have INCREASED by the size of the stake
	expectedUserStAtomBalance := user.stAtomBalance.AddAmount(msg.Amount)
	actualUserStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom)
	// Bank supply of STUATOM should have INCREASED by the size of the stake
	expectedBankSupply := initialStAtomSupply.AddAmount(msg.Amount)
	actualBankSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)

	s.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")
	s.CompareCoins(expectedUserAtomBalance, actualUserAtomBalance, "user ibc/uatom balance")
	s.CompareCoins(expectedzoneAccountAtomBalance, actualzoneAccountAtomBalance, "zoneAccount ibc/uatom balance")
	s.CompareCoins(expectedBankSupply, actualBankSupply, "bank stuatom supply")

	// Confirm deposit record adjustment
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")

	expectedDepositRecordAmount := tc.initialState.depositRecordAmount.Add(msg.Amount)
	actualDepositRecordAmount := records[0].Amount
	s.Require().Equal(expectedDepositRecordAmount, actualDepositRecordAmount, "deposit record amount")
}

func (s *KeeperTestSuite) TestLiquidStake_DifferentRedemptionRates() {
	tc := s.SetupLiquidStake()
	user := tc.user
	msg := tc.validMsg

	// Loop over exchange rates: {0.92, 0.94, ..., 1.2}
	for i := -8; i <= 10; i += 2 {
		redemptionDelta := sdk.NewDecWithPrec(1.0, 1).Quo(sdk.NewDec(10)).Mul(sdk.NewDec(int64(i))) // i = 2 => delta = 0.02
		newRedemptionRate := sdk.NewDec(1.0).Add(redemptionDelta)
		redemptionRateFloat := newRedemptionRate

		// Update rate in host zone
		hz := tc.initialState.hostZone
		hz.RedemptionRate = newRedemptionRate
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

		// Liquid stake for each balance and confirm stAtom minted
		startingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom).Amount
		_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
		s.Require().NoError(err)
		endingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom).Amount
		actualStAtomMinted := endingStAtomBalance.Sub(startingStAtomBalance)

		expectedStAtomMinted := sdk.NewDecFromInt(msg.Amount).Quo(redemptionRateFloat).TruncateInt()
		testDescription := fmt.Sprintf("st atom balance for redemption rate: %v", redemptionRateFloat)
		s.Require().Equal(expectedStAtomMinted, actualStAtomMinted, testDescription)
	}
}

func (s *KeeperTestSuite) TestLiquidStake_RateBelowMinThreshold() {
	tc := s.SetupLiquidStake()
	msg := tc.validMsg

	// Update rate in host zone to below min threshold
	hz := tc.initialState.hostZone
	hz.RedemptionRate = sdk.NewDec(8).Quo(sdk.NewDec(10))
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestLiquidStake_HostZoneNotFound() {
	tc := s.SetupLiquidStake()
	// Update message with invalid denom
	invalidMsg := tc.validMsg
	invalidMsg.HostDenom = "ufakedenom"
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "no host zone found for denom (ufakedenom): host zone not registered")
}

func (s *KeeperTestSuite) TestLiquidStake_IbcCoinParseError() {
	tc := s.SetupLiquidStake()
	// Update hostzone with denom that can't be parsed
	badHostZone := tc.initialState.hostZone
	badHostZone.IbcDenom = "ibc.0atom"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	badCoin := fmt.Sprintf("%v%s", tc.validMsg.Amount, badHostZone.IbcDenom)
	s.Require().EqualError(err, fmt.Sprintf("failed to parse coin (%s): invalid decimal coin expression: %s", badCoin, badCoin))
}

func (s *KeeperTestSuite) TestLiquidStake_NotIbcDenom() {
	tc := s.SetupLiquidStake()
	// Update hostzone with non-ibc denom
	badDenom := "i/uatom"
	badHostZone := tc.initialState.hostZone
	badHostZone.IbcDenom = badDenom
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)
	// Fund the user with the non-ibc denom
	s.FundAccount(tc.user.acc, sdk.NewInt64Coin(badDenom, 1000000000))
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("denom is not an IBC token (%s): invalid token denom", badHostZone.IbcDenom))
}

func (s *KeeperTestSuite) TestLiquidStake_InsufficientBalance() {
	tc := s.SetupLiquidStake()
	// Set liquid stake amount to value greater than account balance
	invalidMsg := tc.validMsg
	balance := tc.user.atomBalance.Amount
	invalidMsg.Amount = balance.Add(sdk.NewInt(1000))
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	expectedErr := fmt.Sprintf("balance is lower than staking amount. staking amount: %v, balance: %v: insufficient funds", balance.Add(sdk.NewInt(1000)), balance)
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestLiquidStake_NoEpochTracker() {
	tc := s.SetupLiquidStake()
	// Remove epoch tracker
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("no epoch number for epoch (%s): not found", epochtypes.STRIDE_EPOCH))
}

func (s *KeeperTestSuite) TestLiquidStake_NoDepositRecord() {
	tc := s.SetupLiquidStake()
	// Remove deposit record
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx, 1)
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("no deposit record for epoch (%d): not found", 1))
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidHostAddress() {
	tc := s.SetupLiquidStake()

	// Update hostzone with invalid address
	badHostZone := tc.initialState.hostZone
	badHostZone.Address = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "could not bech32 decode address cosmosXXX of zone with id: GAIA")
}
