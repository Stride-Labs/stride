package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/Stride-Labs/stride/v23/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v23/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v23/x/stakeibc/types"
)

type RedeemStakeState struct {
	epochNumber                        uint64
	initialNativeEpochUnbondingAmount  sdkmath.Int
	initialStTokenEpochUnbondingAmount sdkmath.Int
}
type RedeemStakeTestCase struct {
	user                 Account
	hostZone             stakeibctypes.HostZone
	zoneAccount          Account
	initialState         RedeemStakeState
	validMsg             stakeibctypes.MsgRedeemStake
	expectedNativeAmount sdkmath.Int
}

func (s *KeeperTestSuite) SetupRedeemStake() RedeemStakeTestCase {
	redeemAmount := sdkmath.NewInt(1_000_000)
	redemptionRate := sdk.MustNewDecFromStr("1.5")
	expectedNativeAmount := sdkmath.NewInt(1_500_000)

	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	s.FundAccount(user.acc, user.atomBalance)
	s.FundAccount(user.acc, user.stAtomBalance)

	depositAddress := stakeibctypes.NewHostZoneDepositAddress(HostChainId)

	zoneAccount := Account{
		acc:           depositAddress,
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	s.FundAccount(zoneAccount.acc, zoneAccount.atomBalance)
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	// TODO define the host zone with total delegation and validators with staked amounts
	hostZone := stakeibctypes.HostZone{
		ChainId:            HostChainId,
		HostDenom:          "uatom",
		Bech32Prefix:       "cosmos",
		RedemptionRate:     redemptionRate,
		TotalDelegations:   sdkmath.NewInt(1234567890),
		DepositAddress:     depositAddress.String(),
		RedemptionsEnabled: true,
	}

	epochTrackerDay := stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     1,
	}

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}

	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: sdkmath.ZeroInt(),
		Denom:             "uatom",
		HostZoneId:        HostChainId,
		Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTrackerDay)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	return RedeemStakeTestCase{
		user:                 user,
		hostZone:             hostZone,
		zoneAccount:          zoneAccount,
		expectedNativeAmount: expectedNativeAmount,
		initialState: RedeemStakeState{
			epochNumber:                        epochTrackerDay.EpochNumber,
			initialNativeEpochUnbondingAmount:  sdkmath.ZeroInt(),
			initialStTokenEpochUnbondingAmount: sdkmath.ZeroInt(),
		},
		validMsg: stakeibctypes.MsgRedeemStake{
			Creator:  user.acc.String(),
			Amount:   redeemAmount,
			HostZone: HostChainId,
			// TODO set this dynamically through test helpers for host zone
			Receiver: "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf",
		},
	}
}

func (s *KeeperTestSuite) TestRedeemStake_Successful() {
	tc := s.SetupRedeemStake()
	initialState := tc.initialState

	msg := tc.validMsg
	user := tc.user
	redeemAmount := msg.Amount

	// Split the message amount in 2, and call redeem stake twice (each with half the amount)
	// This will check that the same user can redeem multiple times
	msg1 := msg
	msg1.Amount = msg1.Amount.Quo(sdkmath.NewInt(2)) // half the amount

	msg2 := msg
	msg2.Amount = msg.Amount.Sub(msg1.Amount) // remaining half

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &msg1)
	s.Require().NoError(err, "no error expected during first redemption")

	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &msg2)
	s.Require().NoError(err, "no error expected during second redemption")

	// User STUATOM balance should have DECREASED by the amount to be redeemed
	expectedUserStAtomBalance := user.stAtomBalance.SubAmount(redeemAmount)
	actualUserStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, "stuatom")
	s.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")

	// Gaia's hostZoneUnbonding NATIVE TOKEN amount should have INCREASED from 0 to the amount redeemed multiplied by the redemption rate
	// Gaia's hostZoneUnbonding STTOKEN amount should have INCREASED from 0 to be amount redeemed
	epochTracker, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx, "day")
	s.Require().True(found, "epoch tracker")
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, epochTracker.EpochNumber)
	s.Require().True(found, "epoch unbonding record")
	hostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochUnbondingRecord.EpochNumber, HostChainId)
	s.Require().True(found, "host zone unbondings by chain ID")

	expectedHostZoneUnbondingNativeAmount := initialState.initialNativeEpochUnbondingAmount.Add(tc.expectedNativeAmount)
	expectedHostZoneUnbondingStTokenAmount := initialState.initialStTokenEpochUnbondingAmount.Add(redeemAmount)

	s.Require().Equal(expectedHostZoneUnbondingNativeAmount, hostZoneUnbonding.NativeTokenAmount, "host zone native unbonding amount")
	s.Require().Equal(expectedHostZoneUnbondingStTokenAmount, hostZoneUnbonding.StTokenAmount, "host zone stToken burn amount")

	// UserRedemptionRecord should have been created with correct amount, sender, receiver, host zone, claimIsPending
	userRedemptionRecords := hostZoneUnbonding.UserRedemptionRecords
	s.Require().Equal(len(userRedemptionRecords), 1)
	userRedemptionRecordId := userRedemptionRecords[0]
	userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, userRedemptionRecordId)
	s.Require().True(found)

	s.Require().Equal(msg.Amount, userRedemptionRecord.StTokenAmount, "redemption record sttoken amount")
	s.Require().Equal(tc.expectedNativeAmount, userRedemptionRecord.NativeTokenAmount, "redemption record native amount")
	s.Require().Equal(msg.Receiver, userRedemptionRecord.Receiver, "redemption record receiver")
	s.Require().Equal(msg.HostZone, userRedemptionRecord.HostZoneId, "redemption record host zone")
	s.Require().False(userRedemptionRecord.ClaimIsPending, "redemption record is not claimable")
	s.Require().NotEqual(hostZoneUnbonding.Status, recordtypes.HostZoneUnbonding_CLAIMABLE, "host zone unbonding should NOT be marked as CLAIMABLE")
}

