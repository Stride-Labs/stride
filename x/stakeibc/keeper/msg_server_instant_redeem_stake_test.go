package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v7/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v7/x/stakeibc/types"
)

type InstantRedeemStakeState struct {
	initialAmount sdkmath.Int
	hostZone      stakeibctypes.HostZone
}

type InstantRedeemStakeTestCase struct {
	user         Account
	zoneAccount  Account
	initialState InstantRedeemStakeState
	validMsg     stakeibctypes.MsgInstantRedeemStake
}

func (s *KeeperTestSuite) SetupInstantRedeemStake(unbondAmount int64, depositAmounts []int64, rewardAmount int64) InstantRedeemStakeTestCase {
	unbondAmountInt := sdkmath.NewInt(unbondAmount)
	initialAmount := sdkmath.NewInt(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(IbcAtom, initialAmount.Int64()),
		stAtomBalance: sdk.NewInt64Coin(StAtom, unbondAmount),
	}
	s.FundAccount(user.acc, user.atomBalance)
	s.FundAccount(user.acc, user.stAtomBalance)

	zoneAddress := stakeibctypes.NewZoneAddress(HostChainId)

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
		StakedBal:      unbondAmountInt,
	}

	for i := 0; i < len(depositAmounts); i++ {
		depositRecord := recordtypes.DepositRecord{
			Id:                 uint64(i),
			DepositEpochNumber: uint64(i),
			HostZoneId:         HostChainId,
			Amount:             sdkmath.NewInt(depositAmounts[i]),
			Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
		}
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)
	}

	reinvestmentDepositId := len(depositAmounts)
	depositRecord := recordtypes.DepositRecord{
		Id:                 uint64(reinvestmentDepositId),
		DepositEpochNumber: uint64(reinvestmentDepositId),
		HostZoneId:         HostChainId,
		Amount:             sdkmath.NewInt(rewardAmount),
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return InstantRedeemStakeTestCase{
		user:        user,
		zoneAccount: zoneAccount,
		initialState: InstantRedeemStakeState{
			initialAmount: initialAmount,
			hostZone:      hostZone,
		},
		validMsg: stakeibctypes.MsgInstantRedeemStake{
			Creator:  user.acc.String(),
			HostZone: HostChainId,
			Amount:   unbondAmountInt,
		},
	}
}

func (s *KeeperTestSuite) getPendingDepositTotal(chainId string) sdkmath.Int {
	depositRecords := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	pendingDepositRecords := s.App.RecordsKeeper.FilterDepositRecords(depositRecords, func(record recordtypes.DepositRecord) (condition bool) {
		return record.Status == recordtypes.DepositRecord_TRANSFER_QUEUE && record.HostZoneId == chainId
	})
	return s.App.RecordsKeeper.SumDepositRecords(pendingDepositRecords)
}

func (s *KeeperTestSuite) TestInstantRedeemStake_Successful() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{300_000, 500_000, 200_000}, 100_000)
	user := tc.user
	zoneAccount := tc.zoneAccount
	initialStAtomSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)
	msg := tc.validMsg
	hostZone := tc.initialState.hostZone
	initialPendingDeposits := s.getPendingDepositTotal(hostZone.ChainId)

	// Validate Instant Redeem Stake
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

	// User STUATOM balance should have DECREASED by the amount unbonded
	expectedUserStAtomBalance := user.stAtomBalance.SubAmount(msg.Amount)
	actualUserStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom)
	s.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")

	// User IBC/UATOM balance should have INCREASED by the amount unbonded
	expectedUserAtomBalance := user.atomBalance.AddAmount(msg.Amount)
	actualUserAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, IbcAtom)
	s.CompareCoins(expectedUserAtomBalance, actualUserAtomBalance, "user ibc/uatom balance")

	// ZoneAccount IBC/UATOM balance should have DECREASED by the size of the stake
	expectedZoneAccountAtomBalance := zoneAccount.atomBalance.SubAmount(msg.Amount)
	actualZoneAccountAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount.acc, IbcAtom)
	s.CompareCoins(expectedZoneAccountAtomBalance, actualZoneAccountAtomBalance, "zoneAccount ibc/uatom balance")

	// Bank supply of STUATOM should have DECREASED by the size of the stake
	expectedBankSupply := initialStAtomSupply.SubAmount(msg.Amount)
	actualBankSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)
	s.CompareCoins(expectedBankSupply, actualBankSupply, "bank stuatom supply")

	// Total pending Deposit Record amounts should have DECREASED by the size of the stake
	expectedPendingDeposits := initialPendingDeposits.Sub(msg.Amount)
	actualPendingDeposits := s.getPendingDepositTotal(hostZone.ChainId)
	s.Require().Equal(expectedPendingDeposits.Int64(), actualPendingDeposits.Int64(), fmt.Sprintf("unexpected pending deposits of %v, expected %v", actualPendingDeposits, expectedPendingDeposits))

	// Validate Instant Redeem Stake, subsequent call will not have funds.
	_, err = s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().EqualError(err, fmt.Sprintf("balance is lower than redemption amount. redemption amount: %v, balance %v: : invalid coins", msg.Amount, actualUserStAtomBalance.Amount))
}

