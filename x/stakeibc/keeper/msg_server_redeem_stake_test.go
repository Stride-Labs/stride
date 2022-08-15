package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type RedeemStakeState struct {
	epochNumber                        uint64
	initialNativeEpochUnbondingAmount  uint64
	initialStTokenEpochUnbondingAmount uint64
}
type RedeemStakeTestCase struct {
	user         Account
	module       Account
	initialState RedeemStakeState
	validMsg     stakeibc.MsgRedeemStake
}

func (suite *KeeperTestSuite) SetupRedeemStake() RedeemStakeTestCase {
	redeemAmount := uint64(1_000_000)
	user := Account{
		acc:           suite.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	suite.FundAccount(user.acc, user.atomBalance)
	suite.FundAccount(user.acc, user.stAtomBalance)

	module := Account{
		acc:           suite.App.AccountKeeper.GetModuleAddress(stakeibc.ModuleName),
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	suite.FundModuleAccount(stakeibc.ModuleName, module.atomBalance)
	suite.FundModuleAccount(stakeibc.ModuleName, module.stAtomBalance)

	// TODO define the host zone with stakedBal and validators with staked amounts
	hostZone := stakeibc.HostZone{
		ChainId:        "GAIA",
		HostDenom:      "uatom",
		Bech32Prefix:   "cosmos",
		RedemptionRate: sdk.NewDec(1.0),
		StakedBal:      1234567890,
	}

	epochTrackerDay := stakeibc.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     1,
	}

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}

	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: uint64(0),
		Denom:             "uatom",
		HostZoneId:        "GAIA",
		Status:            recordtypes.HostZoneUnbonding_BONDED,
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)

	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, hostZone)
	suite.App.StakeibcKeeper.SetEpochTracker(suite.Ctx, epochTrackerDay)
	suite.App.RecordsKeeper.SetEpochUnbondingRecord(suite.Ctx, epochUnbondingRecord)

	return RedeemStakeTestCase{
		user:   user,
		module: module,
		initialState: RedeemStakeState{
			epochNumber:                        epochTrackerDay.EpochNumber,
			initialNativeEpochUnbondingAmount:  uint64(0),
			initialStTokenEpochUnbondingAmount: uint64(0),
		},
		validMsg: stakeibc.MsgRedeemStake{
			Creator:  user.acc.String(),
			Amount:   redeemAmount,
			HostZone: "GAIA",
			// TODO set this dynamically through test helpers for host zone
			Receiver: "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf",
		},
	}
}

func (suite *KeeperTestSuite) TestRedeemStakeSuccessful() {
	tc := suite.SetupRedeemStake()
	initialState := tc.initialState

	msg := tc.validMsg
	user := tc.user
	amt, err := cast.ToInt64E(msg.Amount)
	if err != nil {
		panic(err)
	}
	redeemAmount := sdk.NewInt(amt)

	// get the initial unbonding amount *before* calling liquid stake, so we can use it to calc expected vs actual in diff space
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// User STUATOM balance should have DECREASED by the amount to be redeemed
	expectedUserStAtomBalance := user.stAtomBalance.SubAmount(redeemAmount)
	actualUserStAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, "stuatom")
	suite.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")

	// Gaia's hostZoneUnbonding NATIVE TOKEN amount should have INCREASED from 0 to the amount redeemed multiplied by the redemption rate
	// Gaia's hostZoneUnbonding STTOKEN amount should have INCREASED from 0 to be amount redeemed
	epochTracker, found := suite.App.StakeibcKeeper.GetEpochTracker(suite.Ctx, "day")
	suite.Require().True(found)
	epochUnbondingRecord, found := suite.App.RecordsKeeper.GetEpochUnbondingRecord(suite.Ctx, epochTracker.EpochNumber)
	suite.Require().True(found)
	hostZoneUnbonding, found := suite.App.RecordsKeeper.GetHostZoneUnbondingByChainId(suite.Ctx, epochUnbondingRecord.EpochNumber, "GAIA")
	suite.Require().True(found)

	hostZone, _ := suite.App.StakeibcKeeper.GetHostZone(suite.Ctx, msg.HostZone)
	nativeRedemptionAmount := (redeemAmount.Int64() * hostZone.RedemptionRate.TruncateInt().Int64())
	stTokenBurnAmount := redeemAmount.Int64()

	actualHostZoneUnbondingNativeAmount := int64(hostZoneUnbonding.NativeTokenAmount)
	actualHostZoneUnbondingStTokenAmount := int64(hostZoneUnbonding.StTokenAmount)
	expectedHostZoneUnbondingNativeAmount := int64(initialState.initialNativeEpochUnbondingAmount) + nativeRedemptionAmount
	expectedHostZoneUnbondingStTokenAmount := int64(initialState.initialStTokenEpochUnbondingAmount) + stTokenBurnAmount

	suite.Require().Equal(expectedHostZoneUnbondingNativeAmount, actualHostZoneUnbondingNativeAmount, "host zone native unbonding amount")
	suite.Require().Equal(expectedHostZoneUnbondingStTokenAmount, actualHostZoneUnbondingStTokenAmount, "host zone stToken burn amount")

	// UserRedemptionRecord should have been created with correct amount, sender, receiver, host zone, isClaimable
	userRedemptionRecords := hostZoneUnbonding.UserRedemptionRecords
	suite.Require().Equal(len(userRedemptionRecords), 1)
	userRedemptionRecordId := userRedemptionRecords[0]
	userRedemptionRecord, found := suite.App.RecordsKeeper.GetUserRedemptionRecord(suite.Ctx, userRedemptionRecordId)
	suite.Require().True(found)
	// check amount
	suite.Require().Equal(int64(userRedemptionRecord.Amount), expectedHostZoneUnbondingNativeAmount)
	// check sender
	suite.Require().Equal(userRedemptionRecord.Sender, msg.Creator)
	// check receiver
	suite.Require().Equal(userRedemptionRecord.Receiver, msg.Receiver)
	// check host zone
	suite.Require().Equal(userRedemptionRecord.HostZoneId, msg.HostZone)
	// check isClaimable
	suite.Require().Equal(userRedemptionRecord.IsClaimable, false)
}

