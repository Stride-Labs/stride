package v10_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	v10 "github.com/Stride-Labs/stride/v9/app/upgrades/v10"

	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	recordskeeper "github.com/Stride-Labs/stride/v9/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	cosmosproto "github.com/cosmos/gogoproto/proto"
	deprecatedproto "github.com/golang/protobuf/proto" //nolint:staticcheck
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)

	s.ConfirmUpgradeSucceededs("v10", dummyUpgradeHeight)

	// Check mint parameters after upgrade
	proportions := s.App.MintKeeper.GetParams(s.Ctx).DistributionProportions

	s.Require().Equal(v10.StakingProportion,
		proportions.Staking.String()[:6], "staking")

	s.Require().Equal(v10.CommunityPoolGrowthProportion,
		proportions.CommunityPoolGrowth.String()[:6], "community pool growth")

	s.Require().Equal(v10.StrategicReserveProportion,
		proportions.StrategicReserve.String()[:6], "strategic reserve")

	s.Require().Equal(v10.CommunityPoolSecurityBudgetProportion,
		proportions.CommunityPoolSecurityBudget.String()[:6], "community pool security")

	// Check initial deposit ratio
	govParams := s.App.GovKeeper.GetParams(s.Ctx)
	s.Require().Equal(v10.MinInitialDepositRatio, govParams.MinInitialDepositRatio, "min initial deposit ratio")
}

func (s *UpgradeTestSuite) createCallbackData(id string, callback deprecatedproto.Message) icacallbackstypes.CallbackData {
	return icacallbackstypes.CallbackData{
		CallbackId:   id,
		CallbackArgs: s.mustMarshalCallback(callback),
	}
}

func (s *UpgradeTestSuite) mustMarshalCallback(callback deprecatedproto.Message) []byte {
	callbackBz, err := deprecatedproto.Marshal(callback)
	s.Require().NoError(err)
	return callbackBz
}

func (s *UpgradeTestSuite) mustUnmarshalCallback(callbackBz []byte, callback cosmosproto.Message) {
	err := cosmosproto.Unmarshal(callbackBz, callback)
	s.Require().NoError(err)
}

func (s *UpgradeTestSuite) TestMigrateCallbackData() {
	// Build dummy callback data for each callback type
	initialClaimCallbackArgs := stakeibctypes.ClaimCallback{
		UserRedemptionRecordId: "record-0",
		ChainId:                "chain-0",
		EpochNumber:            1,
	}
	initialDelegateCallbackArgs := stakeibctypes.DelegateCallback{
		HostZoneId:      "host-0",
		DepositRecordId: 1,
		SplitDelegations: []*types.SplitDelegation{{
			Validator: "val-0",
			Amount:    sdkmath.NewInt(1),
		}},
	}
	initialRebalanceCallbackArgs := stakeibctypes.RebalanceCallback{
		HostZoneId: "host-0",
		Rebalancings: []*stakeibctypes.Rebalancing{
			{
				SrcValidator: "val-0",
				DstValidator: "val-1",
				Amt:          sdkmath.NewInt(1),
			},
		},
	}
	initialRedemptionCallbackArgs := stakeibctypes.RedemptionCallback{
		HostZoneId:              "host-0",
		EpochUnbondingRecordIds: []uint64{1, 2, 3},
	}
	initialReinvestCallbackArgs := stakeibctypes.ReinvestCallback{
		HostZoneId:     "host-0",
		ReinvestAmount: sdk.NewCoin("denom", sdkmath.NewInt(1)),
	}
	initialUndelegateCallbackArgs := stakeibctypes.UndelegateCallback{
		HostZoneId: "host-0",
		SplitDelegations: []*types.SplitDelegation{{
			Validator: "val-0",
			Amount:    sdkmath.NewInt(1),
		}},
	}
	initialTransferCallbackArgs := recordstypes.TransferCallback{
		DepositRecordId: 1,
	}

	// Store the callback data
	initialCallbackData := []icacallbackstypes.CallbackData{
		s.createCallbackData(stakeibckeeper.ICACallbackID_Claim, &initialClaimCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Delegate, &initialDelegateCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Rebalance, &initialRebalanceCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Redemption, &initialRedemptionCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Reinvest, &initialReinvestCallbackArgs),
		s.createCallbackData(stakeibckeeper.ICACallbackID_Undelegate, &initialUndelegateCallbackArgs),
		s.createCallbackData(recordskeeper.TRANSFER, &initialTransferCallbackArgs),
	}
	for i := range initialCallbackData {
		initialCallbackData[i].CallbackKey = fmt.Sprintf("key-%d", i)
		initialCallbackData[i].PortId = fmt.Sprintf("port-%d", i)
		initialCallbackData[i].ChannelId = fmt.Sprintf("channel-%d", i)
		s.App.IcacallbacksKeeper.SetCallbackData(s.Ctx, initialCallbackData[i])
	}

	// Migrate the callbacks
	err := v10.MigrateCallbackData(s.Ctx, s.App.IcacallbacksKeeper)
	s.Require().NoError(err, "no error expected when migrating callback data")

	// Check that we can successfully unmarshal each callback with the new type
	finalCallbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(finalCallbackData, len(initialCallbackData), "callback data length")

	for i, finalCallback := range finalCallbackData {
		initialCallback := initialCallbackData[i]
		s.Require().Equal(initialCallback.CallbackId, finalCallback.CallbackId, "callback id for %d", i)

		callbackId := initialCallback.CallbackId
		s.Require().Equal(initialCallback.CallbackKey, finalCallback.CallbackKey, "callback key for %s", callbackId)
		s.Require().Equal(initialCallback.PortId, finalCallback.PortId, "callback port for %s", callbackId)
		s.Require().Equal(initialCallback.ChannelId, finalCallback.ChannelId, "callback channel for %s", callbackId)

		switch callbackId {
		case stakeibckeeper.ICACallbackID_Claim:
			var finalCallbackArgs stakeibctypes.ClaimCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialClaimCallbackArgs, finalCallbackArgs, "claim callback")

		case stakeibckeeper.ICACallbackID_Delegate:
			var finalCallbackArgs stakeibctypes.DelegateCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialDelegateCallbackArgs, finalCallbackArgs, "delegate callback")

		case stakeibckeeper.ICACallbackID_Rebalance:
			var finalCallbackArgs stakeibctypes.RebalanceCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialRebalanceCallbackArgs, finalCallbackArgs, "rebalance callback")

		case stakeibckeeper.ICACallbackID_Redemption:
			var finalCallbackArgs stakeibctypes.RedemptionCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialRedemptionCallbackArgs, finalCallbackArgs, "redemption callback")

		case stakeibckeeper.ICACallbackID_Reinvest:
			var finalCallbackArgs stakeibctypes.ReinvestCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialReinvestCallbackArgs, finalCallbackArgs, "reinvest callback")

		case stakeibckeeper.ICACallbackID_Undelegate:
			var finalCallbackArgs stakeibctypes.UndelegateCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialUndelegateCallbackArgs, finalCallbackArgs, "undelegate callback")

		case recordskeeper.TRANSFER:
			var finalCallbackArgs recordstypes.TransferCallback
			s.mustUnmarshalCallback(finalCallback.CallbackArgs, &finalCallbackArgs)
			s.Require().Equal(initialTransferCallbackArgs, finalCallbackArgs, "transfer callback")
		}
	}
}
