package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	_ "github.com/stretchr/testify/suite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type ValidatorUnbonding struct {
	Validator    string
	UnbondAmount sdkmath.Int
}

type UnbondingTestCase struct {
	hostZone                   types.HostZone
	totalUnbondAmount          sdkmath.Int
	delegationChannelID        string
	delegationPortID           string
	channelStartSequence       uint64
	expectedUnbondingRecordIds []uint64
}

func (s *KeeperTestSuite) SetupTestUnbondFromHostZone(
	totalWeight int64,
	totalStake sdkmath.Int,
	unbondAmount sdkmath.Int,
	validators []*types.Validator,
) UnbondingTestCase {
	delegationAccountOwner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	// Sanity checks:
	//  - total stake matches
	//  - total weights sum to 100
	actualTotalStake := sdkmath.ZeroInt()
	actualTotalWeights := uint64(0)
	for _, validator := range validators {
		actualTotalStake = actualTotalStake.Add(validator.Delegation)
		actualTotalWeights += validator.Weight
	}
	s.Require().Equal(totalStake.Int64(), actualTotalStake.Int64(), "test setup failed - total stake does not match")
	s.Require().Equal(totalWeight, int64(actualTotalWeights), "test setup failed - total weight does not match")

	// Store the validators on the host zone
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		Validators:           validators,
		TotalDelegations:     totalStake,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Store the total unbond amount across two epoch unbonding records
	halfUnbondAmount := unbondAmount.Quo(sdkmath.NewInt(2))
	for i := uint64(1); i <= 2; i++ {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: i,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					NativeTokenAmount: halfUnbondAmount,
				},
			},
		})
	}

	// Mock the epoch tracker to timeout 90% through the epoch
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
		hostZone:                   hostZone,
		totalUnbondAmount:          unbondAmount,
		delegationChannelID:        delegationChannelID,
		delegationPortID:           delegationPortID,
		channelStartSequence:       startSequence,
		expectedUnbondingRecordIds: []uint64{1, 2},
	}
}