func (suite *KeeperTestSuite) TestInvalidCreatorAddress() {
	tc := suite.SetupRedeemStake()
	invalidMsg := tc.validMsg

	// cosmos instead of stride address
	invalidMsg.Creator = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: invalid Bech32 prefix; expected stride, got cosmos: invalid address", invalidMsg.Creator))

	// invalid stride address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid checksum (expected 8dpmg9 got yxp8uf): invalid address", invalidMsg.Creator))

	// empty address
	invalidMsg.Creator = ""
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: empty address string is not allowed: invalid address", invalidMsg.Creator))

	// wrong len address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufabc"
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid character not part of charset: 98: invalid address", invalidMsg.Creator))
}

func (suite *KeeperTestSuite) TestRedeemStakeHostZoneNotFound() {
	tc := suite.SetupRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, "host zone is invalid: fake_host_zone: host zone not registered")
}

func (suite *KeeperTestSuite) TestInvalidReceiverAddress() {
	tc := suite.SetupRedeemStake()

	invalidMsg := tc.validMsg

	// stride instead of cosmos address
	invalidMsg.Receiver = "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, "invalid receiver address (invalid Bech32 prefix; expected cosmos, got stride): invalid address")

	// invalid cosmos address
	invalidMsg.Receiver = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ua"
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, "invalid receiver address (decoding bech32 failed: invalid checksum (expected yxp8uf got yxp8ua)): invalid address")

	// empty address
	invalidMsg.Receiver = ""
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, "invalid receiver address (empty address string is not allowed): invalid address")

	// wrong len address
	invalidMsg.Receiver = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufa"
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, "invalid receiver address (decoding bech32 failed: invalid checksum (expected xp8ugp got xp8ufa)): invalid address")
}

func (suite *KeeperTestSuite) TestRedeemStakeRedeemMoreThanStaked() {
	tc := suite.SetupRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.Amount = uint64(1_000_000_000_000_000)
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, fmt.Sprintf("cannot unstake an amount g.t. staked balance on host zone: %d: invalid amount", invalidMsg.Amount))
}

func (suite *KeeperTestSuite) TestRedeemStakeNoEpochTrackerDay() {
	tc := suite.SetupRedeemStake()

	invalidMsg := tc.validMsg
	suite.App.RecordsKeeper.RemoveEpochUnbondingRecord(suite.Ctx, tc.initialState.epochNumber)
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, "latest epoch unbonding record not found: epoch unbonding record not found")
}

func (suite *KeeperTestSuite) TestRedeemStakeUserAlreadyRedeemedThisEpoch() {
	tc := suite.SetupRedeemStake()

	invalidMsg := tc.validMsg
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().NoError(err)
	_, err = suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)
	suite.Require().EqualError(err, fmt.Sprintf("user already redeemed this epoch: GAIA.1.%s: redemption record already exists", suite.TestAccs[0]))
}

func (suite *KeeperTestSuite) TestRedeemStakeHostZoneNoUnbondings() {
	tc := suite.SetupRedeemStake()

	invalidMsg := tc.validMsg
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: uint64(0),
		Denom:             "uatom",
		HostZoneId:		"NOT_GAIA",
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)
	
	suite.App.RecordsKeeper.SetEpochUnbondingRecord(suite.Ctx, epochUnbondingRecord)
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, "host zone not found in unbondings: GAIA: host zone not registered")
}
