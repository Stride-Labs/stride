package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type RedeemStakeState struct {
	epochNumber                        uint64
	initialNativeEpochUnbondingAmount  sdk.Int
	initialStTokenEpochUnbondingAmount sdk.Int
}
type RedeemStakeTestCase struct {
	user         Account
	hostZone     stakeibctypes.HostZone
	zoneAccount  Account
	initialState RedeemStakeState
	validMsg     stakeibctypes.MsgRedeemStake
}

func (s *KeeperTestSuite) SetupRedeemStake() RedeemStakeTestCase {
	redeemAmount := sdk.NewInt(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	s.FundAccount(user.acc, user.atomBalance)
	s.FundAccount(user.acc, user.stAtomBalance)

	zoneAddress := stakeibctypes.NewZoneAddress(HostChainId)

	zoneAccount := Account{
		acc:           zoneAddress,
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	s.FundAccount(zoneAccount.acc, zoneAccount.atomBalance)
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	// TODO define the host zone with stakedBal and validators with staked amounts
	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      "uatom",
		Bech32Prefix:   "cosmos",
		RedemptionRate: sdk.NewDec(1.0),
		StakedBal:      sdk.NewInt(1234567890),
		Address:        zoneAddress.String(),
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
		NativeTokenAmount: sdk.ZeroInt(),
		Denom:             "uatom",
		HostZoneId:        HostChainId,
		Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTrackerDay)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	return RedeemStakeTestCase{
		user:        user,
		hostZone:    hostZone,
		zoneAccount: zoneAccount,
		initialState: RedeemStakeState{
			epochNumber:                        epochTrackerDay.EpochNumber,
			initialNativeEpochUnbondingAmount:  sdk.ZeroInt(),
			initialStTokenEpochUnbondingAmount: sdk.ZeroInt(),
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

	// get the initial unbonding amount *before* calling liquid stake, so we can use it to calc expected vs actual in diff space
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

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
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, msg.HostZone)
	s.Require().True(found, "host zone")

	nativeRedemptionAmount := redeemAmount.Mul(hostZone.RedemptionRate.TruncateInt())
	stTokenBurnAmount := redeemAmount

	expectedHostZoneUnbondingNativeAmount := initialState.initialNativeEpochUnbondingAmount.Add(nativeRedemptionAmount)
	expectedHostZoneUnbondingStTokenAmount := initialState.initialStTokenEpochUnbondingAmount.Add(stTokenBurnAmount)

	s.Require().Equal(expectedHostZoneUnbondingNativeAmount, hostZoneUnbonding.NativeTokenAmount, "host zone native unbonding amount")
	s.Require().Equal(expectedHostZoneUnbondingStTokenAmount, hostZoneUnbonding.StTokenAmount, "host zone stToken burn amount")

	// UserRedemptionRecord should have been created with correct amount, sender, receiver, host zone, claimIsPending
	userRedemptionRecords := hostZoneUnbonding.UserRedemptionRecords
	s.Require().Equal(len(userRedemptionRecords), 1)
	userRedemptionRecordId := userRedemptionRecords[0]
	userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, userRedemptionRecordId)
	s.Require().True(found)
	// check amount
	s.Require().Equal(expectedHostZoneUnbondingNativeAmount, userRedemptionRecord.Amount, "redemption record amount")
	// check sender
	s.Require().Equal(msg.Creator, userRedemptionRecord.Sender, "redemption record sender")
	// check receiver
	s.Require().Equal(msg.Receiver, userRedemptionRecord.Receiver, "redemption record receiver")
	// check host zone
	s.Require().Equal(msg.HostZone, userRedemptionRecord.HostZoneId, "redemption record host zone")
	// check claimIsPending
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

	s.Require().EqualError(err, "host zone is invalid: fake_host_zone: host zone not registered")
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
	invalidMsg.Amount = sdk.NewInt(1_000_000_000_000_000)
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

func (s *KeeperTestSuite) TestRedeemStake_UserAlreadyRedeemedThisEpoch() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().NoError(err)
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("user already redeemed this epoch: GAIA.1.%s: redemption record already exists", s.TestAccs[0]))
}

func (s *KeeperTestSuite) TestRedeemStake_HostZoneNoUnbondings() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: sdk.ZeroInt(),
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
	badHostZone.Address = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "could not bech32 decode address cosmosXXX of zone with id: GAIA")
}
