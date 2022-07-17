package keeper_test

import (
	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	// "github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"
)

// type Account struct {
// 	acc           sdk.AccAddress
// 	atomBalance   sdk.Coin
// 	stAtomBalance sdk.Coin
// }

// type State struct {
// 	depositRecordAmount int64
// 	hostZone            types.HostZone
// }

type RedeemStakeTestCase struct {
	user         Account
	module       Account
	initialState State
	validMsg     stakeibc.MsgRedeemStake
}

func (suite *KeeperTestSuite) SetupRedeemStake() RedeemStakeTestCase {
	redeemAmount := int64(1_000_000)
	initialDepositAmount := int64(1_000_000)
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
		IBCDenom:       "ibc/uatom",
		Bech32Prefix:   "cosmos",
		RedemptionRate: sdk.NewDec(1.0),
		StakedBal:      1234567890,
	}

	epochTrackerDay := stakeibc.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     1,
	}

	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             initialDepositAmount,
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
	suite.App.RecordsKeeper.SetDepositRecord(suite.Ctx, initialDepositRecord)
	suite.App.RecordsKeeper.SetEpochUnbondingRecord(suite.Ctx, epochUnbondingRecord)
	// TODO  why do we need to set this to 2 instead of 1? revisit
	suite.App.RecordsKeeper.SetEpochUnbondingRecordCount(suite.Ctx, 2)

	return RedeemStakeTestCase{
		user:   user,
		module: module,
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
	redeemAmount := sdk.NewInt(msg.Amount)

	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// User STUATOM balance should have DECREASED by the amount to be redeemed
	expectedUserStAtomBalance := user.stAtomBalance.SubAmount(redeemAmount)
	actualUserStAtomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, user.acc, "stuatom")
	suite.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")

	// Gaia's hostZoneUnbonding amount should have INCREASED from 0 to be amount redeemed multiplied by the redemption rate
	// TODO how can we check the INCREASE rather than the absolute amount here?
	epochUnbondingRecord, found := suite.App.RecordsKeeper.GetLatestEpochUnbondingRecord(suite.Ctx)
	suite.Require().True(found)
	hostZoneUnbondingGaia, found := epochUnbondingRecord.HostZoneUnbondings["GAIA"]
	suite.Require().True(found)
	actualHostZoneUnbondingGaiaAmount := int64(hostZoneUnbondingGaia.Amount)
	hostZone, _ := suite.App.StakeibcKeeper.GetHostZone(suite.Ctx, msg.HostZone)
	expectedHostZoneUnbondingGaiaAmount := redeemAmount.Int64() * hostZone.RedemptionRate.TruncateInt().Int64()
	suite.Require().Equal(expectedHostZoneUnbondingGaiaAmount, actualHostZoneUnbondingGaiaAmount, "host zone unbonding amount")

	suite.Require().Equal(1, 2)
}

func (suite *KeeperTestSuite) TestRedeemStakeHostZoneNotFound() {
	tc := suite.SetupRedeemStake()
	// Update message with invalid denom
	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, "host zone is invalid: fake_host_zone: host zone not registered")
}