func (s *KeeperTestSuite) TestRedeemStake_InvalidCreatorAddress() {
	tc := s.SetupRedeemStake()
	invalidMsg := tc.validMsg

	// cosmos instead of stride address
	invalidMsg.Creator = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: invalid Bech32 prefix; expected stride, got cosmos: invalid address", invalidMsg.Creator))

	// invalid stride address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid checksum (expected 8dpmg9 got yxp8uf): invalid address", invalidMsg.Creator))

	// empty address
	invalidMsg.Creator = ""
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: empty address string is not allowed: invalid address", invalidMsg.Creator))

	// wrong len address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufabc"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid character not part of charset: 98: invalid address", invalidMsg.Creator))
}

func (s *KeeperTestSuite) TestRedeemStake_HostZoneNotFound() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "host zone fake_host_zone not found: host zone not found")
}

func (s *KeeperTestSuite) TestRedeemStake_RateAboveMaxThreshold() {
	tc := s.SetupRedeemStake()

	hz := tc.hostZone
	hz.RedemptionRate = sdk.NewDec(100)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestRedeemStake_InvalidReceiverAddress() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg

	// stride instead of cosmos address
	invalidMsg.Receiver = "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (invalid Bech32 prefix; expected cosmos, got stride): invalid address")

	// invalid cosmos address
	invalidMsg.Receiver = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ua"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (decoding bech32 failed: invalid checksum (expected yxp8uf got yxp8ua)): invalid address")

	// empty address
	invalidMsg.Receiver = ""
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (empty address string is not allowed): invalid address")

	// wrong len address
	invalidMsg.Receiver = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufa"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (decoding bech32 failed: invalid checksum (expected xp8ugp got xp8ufa)): invalid address")
}

func (s *KeeperTestSuite) TestRedeemStake_RedeemMoreThanStaked() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.Amount = sdkmath.NewInt(1_000_000_000_000_000)
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("cannot unstake an amount g.t. staked balance on host zone: %v: invalid amount", invalidMsg.Amount))
}

func (s *KeeperTestSuite) TestRedeemStake_NoEpochTrackerDay() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, tc.initialState.epochNumber)
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "latest epoch unbonding record not found: epoch unbonding record not found")
}

func (s *KeeperTestSuite) TestRedeemStake_HostZoneNoUnbondings() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: sdkmath.ZeroInt(),
		Denom:             "uatom",
		HostZoneId:        "NOT_GAIA",
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)

	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "host zone not found in unbondings: GAIA: host zone not registered")
}

func (s *KeeperTestSuite) TestRedeemStake_InvalidHostAddress() {
	tc := s.SetupRedeemStake()

	// Update hostzone with invalid address
	badHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	badHostZone.DepositAddress = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "could not bech32 decode address cosmosXXX of zone with id: GAIA")
}

func (s *KeeperTestSuite) TestRedeemStake_HaltedZone() {
	tc := s.SetupRedeemStake()

	// Update hostzone with halted
	haltedHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	haltedHostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haltedHostZone)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "host zone GAIA is halted: Halted host zone found")
}

func (s *KeeperTestSuite) TestRedeemStake_RedemptionsDisabled() {
	tc := s.SetupRedeemStake()

	// Update hostzone with halted
	haltedHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	haltedHostZone.RedemptionsEnabled = false
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haltedHostZone)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "redemptions disabled")
}
