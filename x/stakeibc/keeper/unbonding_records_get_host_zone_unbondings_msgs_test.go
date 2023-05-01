package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"

	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type UnbondingTestCase struct {
	totalUnbondAmount     sdkmath.Int
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZone              types.HostZone
	lightClientTime       uint64
	totalWgt              uint64
	expectedUnbondings    map[string]int64
	valNames              []string
	delegationChannelID   string
	delegationPortID      string
	channelStartSequence  uint64
}

func (s *KeeperTestSuite) SetupUnbonding() UnbondingTestCase {
	delegationAccountOwner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	hostVal1Addr := "cosmos_VALIDATOR_1"
	hostVal2Addr := "cosmos_VALIDATOR_2"
	hostVal3Addr := "cosmos_VALIDATOR_3"
	hostVal4Addr := "cosmos_VALIDATOR_4"
	valNames := []string{hostVal1Addr, hostVal2Addr, hostVal3Addr, hostVal4Addr}

	totalUnbondAmount := sdkmath.NewInt(1_000_000)
	amtVal1 := sdkmath.NewInt(1_000_000)
	amtVal2 := sdkmath.NewInt(2_000_000)
	wgtVal1 := uint64(1)
	wgtVal2 := uint64(2)
	totalWgt := uint64(5)

	expectedUnbondings := map[string]int64{
		hostVal1Addr: 200_000,
		hostVal2Addr: 400_000,
		hostVal3Addr: 400_000,
		hostVal4Addr: 0,
	}

	// 2022-08-12T19:51, a random time in the past
	unbondingTime := uint64(1660348276)
	lightClientTime := unbondingTime + 1

	//  define the host zone with total delegation and validators with staked amounts
	validators := []*types.Validator{
		{
			Address:    hostVal1Addr,
			Delegation: amtVal1,
			Weight:     wgtVal1,
		},
		{
			Address:    hostVal2Addr,
			Delegation: amtVal2,
			Weight:     wgtVal2,
		},
		{
			Address: hostVal3Addr,
			// Delegation and Weight are the same as Val2, to test tie breaking
			Delegation: amtVal2,
			Weight:     wgtVal2,
		},
		{
			Address: hostVal4Addr,
			// Zero weight validator
			Delegation: sdkmath.NewInt(0),
			Weight:     0,
		},
	}

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		Bech32Prefix:         "cosmos",
		Validators:           validators,
		DelegationIcaAddress: "cosmos_DELEGATION",
		ConnectionId:         ibctesting.FirstConnectionID,
	}

	// list of epoch unbonding records
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber:        0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
		{
			EpochNumber:        1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
	}

	// for each epoch unbonding record, add a host zone unbonding record and append the record
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
			NativeTokenAmount: totalUnbondAmount.Quo(sdkmath.NewInt(2)),
			Denom:             "uatom",
			HostZoneId:        "GAIA",
			UnbondingTime:     unbondingTime, // 2022-08-12T19:52
			Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		}
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// This will make the current time 90% through the epoch
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,                                                // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// Get tx seq number before the ICA was submitted to check whether an ICA was submitted
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, delegationPortID, delegationChannelID)
	s.Require().True(found, "sequence number not found before ica")

	return UnbondingTestCase{
		totalUnbondAmount:     totalUnbondAmount,
		hostZone:              hostZone,
		epochUnbondingRecords: epochUnbondingRecords,
		lightClientTime:       lightClientTime, // 2022-08-12T19:51.000001, 1ns after the unbonding time
		totalWgt:              totalWgt,
		valNames:              valNames,
		expectedUnbondings:    expectedUnbondings,
		delegationChannelID:   delegationChannelID,
		delegationPortID:      delegationPortID,
		channelStartSequence:  startSequence,
	}
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful() {
	tc := s.SetupUnbonding()

	// Submit unbonding
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().NoError(err)

	// verify the callback attributes are as expected
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 1, "there should only be one callback data stored")

	var actualCallbackResult types.UndelegateCallback
	err = proto.Unmarshal(callbackData[0].CallbackArgs, &actualCallbackResult)
	s.Require().NoError(err, "no error expected when unmarshalling callback args")

	// The zero weight validator should not be included in the message
	// Confirm the callback data
	expectedNumMessages := len(tc.hostZone.Validators) - 1
	s.Require().Equal(tc.hostZone.ChainId, actualCallbackResult.HostZoneId, "host zone id in success unbonding case")
	s.Require().Len(actualCallbackResult.SplitDelegations, expectedNumMessages, "number of split delegations in success unbonding case")

	// Confirm unbonding amounts
	actualUnbondings := make(map[string]int64)
	for _, split := range actualCallbackResult.SplitDelegations {
		actualUnbondings[split.Validator] = split.Amount.Int64()
	}
	actualUnbondings[tc.valNames[3]] = 0 // zero-weight val

	for _, valAddress := range tc.valNames {
		s.Require().Equal(tc.expectedUnbondings[valAddress], actualUnbondings[valAddress], "val %s address", valAddress)
	}

	// Confirm the channel sequence number incremented with the ICA send
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence+1, endSequence, "sequence number should have incremented")

	// TODO [LSM] Check that event was emitted with proper unbond amount
	// expectedUnbondAmount := tc.amtToUnbond.Mul(sdkmath.NewInt(int64(len(tc.epochUnbondingRecords))))
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_WrongChainId() {
	tc := s.SetupUnbonding()

	// Change the chainId so that it is not found in the store
	tc.hostZone.ChainId = "nonExistentChainId"

	// Call unbond - we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().Nil(err, "unbond should not have thrown an error - it should have simply ignored the host zone")

	// Confirm no ICAs were sent
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence, endSequence, "sequence number should stay the same since no messages were sent")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_NoEpochUnbondingRecords() {
	tc := s.SetupUnbonding()

	// Delete all the epoch unbonding records
	for i := range tc.epochUnbondingRecords {
		s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, uint64(i))
	}

	// Confirm they were deleted
	epochUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	s.Require().Empty(epochUnbondingRecords, "number of epoch unbonding records should be 0 after deletion")

	// Call unbond - we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().Nil(err, "unbond should not have thrown an error - it should have simply ignored the host zone")

	// Confirm no ICAs were sent
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence, endSequence, "sequence number should stay the same since no messages were sent")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_UnbondingTooMuch() {
	tc := s.SetupUnbonding()

	// iterate the validators and set all their delegated amounts to 0
	for i := range tc.hostZone.Validators {
		tc.hostZone.Validators[i].Delegation = sdkmath.ZeroInt()
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	// Call unbond - it should fail from insufficient delegation
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().EqualError(err,
		fmt.Sprintf("Could not unbond %v on Host Zone %s, unable to balance the unbond amount across validators",
			tc.totalUnbondAmount, tc.hostZone.ChainId))
}
