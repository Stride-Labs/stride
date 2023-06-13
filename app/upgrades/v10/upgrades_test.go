package v10_test

import (
	"fmt"
	"testing"
	"time"

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

	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
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

	// Remove localhost client from client keeper
	clientParams := s.App.IBCKeeper.ClientKeeper.GetParams(s.Ctx)
	clientParams.AllowedClients = []string{}
	s.App.IBCKeeper.ClientKeeper.SetParams(s.Ctx, clientParams)

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

	// Check localhost client was added
	clientParams = s.App.IBCKeeper.ClientKeeper.GetParams(s.Ctx)
	s.Require().Contains(clientParams.AllowedClients, "09-localhost")
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

func (s *UpgradeTestSuite) TestMigrateDistributorAddress() {
	ck := s.App.ClaimKeeper

	// Airdrops
	strideAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "stride",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "stride-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(4143585840),
		DistributorAddress: "stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev",
	}

	gaiaAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "gaia",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "cosmoshub-4",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(191138794312),
		DistributorAddress: "stride1fmh0ysk5nt9y2cj8hddms5ffj2dhys55xkkjwz",
	}

	osmosisAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "osmosis",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "osmosis-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(72895369704),
		DistributorAddress: "stride1zlu2l3lx5tqvzspvjwsw9u0e907kelhqae3yhk",
	}

	junoAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "juno",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "juno-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(10967183382),
		DistributorAddress: "stride14k9g9zpgaycpey9840nnpa66l4nd6lu7g7t74c",
	}

	starsAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "stars",
		AirdropStartTime:   time.Date(2022, 11, 22, 14, 53, 52, 0, time.UTC),
		AutopilotEnabled:   false,
		ChainId:            "stargaze-1",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(1013798205),
		DistributorAddress: "stride12pum4adk5dhp32d90f8g8gfwujm0gwxqnrdlum",
	}

	evmosAirdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "evmos",
		AirdropStartTime:   time.Date(2023, 4, 3, 16, 0, 0, 0, time.UTC),
		AutopilotEnabled:   true,
		ChainId:            "evmos_9001-2",
		ClaimDenom:         "ustrd",
		ClaimedSoFar:       sdk.NewInt(13491005333),
		DistributorAddress: "stride10dy5pmc2fq7fnmufjfschkfrxaqnpykl6ezy5j",
	}

	claimParams, err := ck.GetParams(s.Ctx)
	s.Require().NoError(err, "no error expected when getting claim params")

	// Set the airdrops on the claim params
	claimParams.Airdrops = []*claimtypes.Airdrop{&strideAirdrop, &gaiaAirdrop, &osmosisAirdrop, &junoAirdrop, &starsAirdrop, &evmosAirdrop}
	err = ck.SetParams(s.Ctx, claimParams)
	s.Require().NoError(err, "no error expected when setting claim params")

	// Migrate the airdrop distributor addresses
	err = v10.MigrateClaimDistributorAddress(s.Ctx, ck)
	s.Require().NoError(err, "no error expected when migrating distributor addresses")

	// Iterate allAirdrops and make sure the updated airdrops are equivalent, except for the distributor address
	for _, airdrop := range claimParams.Airdrops {
		airdropToCompare := ck.GetAirdropByIdentifier(s.Ctx, airdrop.AirdropIdentifier)
		s.Require().Equal(airdropToCompare.AirdropIdentifier, airdrop.AirdropIdentifier, "airdrop identifier")
		s.Require().Equal(airdropToCompare.AirdropStartTime, airdrop.AirdropStartTime, "airdrop start time")
		s.Require().Equal(airdropToCompare.AutopilotEnabled, airdrop.AutopilotEnabled, "airdrop autopilot enabled")
		s.Require().Equal(airdropToCompare.ChainId, airdrop.ChainId, "airdrop chain id")
		s.Require().Equal(airdropToCompare.ClaimDenom, airdrop.ClaimDenom, "airdrop claim denom")
		s.Require().Equal(airdropToCompare.ClaimedSoFar, airdrop.ClaimedSoFar, "airdrop claimed so far")

		newDistributorAddress := v10.NewDistributorAddresses[airdrop.AirdropIdentifier]
		s.Require().Equal(newDistributorAddress, airdropToCompare.DistributorAddress, "airdrop distributor address updated")
	}

}
