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

type RedeemStakeTestCase struct {
	user                        Account
	module                      Account
	initialState                State
	initialEpochUnbondingAmount uint64
	validMsg                    stakeibc.MsgRedeemStake
}

func (suite *KeeperTestSuite) SetupRedeemStake() RedeemStakeTestCase {
	redeemAmount := uint64(1_000_000)
	initialDepositAmount := uint64(1_000_000)
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
		Id:                   1,
		UnbondingEpochNumber: 1,
		HostZoneUnbondings:   make(map[string]*recordtypes.HostZoneUnbonding),
	}

	epochUnbondingRecord.HostZoneUnbondings["GAIA"] = &recordtypes.HostZoneUnbonding{
		Amount:     uint64(0),
		Denom:      "uatom",
		HostZoneId: "GAIA",
		Status:     recordtypes.HostZoneUnbonding_BONDED,
	}

	suite.App.StakeibcKeeper.SetHostZone(suite.Ctx, hostZone)
	suite.App.StakeibcKeeper.SetEpochTracker(suite.Ctx, epochTrackerDay)
	suite.App.RecordsKeeper.AppendEpochUnbondingRecord(suite.Ctx, epochUnbondingRecord)

	return RedeemStakeTestCase{
		user:                        user,
		module:                      module,
		initialEpochUnbondingAmount: uint64(0),
		initialState: State{
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
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

	msg := tc.validMsg
	user := tc.user
	redeemAmount := sdk.NewInt(cast.ToInt64(msg.Amount))

	// get the initial unbonding amount *before* calling liquid stake, so we can use it to calc expected vs actual in diff space
	actualHostZoneUnbondingGaiaAmountStart := int64(tc.initialEpochUnbondingAmount)

	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// User STUATOM balance should have DECREASED by the amount to be redeemed
	expectedUserStAtomBalance := user.stAtomBalance.SubAmount(redeemAmount)
	actualUserStAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, "stuatom")
	suite.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")

	// Gaia's hostZoneUnbonding amount should have INCREASED from 0 to be amount redeemed multiplied by the redemption rate
	epochUnbondingRecord, found := suite.App.RecordsKeeper.GetLatestEpochUnbondingRecord(suite.Ctx)
	suite.Require().True(found)
	hostZoneUnbondingGaia, found := epochUnbondingRecord.HostZoneUnbondings["GAIA"]
	suite.Require().True(found)
	actualHostZoneUnbondingGaiaAmount := int64(hostZoneUnbondingGaia.Amount)
	hostZone, _ := suite.App.StakeibcKeeper.GetHostZone(suite.Ctx, msg.HostZone)
	expectedHostZoneUnbondingGaiaAmount := redeemAmount.Int64() * hostZone.RedemptionRate.TruncateInt().Int64()
	suite.Require().Equal(expectedHostZoneUnbondingGaiaAmount-actualHostZoneUnbondingGaiaAmountStart, actualHostZoneUnbondingGaiaAmount-actualHostZoneUnbondingGaiaAmountStart, "host zone unbonding amount")

	// UserRedemptionRecord should have been created with correct amount, sender, receiver, host zone, isClaimable
	userRedemptionRecords := hostZoneUnbondingGaia.UserRedemptionRecords
	suite.Require().Equal(len(userRedemptionRecords), 1)
	userRedemptionRecordId := userRedemptionRecords[0]
	userRedemptionRecord, found := suite.App.RecordsKeeper.GetUserRedemptionRecord(suite.Ctx, userRedemptionRecordId)
	suite.Require().True(found)
	// check amount
	suite.Require().Equal(int64(userRedemptionRecord.Amount), expectedHostZoneUnbondingGaiaAmount)
	// check sender
	suite.Require().Equal(userRedemptionRecord.Sender, msg.Creator)
	// check receiver
	suite.Require().Equal(userRedemptionRecord.Receiver, msg.Receiver)
	// check host zone
	suite.Require().Equal(userRedemptionRecord.HostZoneId, msg.HostZone)
	// check isClaimable
	suite.Require().Equal(userRedemptionRecord.IsClaimable, false)

	// make sure stTokens that were transfered to the module account were burned (stAsset supply should decrease by redeemAmount)
	expectedStAssetSupply := tc.module.stAtomBalance.Amount.Int64() + tc.user.stAtomBalance.Amount.Int64() - redeemAmount.Int64()
	actualStAssetSupply := suite.App.BankKeeper.GetSupply(suite.Ctx, "stuatom")
	suite.Require().Equal(expectedStAssetSupply, actualStAssetSupply.Amount.Int64())
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
	suite.App.RecordsKeeper.SetEpochUnbondingRecord(suite.Ctx, recordtypes.EpochUnbondingRecord{})
	suite.App.RecordsKeeper.SetEpochUnbondingRecordCount(suite.Ctx, 0)
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
		Id:                 1,
		HostZoneUnbondings: make(map[string]*recordtypes.HostZoneUnbonding),
	}
	epochUnbondingRecord.HostZoneUnbondings["NOT_GAIA"] = &recordtypes.HostZoneUnbonding{
		Amount: uint64(0),
		Denom:  "uatom",
	}
	suite.App.RecordsKeeper.AppendEpochUnbondingRecord(suite.Ctx, epochUnbondingRecord)
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, "host zone not found in unbondings: GAIA: host zone not registered")
}