// Helper function to check that an undelegation ICA was submitted and that the callback data
// holds the expected unbondings for each validator
func (s *KeeperTestSuite) CheckUnbondingMessages(tc UnbondingTestCase, expectedUnbondings []ValidatorUnbonding) {
	// Trigger unbonding
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().NoError(err, "no error expected when calling unbond from host")

	// Check that sequence number incremented from a sent ICA
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence+1, endSequence, "sequence number should have incremented")

	// Check that callback data was stored
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 1, "there should only be one callback data stored")

	// Check host zone and epoch unbonding record id's
	var actualCallback types.UndelegateCallback
	err = proto.Unmarshal(callbackData[0].CallbackArgs, &actualCallback)
	s.Require().NoError(err, "no error expected when unmarshalling callback args")

	s.Require().Equal(HostChainId, actualCallback.HostZoneId, "chain-id on callback")
	s.Require().Equal(tc.expectedUnbondingRecordIds, actualCallback.EpochUnbondingRecordIds, "unbonding record id's on callback")

	// Check splits from callback data align with expected unbondings
	s.Require().Len(actualCallback.SplitDelegations, len(expectedUnbondings), "number of unbonding messages")
	for i, expected := range expectedUnbondings {
		actualSplit := actualCallback.SplitDelegations[i]
		s.Require().Equal(expected.Validator, actualSplit.Validator, "callback message validator - index %d", i)
		s.Require().Equal(expected.UnbondAmount.Int64(), actualSplit.Amount.Int64(), "callback message amount - index %d", i)
	}

	// Check the delegation change in progress was incremented from each that had an unbonding
	actualHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")

	for _, actualValidator := range actualHostZone.Validators {
		validatorUnbonded := false
		for _, unbondedVal := range expectedUnbondings {
			if actualValidator.Address == unbondedVal.Validator {
				validatorUnbonded = true
			}
		}

		expectedDelegationChangesInProgress := 0
		if validatorUnbonded {
			expectedDelegationChangesInProgress = 1
		}
		s.Require().Equal(expectedDelegationChangesInProgress, int(actualValidator.DelegationChangesInProgress),
			"validator %s delegation changes in progress", actualValidator.Address)
	}

	// Check that the unbond event was emitted with the proper unbond amount
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyTotalUnbondAmount, tc.totalUnbondAmount.String())
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_UnbondOnlyZeroWeightVals() {
	// Native Stake:       1000
	// LSM Stake:           250
	// Total Stake:        1250
	//
	// Unbond Amount:        50
	// Stake After Unbond: 1200
	totalUnbondAmount := sdkmath.NewInt(50)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 1200 = 120, Capacity: 100-120 = -20 -> 0
		// No capacity -> unbondings
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 1200 = 420, Capacity: 420-420 = 0
		// No capacity -> unbondings
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 1200 = 360, Capacity: 300-360 = -60 -> 0
		// No capacity -> unbondings
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 1200 = 120, Capacity: 200 - 120 = 80
		// >>> Ratio: 110/200 = 0.55 -> #3 Priority <<<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 1200 = 180, Capacity: 160-180 = -20 -> 0
		// No capacity -> unbondings
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 50 - 40 = 10 unbond remaining
		// valE has #2 priority - unbond up to remaining
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(10)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_UnbondTotalLessThanTotalLSM() {
	// Native Stake:       1000
	// LSM Stake:           250
	// Total Stake:        1250
	//
	// Unbond Amount:       150
	// Stake After Unbond: 1100
	totalUnbondAmount := sdkmath.NewInt(150)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 1100 = 110, Capacity: 100-110 = -10 -> 0
		// No capacity -> unbondings
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 1100 = 385, Capacity: 420-385 = 35
		// >>> Ratio: 385/420 = 0.91 -> Priority #4 <<<
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 1100 = 330, Capacity: 300-330 = -30 -> 0
		// No capacity -> unbondings
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 1100 = 110, Capacity: 200 - 110 = 90
		// >>> Ratio: 110/200 = 0.55 -> Priority #3 <<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 1100 = 165, Capacity: 160-165 = -5 -> 0
		// No capacity -> unbondings
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 150 - 40 = 110 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 150 - 40 - 30 = 80 unbond remaining
		// valF has #3 priority - unbond up to remaining
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(80)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_UnbondTotalGreaterThanTotalLSM() {
	// Native Stake: 1000
	// LSM Stake:     250
	// Total Stake:  1250
	//
	// Unbond Amount:      350
	// Stake After Unbond: 900
	totalUnbondAmount := sdkmath.NewInt(350)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 900 = 90, Capacity: 100-90 = 10
		// >>> Ratio: 90/100 = 0.9 -> Priority #7 <<<
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 900 = 315, Capacity: 420-315 = 105
		// >>> Ratio: 315/420 = 0.75 -> Priority #4 <<<
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 900 = 270, Capacity: 300-270 = 30
		// >>> Ratio: 270/300 = 0.9 -> Priority #6 <<<
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 900 = 90, Capacity: 200 - 90 = 110
		// >>> Ratio: 90/200 = 0.45 -> Priority #3 <<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 900 = 135, Capacity: 160-135 = 25
		// >>> Ratio: 135/160 = 0.85 -> Priority #5 <<<
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 350 - 40 = 310 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 310 - 30 = 280 unbond remaining
		// valF has #3 priority - unbond up to capacity at 110
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(110)},
		// 280 - 110 = 170 unbond remaining
		// valB has #4 priority - unbond up to capacity at 105
		{Validator: "valB", UnbondAmount: sdkmath.NewInt(105)},
		// 170 - 105 = 65 unbond remaining
		// valG has #5 priority - unbond up to capacity at 25
		{Validator: "valG", UnbondAmount: sdkmath.NewInt(25)},
		// 65 - 25 = 40 unbond remaining
		// valD has #6 priority - unbond up to capacity at 30
		{Validator: "valD", UnbondAmount: sdkmath.NewInt(30)},
		// 40 - 30 = 10 unbond remaining
		// valA has #7 priority - unbond up to remaining
		{Validator: "valA", UnbondAmount: sdkmath.NewInt(10)},
	}

	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUnbondingMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_NoDelegationAccount() {
	// Call unbond on a host zone without a delegation account - it should error
	invalidHostZone := types.HostZone{
		ChainId:              HostChainId,
		DelegationIcaAddress: "",
	}
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "no delegation account found for GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_ZeroUnbondAmount() {
	totalWeight := int64(0)
	totalStake := sdkmath.ZeroInt()
	totalUnbondAmount := sdkmath.ZeroInt()
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, []*types.Validator{})

	// Call unbond - it should NOT error since the unbond amount was 0 - but it should short circuit
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().Nil(err, "unbond should not have thrown an error - it should have simply ignored the host zone")

	// Confirm no ICAs were sent
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence, endSequence, "sequence number should stay the same since no messages were sent")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_ZeroValidatorWeights() {
	// Setup the test with all zero-weight validators
	totalWeight := int64(0)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(10)
	validators := []*types.Validator{
		{Address: "valA", Weight: 0, Delegation: sdkmath.NewInt(25)},
		{Address: "valB", Weight: 0, Delegation: sdkmath.NewInt(50)},
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(25)},
	}
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)

	// Call unbond - it should fail
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().ErrorContains(err, "No non-zero validators found for host zone")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_InsufficientDelegations() {
	// Setup the test where the total unbond amount is greater than the current delegations
	totalWeight := int64(100)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(200)
	validators := []*types.Validator{
		{Address: "valA", Weight: 25, Delegation: sdkmath.NewInt(25)},
		{Address: "valB", Weight: 50, Delegation: sdkmath.NewInt(50)},
		{Address: "valC", Weight: 25, Delegation: sdkmath.NewInt(25)},
	}
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)

	// Call unbond - it should fail
	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, tc.hostZone)
	s.Require().ErrorContains(err, "Cannot calculate target delegation if final amount is less than or equal to zero")
}

