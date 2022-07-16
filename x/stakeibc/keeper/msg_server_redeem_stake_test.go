package keeper_test

import (
	// "fmt"

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
	stakeAmount := int64(1_000_000)
	initialDepositAmount := int64(1_000_000)
	user := Account{
		acc:           suite.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 0),
	}
	suite.FundAccount(user.acc, user.atomBalance)

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

	return RedeemStakeTestCase{
		user:   user,
		module: module,
		initialState: State{
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
		},
		validMsg: stakeibc.MsgRedeemStake{
			Creator:   user.acc.String(),
			Amount:   stakeAmount,
			HostZone: "GAIA",
			// TODO set this dynamically through test helpers for host zone
			Receiver: "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf",
		},
	}
}

func (suite *KeeperTestSuite) TestRedeemStakeSuccessful() {
	// TODO uncomment these and use them properly in tests below
	tc := suite.SetupRedeemStake()
	// user := tc.user
	// module := tc.module
	msg := tc.validMsg
	// stakeAmount := sdk.NewInt(msg.Amount)

	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// TODO tests for the redeem stake logic

	suite.Require().Equal(1, 2)
}

func (suite *KeeperTestSuite) TestRedeemStakeHostZoneNotFound() {
	tc := suite.SetupRedeemStake()
	// Update message with invalid denom
	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := suite.msgServer.RedeemStake(sdk.WrapSDKContext(suite.Ctx), &invalidMsg)

	suite.Require().EqualError(err, "host zone not registered: host zone is invalid: fake_host_zone")
}