func (s *KeeperTestSuite) TestInstantRedeemStake_RedeemJustMoreThanStaked() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{500_000, 500_000}, 0)

	invalidMsg := tc.validMsg
	invalidMsg.Amount = sdkmath.NewInt(1_000_001)
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("balance is lower than redemption amount. redemption amount: %v, balance %v: : invalid coins", invalidMsg.Amount, tc.initialState.initialAmount))
}

func (s *KeeperTestSuite) TestInstantRedeemStake_DifferentRedemptionRates() {
	tc := s.SetupInstantRedeemStake(1_000_000, []int64{100_000, 100_000}, 0)
	user := tc.user
	msg := tc.validMsg
	msg.Amount = sdkmath.NewInt(1_000) // Do a bunch of smaller amounts so we don't run out

	// Loop over exchange rates: {0.92, 0.94, ..., 1.2}
	for i := -8; i <= 10; i += 2 {
		redemptionDelta := sdk.NewDecWithPrec(1.0, 1).Quo(sdk.NewDec(10)).Mul(sdk.NewDec(int64(i))) // i = 2 => delta = 0.02
		newRedemptionRate := sdk.NewDec(1.0).Add(redemptionDelta)
		redemptionRateFloat := newRedemptionRate

		// Update rate in host zone
		hz := tc.initialState.hostZone
		hz.RedemptionRate = newRedemptionRate
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

		// Instant redeem stake for each balance and confirm Atom redeemed
		startingAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, IbcAtom).Amount
		_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &msg)
		s.Require().NoError(err)
		endingAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, IbcAtom).Amount
		actualAtomRedeemed := endingAtomBalance.Sub(startingAtomBalance)

		expectedAtomRedeemed := sdk.NewDecFromInt(msg.Amount).Mul(redemptionRateFloat).TruncateInt()
		testDescription := fmt.Sprintf("atom balance for redemption rate: %v, expectedAtomRedeemed = %v, actualAtomRedeemed = %v", redemptionRateFloat, expectedAtomRedeemed, actualAtomRedeemed)
		s.Require().Equal(expectedAtomRedeemed, actualAtomRedeemed, testDescription)
	}
}

// It shouldn't be true ever that there are no deposit records, but just to be sure, and catch a different out of funds case
func (s *KeeperTestSuite) TestInstantRedeemStake_RedeemWithNoDepositRecords() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{}, 0)

	invalidMsg := tc.validMsg
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("cannot remove an amount %v g.t. pending deposit balance on host zone: %v: invalid amount", invalidMsg.Amount, 0))
}

func (s *KeeperTestSuite) TestInstantRedeemStake_InvalidCreatorAddress() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{500_000, 500_000}, 0)
	invalidMsg := tc.validMsg

	// cosmos instead of stride address
	invalidMsg.Creator = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: invalid Bech32 prefix; expected stride, got cosmos: invalid address", invalidMsg.Creator))

	// invalid stride address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err = s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid checksum (expected 8dpmg9 got yxp8uf): invalid address", invalidMsg.Creator))

	// empty address
	invalidMsg.Creator = ""
	_, err = s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: empty address string is not allowed: invalid address", invalidMsg.Creator))

	// wrong len address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufabc"
	_, err = s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid character not part of charset: 98: invalid address", invalidMsg.Creator))
}

func (s *KeeperTestSuite) TestInstantRedeemStake_HostZoneNotFound() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{500_000, 500_000}, 0)

	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "host zone is invalid: fake_host_zone: host zone not registered")
}

func (s *KeeperTestSuite) TestInstantRedeemStake_RateAboveMaxThreshold() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{500_000, 500_000}, 0)

	hz := tc.initialState.hostZone
	hz.RedemptionRate = sdk.NewDec(100)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestInstantRedeemStake_RedeemMoreThanStaked() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{500_000, 500_000}, 0)

	invalidMsg := tc.validMsg
	invalidMsg.Amount = sdkmath.NewInt(1_000_000_000_000_000)
	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("balance is lower than redemption amount. redemption amount: %v, balance %v: : invalid coins", invalidMsg.Amount, tc.initialState.initialAmount))
}

func (s *KeeperTestSuite) TestInstantRedeemStake_InvalidHostAddress() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{500_000, 500_000}, 0)

	// Update hostzone with invalid address
	badHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	badHostZone.Address = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "could not bech32 decode address cosmosXXX of zone with id: GAIA")
}

func (s *KeeperTestSuite) TestInstantRedeemStake_HaltedZone() {
	tc := s.SetupInstantRedeemStake(int64(1_000_000), []int64{500_000, 500_000}, 0)

	// Update hostzone with halted
	haltedHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	haltedHostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haltedHostZone)

	_, err := s.GetMsgServer().InstantRedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "halted host zone found for zone (GAIA): Halted host zone found")
}