func (s *KeeperTestSuite) TestUnbondFromHostZone_ICAFailed() {
	// Validator setup here is arbitrary as long as the totals match
	totalWeight := int64(100)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(10)
	validators := []*types.Validator{{Address: "valA", Weight: 100, Delegation: sdkmath.NewInt(100)}}
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, totalUnbondAmount, validators)

	// Remove the connection ID from the host zone so that the ICA fails
	invalidHostZone := tc.hostZone
	invalidHostZone.ConnectionId = ""

	err := s.App.StakeibcKeeper.UnbondFromHostZone(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "unable to submit unbonding ICA for GAIA")
}

func (s *KeeperTestSuite) TestGetBalanceRatio() {
	testCases := []struct {
		unbondCapacity keeper.ValidatorUnbondCapacity
		expectedRatio  sdk.Dec
		errorExpected  bool
	}{
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(0),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.ZeroDec(),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(25),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.MustNewDecFromStr("0.25"),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(75),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.MustNewDecFromStr("0.75"),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(150),
				CurrentDelegation:  sdkmath.NewInt(100),
			},
			expectedRatio: sdk.MustNewDecFromStr("1.5"),
			errorExpected: false,
		},
		{
			unbondCapacity: keeper.ValidatorUnbondCapacity{
				BalancedDelegation: sdkmath.NewInt(100),
				CurrentDelegation:  sdkmath.NewInt(0),
			},
			errorExpected: true,
		},
	}
	for _, tc := range testCases {
		balanceRatio, err := tc.unbondCapacity.GetBalanceRatio()
		if tc.errorExpected {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedRatio.String(), balanceRatio.String())
		}
	}
}

func (s *KeeperTestSuite) TestGetTotalUnbondAmountAndRecordsIds() {
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: uint64(1),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(1),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(2),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
		{
			EpochNumber: uint64(2),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(3),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(4),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
		{
			EpochNumber: uint64(3),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Status
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(5),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
				{
					// Different Status
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(6),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
			},
		},
		{
			EpochNumber: uint64(4),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Host and Status
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(7),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(8),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
	}

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	expectedUnbondAmount := int64(1 + 3 + 8)
	expectedRecordIds := []uint64{1, 2, 4}

	actualUnbondAmount, actualRecordIds := s.App.StakeibcKeeper.GetTotalUnbondAmountAndRecordsIds(s.Ctx, HostChainId)
	s.Require().Equal(expectedUnbondAmount, actualUnbondAmount.Int64(), "unbonded amount")
	s.Require().Equal(expectedRecordIds, actualRecordIds, "epoch unbonding record IDs")
}

func (s *KeeperTestSuite) TestGetValidatorUnbondCapacity() {
	// Start with the expected returned list of validator capacities
	expectedUnbondCapacity := []keeper.ValidatorUnbondCapacity{
		{
			ValidatorAddress:   "valA",
			CurrentDelegation:  sdkmath.NewInt(50),
			BalancedDelegation: sdkmath.NewInt(0),
			Capacity:           sdkmath.NewInt(50),
		},
		{
			ValidatorAddress:   "valB",
			CurrentDelegation:  sdkmath.NewInt(200),
			BalancedDelegation: sdkmath.NewInt(5),
			Capacity:           sdkmath.NewInt(195),
		},
		{
			ValidatorAddress:   "valC",
			CurrentDelegation:  sdkmath.NewInt(1089),
			BalancedDelegation: sdkmath.NewInt(1000),
			Capacity:           sdkmath.NewInt(89),
		},
	}

	// Build list of input validators and map of balanced delegations from expected list
	validators := []*types.Validator{}
	balancedDelegations := map[string]sdkmath.Int{}
	for _, validatorCapacity := range expectedUnbondCapacity {
		validators = append(validators, &types.Validator{
			Address:    validatorCapacity.ValidatorAddress,
			Delegation: validatorCapacity.CurrentDelegation,
		})
		balancedDelegations[validatorCapacity.ValidatorAddress] = validatorCapacity.BalancedDelegation
	}

	// Add validators with no capacity - none of these should be in the returned list
	deficits := []int64{0, 10, 50}
	valAddresses := []string{"valD", "valE", "valF"}
	for i, deficit := range deficits {
		address := valAddresses[i]

		// the delegation amount is arbitrary here
		// all that mattesr is that it's less than the balance delegation
		currentDelegation := sdkmath.NewInt(50)
		balancedDelegation := currentDelegation.Add(sdkmath.NewInt(deficit))

		validators = append(validators, &types.Validator{
			Address:    address,
			Delegation: currentDelegation,
		})
		balancedDelegations[address] = balancedDelegation
	}

	// Check capacity matches expectations
	actualUnbondCapacity := s.App.StakeibcKeeper.GetValidatorUnbondCapacity(s.Ctx, validators, balancedDelegations)
	s.Require().Len(actualUnbondCapacity, len(expectedUnbondCapacity), "number of expected unbondings")

	for i, expected := range expectedUnbondCapacity {
		address := expected.ValidatorAddress
		actual := actualUnbondCapacity[i]
		s.Require().Equal(expected.ValidatorAddress, actual.ValidatorAddress, "address for %s", address)
		s.Require().Equal(expected.CurrentDelegation.Int64(), actual.CurrentDelegation.Int64(), "current for %s", address)
		s.Require().Equal(expected.BalancedDelegation.Int64(), actual.BalancedDelegation.Int64(), "balanced for %s", address)
		s.Require().Equal(expected.Capacity.Int64(), actual.Capacity.Int64(), "capacity for %s", address)
	}
}

func (s *KeeperTestSuite) TestSortUnbondingCapacityByPriority() {
	// First we define what the ideal list will look like after sorting
	expectedSortedCapacities := []keeper.ValidatorUnbondCapacity{
		// Zero-weight validator's
		{
			// (1) Ratio: 0, Capacity: 100
			ValidatorAddress:   "valE",
			BalancedDelegation: sdkmath.NewInt(0),
			CurrentDelegation:  sdkmath.NewInt(100), // ratio = 0/100
			Capacity:           sdkmath.NewInt(100),
		},
		{
			// (2) Ratio: 0, Capacity: 25
			ValidatorAddress:   "valC",
			BalancedDelegation: sdkmath.NewInt(0),
			CurrentDelegation:  sdkmath.NewInt(25), // ratio = 0/25
			Capacity:           sdkmath.NewInt(25),
		},
		{
			// (3) Ratio: 0, Capacity: 25
			// Same ratio and capacity as above but name is tie breaker
			ValidatorAddress:   "valD",
			BalancedDelegation: sdkmath.NewInt(0),
			CurrentDelegation:  sdkmath.NewInt(25), // ratio = 0/25
			Capacity:           sdkmath.NewInt(25),
		},
		// Non-zero-weight validator's
		{
			// (4) Ratio: 0.1
			ValidatorAddress:   "valB",
			BalancedDelegation: sdkmath.NewInt(1),
			CurrentDelegation:  sdkmath.NewInt(10), // ratio = 1/10
			Capacity:           sdkmath.NewInt(9),
		},
		{
			// (5) Ratio: 0.25
			ValidatorAddress:   "valH",
			BalancedDelegation: sdkmath.NewInt(250),
			CurrentDelegation:  sdkmath.NewInt(1000), // ratio = 250/1000
			Capacity:           sdkmath.NewInt(750),
		},
		{
			// (6) Ratio: 0.5, Capacity: 100
			ValidatorAddress:   "valF",
			BalancedDelegation: sdkmath.NewInt(100),
			CurrentDelegation:  sdkmath.NewInt(200), // ratio = 100/200
			Capacity:           sdkmath.NewInt(100),
		},
		{
			// (7) Ratio: 0.5, Capacity: 100
			// Same ratio and capacity as above - name is tie breaker
			ValidatorAddress:   "valI",
			BalancedDelegation: sdkmath.NewInt(100),
			CurrentDelegation:  sdkmath.NewInt(200), // ratio = 100/200
			Capacity:           sdkmath.NewInt(100),
		},
		{
			// (8) Ratio: 0.5, Capacity: 50
			// Same ratio as above but capacity is lower
			ValidatorAddress:   "valG",
			BalancedDelegation: sdkmath.NewInt(50),
			CurrentDelegation:  sdkmath.NewInt(100), // ratio = 50/100
			Capacity:           sdkmath.NewInt(50),
		},
		{
			// (9) Ratio: 0.6
			ValidatorAddress:   "valA",
			BalancedDelegation: sdkmath.NewInt(6),
			CurrentDelegation:  sdkmath.NewInt(10), // ratio = 6/10
			Capacity:           sdkmath.NewInt(4),
		},
	}

	// Define the shuffled ordering of the array above by just specifying
	// the validator addresses an a randomized order
	shuffledOrder := []string{
		"valA",
		"valD",
		"valG",
		"valF",
		"valE",
		"valB",
		"valH",
		"valI",
		"valC",
	}

	// Use ordering above in combination with the data structures from the
	// expected list to shuffle the expected list into a list that will be the
	// input to this function
	inputCapacities := []keeper.ValidatorUnbondCapacity{}
	for _, shuffledValAddress := range shuffledOrder {
		for _, capacity := range expectedSortedCapacities {
			if capacity.ValidatorAddress == shuffledValAddress {
				inputCapacities = append(inputCapacities, capacity)
			}
		}
	}

	// Sort the list
	actualSortedCapacities, err := keeper.SortUnbondingCapacityByPriority(inputCapacities)
	s.Require().NoError(err)
	s.Require().Len(actualSortedCapacities, len(expectedSortedCapacities), "number of capacities")

	// To make the error easier to understand, we first compare just the list of validator addresses
	actualValidators := []string{}
	for _, actual := range actualSortedCapacities {
		actualValidators = append(actualValidators, actual.ValidatorAddress)
	}
	expectedValidators := []string{}
	for _, expected := range expectedSortedCapacities {
		expectedValidators = append(expectedValidators, expected.ValidatorAddress)
	}
	s.Require().Equal(expectedValidators, actualValidators, "validator order")

	// Then we'll do a sanity check on each field
	// If the above passes and this fails, that likely means the test was setup improperly
	for i, expected := range expectedSortedCapacities {
		actual := actualSortedCapacities[i]
		address := expected.ValidatorAddress
		s.Require().Equal(expected.ValidatorAddress, actual.ValidatorAddress, "validator %d address", i+1)
		s.Require().Equal(expected.BalancedDelegation, actual.BalancedDelegation, "validator %s balanced", address)
		s.Require().Equal(expected.CurrentDelegation, actual.CurrentDelegation, "validator %s current", address)
		s.Require().Equal(expected.Capacity, actual.Capacity, "validator %s capacity", address)
	}
}

func (s *KeeperTestSuite) TestGetUnbondingICAMessages() {
	delegationAddress := "cosmos_DELEGATION"

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		DelegationIcaAddress: delegationAddress,
	}

	validatorCapacities := []keeper.ValidatorUnbondCapacity{
		{ValidatorAddress: "val1", Capacity: sdkmath.NewInt(100)},
		{ValidatorAddress: "val2", Capacity: sdkmath.NewInt(200)},
		{ValidatorAddress: "val3", Capacity: sdkmath.NewInt(300)},
		{ValidatorAddress: "val4", Capacity: sdkmath.NewInt(400)},
	}

	testCases := []struct {
		name               string
		totalUnbondAmount  sdkmath.Int
		expectedUnbondings []ValidatorUnbonding
		expectedError      string
	}{
		{
			name:              "unbond val1 partially",
			totalUnbondAmount: sdkmath.NewInt(50),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(50)},
			},
		},
		{
			name:              "unbond val1 fully",
			totalUnbondAmount: sdkmath.NewInt(100),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
			},
		},
		{
			name:              "unbond val1 fully and val2 partially",
			totalUnbondAmount: sdkmath.NewInt(200),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(100)},
			},
		},
		{
			name:              "unbond val1 val2 fully",
			totalUnbondAmount: sdkmath.NewInt(300),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
			},
		},
		{
			name:              "unbond val1 val2 fully and val3 partially",
			totalUnbondAmount: sdkmath.NewInt(450),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
				{Validator: "val3", UnbondAmount: sdkmath.NewInt(150)},
			},
		},
		{
			name:              "unbond val1 val2 and val3 fully",
			totalUnbondAmount: sdkmath.NewInt(600),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
				{Validator: "val3", UnbondAmount: sdkmath.NewInt(300)},
			},
		},
		{
			name:              "full unbonding",
			totalUnbondAmount: sdkmath.NewInt(1000),
			expectedUnbondings: []ValidatorUnbonding{
				{Validator: "val1", UnbondAmount: sdkmath.NewInt(100)},
				{Validator: "val2", UnbondAmount: sdkmath.NewInt(200)},
				{Validator: "val3", UnbondAmount: sdkmath.NewInt(300)},
				{Validator: "val4", UnbondAmount: sdkmath.NewInt(400)},
			},
		},
		{
			name:              "insufficient delegation",
			totalUnbondAmount: sdkmath.NewInt(1001),
			expectedError:     "unable to unbond full amount",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Get the unbonding ICA messages for the test case
			actualMessages, actualSplits, actualError := s.App.StakeibcKeeper.GetUnbondingICAMessages(
				hostZone,
				tc.totalUnbondAmount,
				validatorCapacities,
			)

			// If this is an error test case, check the error message
			if tc.expectedError != "" {
				s.Require().ErrorContains(actualError, tc.expectedError, "error expected")
				return
			}

			// For the success case, check the error number of unbondings
			s.Require().NoError(actualError, "no error expected when unbonding %v", tc.totalUnbondAmount)
			s.Require().Len(actualMessages, len(tc.expectedUnbondings), "number of undelegate messages")
			s.Require().Len(actualSplits, len(tc.expectedUnbondings), "number of validator splits")

			// Check each unbonding
			for i, expected := range tc.expectedUnbondings {
				valAddress := expected.Validator
				actualMsg := actualMessages[i].(*stakingtypes.MsgUndelegate)
				actualSplit := actualSplits[i]

				// Check the ICA message
				s.Require().Equal(valAddress, actualMsg.ValidatorAddress, "ica message validator")
				s.Require().Equal(delegationAddress, actualMsg.DelegatorAddress, "ica message delegator for %s", valAddress)
				s.Require().Equal(Atom, actualMsg.Amount.Denom, "ica message denom for %s", valAddress)
				s.Require().Equal(expected.UnbondAmount.Int64(), actualMsg.Amount.Amount.Int64(),
					"ica message amount for %s", valAddress)

				// Check the callback
				s.Require().Equal(expected.Validator, actualSplit.Validator, "callback validator for %s", valAddress)
				s.Require().Equal(expected.UnbondAmount.Int64(), actualSplit.Amount.Int64(), "callback amount %s", valAddress)
			}
		})
	}
}
