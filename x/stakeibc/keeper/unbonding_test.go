package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	_ "github.com/stretchr/testify/suite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
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
	delegationAccountOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
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
		RedemptionRate:       sdk.OneDec(),
		MaxMessagesPerIcaTx:  32,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Store the total unbond amount across two epoch unbonding records
	// and create a user redemption record for each
	halfUnbondAmount := unbondAmount.Quo(sdkmath.NewInt(2))
	for i := uint64(1); i <= 2; i++ {
		redemptionRecordId := fmt.Sprintf("id-%d", i)

		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: i,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:            HostChainId,
					Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					StTokenAmount:         halfUnbondAmount,
					NativeTokenAmount:     halfUnbondAmount,
					UserRedemptionRecords: []string{redemptionRecordId},
				},
			},
		})

		s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
			Id:                redemptionRecordId,
			StTokenAmount:     halfUnbondAmount,
			NativeTokenAmount: halfUnbondAmount,
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
	s.Require().Len(actualCallback.SplitUndelegations, len(expectedUnbondings), "number of unbonding messages")
	for i, expected := range expectedUnbondings {
		actualSplit := actualCallback.SplitUndelegations[i]
		s.Require().Equal(expected.Validator, actualSplit.Validator, "callback message validator - index %d", i)
		s.Require().Equal(expected.UnbondAmount.Int64(), actualSplit.NativeTokenAmount.Int64(), "callback message amount - index %d", i)
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

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_UnbondIgnoresSlashQueryInProgress() {
	// Native Stake:       100
	// LSM Stake:           0
	// Total Stake:        100
	//
	// Slash Query In Progress Stake: 25
	// Eligible Stake: 		75
	//
	// Unbond Amount:        20
	// Stake After Unbond: 80
	// Eligible Stake After Unbond 45
	totalUnbondAmount := sdkmath.NewInt(20)
	totalStake := sdkmath.NewInt(100)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 25, Weight: 15%, Balanced: (15/75) * 55= 11, Capacity: 25-11 = 14 > 0
		{Address: "valA", Weight: 15, Delegation: sdkmath.NewInt(25)},
		// Current: 25, Weight: 20%, Balanced: (20/75) * 55 = 14.66, Capacity: 25-14.66 = 10.44 > 0
		{Address: "valB", Weight: 20, Delegation: sdkmath.NewInt(25)},
		// Current: 25, Weight: 40%, Balanced: (40/75) * 55 = 29.33, Capacity: 25-29.33 < 0
		{Address: "valC", Weight: 40, Delegation: sdkmath.NewInt(25)},
		// Current: 25, Weight: 25%, Slash-Query-In-Progress so ignored
		{Address: "valD", Weight: 25, Delegation: sdkmath.NewInt(25), SlashQueryInProgress: true},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valA has #1 priority - unbond up to 14
		{Validator: "valA", UnbondAmount: sdkmath.NewInt(14)},
		// 20 - 14 = 6 unbond remaining
		// valB has #2 priority - unbond up to remaining
		{Validator: "valB", UnbondAmount: sdkmath.NewInt(6)},
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

func (s *KeeperTestSuite) TestUnbondFromHostZone_Successful_RefreshedNativeAmount() {
	// Total Stake: 1000
	//
	// Unbond Amount with old redemption rate (RR = 1): 100
	// Unbond Amount with new redemption rate (RR = 1.5): 150
	//
	// Stake After Unbond: 850
	updatedRedemptionRate := sdk.MustNewDecFromStr("1.5")
	unbondAmountWithOldRate := sdkmath.NewInt(100)
	unbondAmountWithNewRate := sdkmath.NewInt(150)
	totalStake := sdkmath.NewInt(1000)
	totalWeight := int64(100)

	// Since this test is only intended to check the native token refresh,
	// we don't need more than 1 validator
	// That validator should unbond the full amount with the new redemption rate
	validators := []*types.Validator{
		{Address: "valA", Weight: 100, Delegation: totalStake},
	}
	expectedUnbondings := []ValidatorUnbonding{
		{Validator: "valA", UnbondAmount: unbondAmountWithNewRate},
	}

	// Setup using default, and then override the redemption rate value to update it from 1.0 to 1.5
	tc := s.SetupTestUnbondFromHostZone(totalWeight, totalStake, unbondAmountWithOldRate, validators)

	hostZone := s.MustGetHostZone(HostChainId)
	hostZone.RedemptionRate = updatedRedemptionRate
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Finally check that the unbondings matched - mostly checking that there was a greater amount
	// unbonded than was originally in the host zone unbonding record
	tc.totalUnbondAmount = unbondAmountWithNewRate
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

func (s *KeeperTestSuite) TestGetQueuedHostZoneUnbondingRecords() {
	// This function returns a mapping of epoch unbonding record ID (i.e. epoch number) -> hostZoneUnbonding
	// For the purposes of this test, the NativeTokenAmount is used in place of the host zone unbonding record
	// for the purposes of validating the proper record was selected. In other words, after this function,
	// we just verify that the native token amounts of the output line up with the expected map below
	expectedEpochUnbondingRecordIds := []uint64{1, 2, 4}
	expectedHostZoneUnbondingMap := map[uint64]int64{1: 1, 2: 3, 4: 8} // includes only the relevant records below

	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			// Has relevant host zone unbonding, so epoch number is included
			EpochNumber: uint64(1),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Included
					HostZoneId:    HostChainId,
					StTokenAmount: sdkmath.NewInt(1),
					Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:    OsmoChainId,
					StTokenAmount: sdkmath.NewInt(2),
					Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
		{
			// Has relevant host zone unbonding, so epoch number is included
			EpochNumber: uint64(2),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Included
					HostZoneId:    HostChainId,
					StTokenAmount: sdkmath.NewInt(3),
					Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:    OsmoChainId,
					StTokenAmount: sdkmath.NewInt(4),
					Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
		{
			// No relevant host zone unbonding, epoch number not included
			EpochNumber: uint64(3),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Status
					HostZoneId:    HostChainId,
					StTokenAmount: sdkmath.NewInt(5),
					Status:        recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
				{
					// Different Status
					HostZoneId:    OsmoChainId,
					StTokenAmount: sdkmath.NewInt(6),
					Status:        recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
			},
		},
		{
			// Has relevant host zone unbonding (the retry one), so epoch number is included
			EpochNumber: uint64(4),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Host and Status
					HostZoneId:    OsmoChainId,
					StTokenAmount: sdkmath.NewInt(7),
					Status:        recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					// Included
					HostZoneId:                HostChainId,
					StTokenAmount:             sdkmath.NewInt(8),
					Status:                    recordtypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE,
					UndelegationTxsInProgress: 0,
				},
			},
		},
		{
			// Has no relevant unbondings
			EpochNumber: uint64(5),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Host and Status
					HostZoneId:    OsmoChainId,
					StTokenAmount: sdkmath.NewInt(9),
					Status:        recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					// Retry record, but has an undelegation in progress
					HostZoneId:                HostChainId,
					StTokenAmount:             sdkmath.NewInt(10),
					Status:                    recordtypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE,
					UndelegationTxsInProgress: 1,
				},
			},
		},
	}

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	actualEpochIds, actualHostZoneMap := s.App.StakeibcKeeper.GetQueuedHostZoneUnbondingRecords(s.Ctx, HostChainId)
	s.Require().Equal(expectedEpochUnbondingRecordIds, actualEpochIds, "epoch unbonding record IDs")
	for epochNumber, actualHostZoneUnbonding := range actualHostZoneMap {
		expectedHostZoneUnbonding := expectedHostZoneUnbondingMap[epochNumber]
		s.Require().Equal(expectedHostZoneUnbonding, actualHostZoneUnbonding.StTokenAmount.Int64(),
			"host zone unbonding record sttoken amount")
	}
}

func (s *KeeperTestSuite) TestGetTotalUnbondAmount() {
	hostZoneUnbondingRecords := map[uint64]recordtypes.HostZoneUnbonding{
		1: {NativeTokensToUnbond: sdkmath.NewInt(1)},
		2: {NativeTokensToUnbond: sdkmath.NewInt(2)},
		3: {NativeTokensToUnbond: sdkmath.NewInt(3)},
		4: {NativeTokensToUnbond: sdkmath.NewInt(4)},
	}
	expectedUnbondAmount := sdkmath.NewInt(1 + 2 + 3 + 4)

	actualUnbondAmount := s.App.StakeibcKeeper.GetTotalUnbondAmount(hostZoneUnbondingRecords)
	s.Require().Equal(expectedUnbondAmount, actualUnbondAmount, "unbond amount")

	emptyUnbondings := map[uint64]recordtypes.HostZoneUnbonding{}
	actualUnbondAmount = s.App.StakeibcKeeper.GetTotalUnbondAmount(emptyUnbondings)
	s.Require().Zero(actualUnbondAmount.Int64(), "expected zero unbondings")
}

func (s *KeeperTestSuite) TestRefreshUserRedemptionRecordNativeAmounts() {
	// Define the expected redemption records after the function is called
	redemptionRate := sdk.MustNewDecFromStr("1.999")
	expectedUserRedemptionRecords := []recordtypes.UserRedemptionRecord{
		// StTokenAmount: 1000 * 1.999 = 1999 Native
		{Id: "A", StTokenAmount: sdkmath.NewInt(1000), NativeTokenAmount: sdkmath.NewInt(1999)},
		// StTokenAmount: 999 * 1.999 = 1997.001, Truncated to 1997 Native
		{Id: "B", StTokenAmount: sdkmath.NewInt(999), NativeTokenAmount: sdkmath.NewInt(1997)},
		// StTokenAmount: 100 * 1.999 = 199.9, Truncated to 199 Native
		{Id: "C", StTokenAmount: sdkmath.NewInt(100), NativeTokenAmount: sdkmath.NewInt(199)},
	}
	expectedTotalNativeAmount := sdkmath.NewInt(1999 + 1997 + 199)

	// Create the initial records which do not have the end native amount
	for _, expectedUserRedemptionRecord := range expectedUserRedemptionRecords {
		initialUserRedemptionRecord := expectedUserRedemptionRecord
		initialUserRedemptionRecord.NativeTokenAmount = sdkmath.ZeroInt()
		s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, initialUserRedemptionRecord)
	}

	// Call the refresh user redemption record function
	// Note: an extra ID ("D"), is passed into this function that will be ignored
	// since there's not user redemption record for "D"
	redemptionRecordIds := []string{"A", "B", "C", "D"}
	actualTotalNativeAmount := s.App.StakeibcKeeper.RefreshUserRedemptionRecordNativeAmounts(
		s.Ctx,
		HostChainId,
		redemptionRecordIds,
		redemptionRate,
	)

	// Confirm the summation is correct and the user redemption records were updated
	s.Require().Equal(expectedTotalNativeAmount.Int64(), actualTotalNativeAmount.Int64(), "total native amount")
	for _, expectedRecord := range expectedUserRedemptionRecords {
		actualRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, expectedRecord.Id)
		s.Require().True(found, "record %s should have been found", expectedRecord.Id)
		s.Require().Equal(expectedRecord.NativeTokenAmount.Int64(), actualRecord.NativeTokenAmount.Int64(),
			"record %s native amount", expectedRecord.Id)
	}
}

// Tests RefreshUnbondingNativeTokenAmounts which indirectly tests
// RefreshHostZoneUnbondingNativeTokenAmount and RefreshUserRedemptionRecordNativeAmounts
func (s *KeeperTestSuite) TestRefreshUnbondingNativeTokenAmounts() {
	chainA := "chain-0"
	chainB := "chain-1"
	epochNumberA := uint64(1)
	epochNumberB := uint64(2)

	// Create the epoch unbonding records
	// It doesn't need the host zone unbonding records since they'll be added
	// in the tested function
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumberA,
	})
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumberB,
	})

	// Create two host zones, with different redemption rates
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:        chainA,
		RedemptionRate: sdk.MustNewDecFromStr("1.5"),
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:        chainB,
		RedemptionRate: sdk.MustNewDecFromStr("2.0"),
	})

	// Create the user redemption records
	userRedemptionRecords := []recordtypes.UserRedemptionRecord{
		// chainA - Redemption Rate: 1.5
		{Id: "A", StTokenAmount: sdkmath.NewInt(1000)}, // native: 1500
		{Id: "B", StTokenAmount: sdkmath.NewInt(2000)}, // native: 3000
		// chainB - Redemption Rate: 2.0
		{Id: "C", StTokenAmount: sdkmath.NewInt(3000)}, // native: 6000
		{Id: "D", StTokenAmount: sdkmath.NewInt(4000)}, // native: 8000
	}
	expectedUserNativeAmounts := map[string]sdkmath.Int{
		"A": sdkmath.NewInt(1500),
		"B": sdkmath.NewInt(3000),
		"C": sdkmath.NewInt(6000),
		"D": sdkmath.NewInt(8000),
	}
	for _, redemptionRecord := range userRedemptionRecords {
		s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, redemptionRecord)
	}

	// Define the two host zone unbonding records
	initialHostZoneUnbondingA := recordtypes.HostZoneUnbonding{
		HostZoneId:            chainA,
		UserRedemptionRecords: []string{"A", "B"},
		StTokenAmount:         sdkmath.NewInt(10_000),
		Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	expectedHostZoneUnbondAmountA := expectedUserNativeAmounts["A"].Add(expectedUserNativeAmounts["B"]).Int64()
	expectedStToBurnAmountA := initialHostZoneUnbondingA.StTokenAmount.Int64()

	initialHostZoneUnbondingB := recordtypes.HostZoneUnbonding{
		HostZoneId:            chainB,
		UserRedemptionRecords: []string{"C", "D"},
		StTokenAmount:         sdkmath.NewInt(20_000),
		Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	expectedHostZoneUnbondAmountB := expectedUserNativeAmounts["C"].Add(expectedUserNativeAmounts["D"]).Int64()
	expectedStToBurnAmountB := initialHostZoneUnbondingB.StTokenAmount.Int64()

	// Call refresh for both hosts
	epochToHostZoneMap := map[uint64]recordtypes.HostZoneUnbonding{
		epochNumberA: initialHostZoneUnbondingA,
		epochNumberB: initialHostZoneUnbondingB,
	}
	refreshedEpochToHostZoneMapA, err := s.App.StakeibcKeeper.RefreshUnbondingNativeTokenAmounts(s.Ctx, chainA, epochToHostZoneMap)
	s.Require().NoError(err, "no error expected when refreshing unbond amount")
	refreshedEpochToHostZoneMapB, err := s.App.StakeibcKeeper.RefreshUnbondingNativeTokenAmounts(s.Ctx, chainB, epochToHostZoneMap)
	s.Require().NoError(err, "no error expected when refreshing unbond amount")

	// Confirm the host zone unbonding records were updated
	updatedHostZoneUnbondingA, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochNumberA, chainA)
	s.Require().True(found, "host zone unbonding record for %s should have been found", chainA)

	actualNativeAmountA := updatedHostZoneUnbondingA.NativeTokenAmount.Int64()
	actualNativeToUnbondAmountA := updatedHostZoneUnbondingA.NativeTokensToUnbond.Int64()
	actualStToBurnA := updatedHostZoneUnbondingA.StTokensToBurn.Int64()
	s.Require().Equal(expectedHostZoneUnbondAmountA, actualNativeAmountA, "host zone unbonding native amount A")
	s.Require().Equal(expectedHostZoneUnbondAmountA, actualNativeToUnbondAmountA, "host zone unbonding amount to unbond A")
	s.Require().Equal(expectedStToBurnAmountA, actualStToBurnA, "host zone unbonding amount to burn A")

	updatedHostZoneUnbondingB, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochNumberB, chainB)
	s.Require().True(found, "host zone unbonding record for %s should have been found", chainB)

	actualNativeAmountB := updatedHostZoneUnbondingB.NativeTokenAmount.Int64()
	actualNativeToUnbondAmountB := updatedHostZoneUnbondingB.NativeTokensToUnbond.Int64()
	actualStToBurnB := updatedHostZoneUnbondingB.StTokensToBurn.Int64()
	s.Require().Equal(expectedHostZoneUnbondAmountB, actualNativeAmountB, "host zone unbonding native amount B")
	s.Require().Equal(expectedHostZoneUnbondAmountB, actualNativeToUnbondAmountB, "host zone unbonding amount to unbond B")
	s.Require().Equal(expectedStToBurnAmountB, actualStToBurnB, "host zone unbonding amount to burn B")

	// Confirm all user redemption records were updated
	for id, expectedNativeAmount := range expectedUserNativeAmounts {
		record, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, id)
		s.Require().True(found, "user redemption record for %s should have been found", id)
		s.Require().Equal(expectedNativeAmount, record.NativeTokenAmount, "user redemption record %s native amount", id)
	}

	// Confirm the returned map also has the updated values
	returnedNativeAmountA := refreshedEpochToHostZoneMapA[epochNumberA].NativeTokensToUnbond.Int64()
	returnedNativeAmountB := refreshedEpochToHostZoneMapB[epochNumberB].NativeTokensToUnbond.Int64()
	s.Require().Equal(expectedHostZoneUnbondAmountA, returnedNativeAmountA, "returned map native amount A")
	s.Require().Equal(expectedHostZoneUnbondAmountB, returnedNativeAmountB, "returned map native amount B")

	// Remove one of the host zones and confirm it errors
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, chainA)
	_, err = s.App.StakeibcKeeper.RefreshUnbondingNativeTokenAmounts(s.Ctx, chainA, epochToHostZoneMap)
	s.Require().ErrorContains(err, "host zone not found")
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
				s.Require().Equal(expected.UnbondAmount.Int64(), actualSplit.NativeTokenAmount.Int64(), "callback native amount %s", valAddress)
			}
		})
	}
}

func (s *KeeperTestSuite) TestBatchSubmitUndelegateICAMessages() {
	// The test will submit ICA's across 10 validators, in batches of 3
	// There should be 4 ICA's submitted
	batchSize := 3
	numValidators := 10
	expectedNumberOfIcas := 4
	epochUnbondingRecordIds := []uint64{1} // arbitrary

	// Create the delegation ICA channel
	delegationAccountOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	// Create a host zone
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
	}

	// Build the ICA messages and callback for each validator
	var validators []*types.Validator
	var undelegateMsgs []proto.Message
	var unbondings []*types.SplitUndelegation
	for i := 0; i < numValidators; i++ {
		validatorAddress := fmt.Sprintf("val%d", i)
		validators = append(validators, &types.Validator{Address: validatorAddress})

		undelegateMsgs = append(undelegateMsgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: validatorAddress,
			Amount:           sdk.NewCoin(hostZone.HostDenom, sdkmath.NewInt(100)),
		})

		unbondings = append(unbondings, &types.SplitUndelegation{
			Validator:         validatorAddress,
			NativeTokenAmount: sdkmath.NewInt(100),
		})
	}

	// Store the validators on the host zone
	hostZone.Validators = validators
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock the epoch tracker to timeout 90% through the epoch
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,                                                // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// Get tx seq number before the ICA was submitted to check whether an ICA was submitted
	startSequence := s.MustGetNextSequenceNumber(delegationPortID, delegationChannelID)

	// Submit the unbondings
	numTxsSubmitted, err := s.App.StakeibcKeeper.BatchSubmitUndelegateICAMessages(
		s.Ctx,
		hostZone,
		epochUnbondingRecordIds,
		undelegateMsgs,
		unbondings,
		batchSize,
	)
	s.Require().NoError(err, "no error expected when submitting batches")
	s.Require().Equal(numTxsSubmitted, uint64(expectedNumberOfIcas), "returned number of txs submitted")

	// Confirm the sequence number iterated by the expected number of ICAs
	endSequence := s.MustGetNextSequenceNumber(delegationPortID, delegationChannelID)
	s.Require().Equal(startSequence+uint64(expectedNumberOfIcas), endSequence, "expected number of ICA submissions")

	// Confirm the number of callback data's matches the expected number of ICAs
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Equal(expectedNumberOfIcas, len(callbackData), "number of callback datas")

	// Remove the connection ID from the host zone and try again, it should fail
	invalidHostZone := hostZone
	invalidHostZone.ConnectionId = ""
	_, err = s.App.StakeibcKeeper.BatchSubmitUndelegateICAMessages(
		s.Ctx,
		invalidHostZone,
		epochUnbondingRecordIds,
		undelegateMsgs,
		unbondings,
		batchSize,
	)
	s.Require().ErrorContains(err, "unable to submit unbonding ICA")
}

func (s *KeeperTestSuite) SetupInitiateAllHostZoneUnbondings() {
	s.CreateICAChannel("GAIA.DELEGATION")

	gaiaValAddr := "cosmos_VALIDATOR"
	osmoValAddr := "osmo_VALIDATOR"
	gaiaDelegationAddr := "cosmos_DELEGATION"
	osmoDelegationAddr := "osmo_DELEGATION"

	//  define the host zone with total delegation and validators with staked amounts
	gaiaValidators := []*types.Validator{
		{
			Address:    gaiaValAddr,
			Delegation: sdkmath.NewInt(5_000_000),
			Weight:     uint64(10),
		},
		{
			Address:    gaiaValAddr + "2",
			Delegation: sdkmath.NewInt(3_000_000),
			Weight:     uint64(6),
		},
	}
	osmoValidators := []*types.Validator{
		{
			Address:    osmoValAddr,
			Delegation: sdkmath.NewInt(5_000_000),
			Weight:     uint64(10),
		},
	}
	hostZones := []types.HostZone{
		{
			ChainId:              HostChainId,
			HostDenom:            Atom,
			Bech32Prefix:         GaiaPrefix,
			UnbondingPeriod:      14,
			Validators:           gaiaValidators,
			DelegationIcaAddress: gaiaDelegationAddr,
			TotalDelegations:     sdkmath.NewInt(5_000_000),
			ConnectionId:         ibctesting.FirstConnectionID,
			RedemptionRate:       sdk.OneDec(),
			MaxMessagesPerIcaTx:  32,
		},
		{
			ChainId:              OsmoChainId,
			HostDenom:            Osmo,
			Bech32Prefix:         OsmoPrefix,
			UnbondingPeriod:      21,
			Validators:           osmoValidators,
			DelegationIcaAddress: osmoDelegationAddr,
			TotalDelegations:     sdkmath.NewInt(5_000_000),
			ConnectionId:         ibctesting.FirstConnectionID,
			RedemptionRate:       sdk.OneDec(),
			MaxMessagesPerIcaTx:  32,
		},
	}
	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// list of epoch unbonding records
	epochNumber := uint64(5)

	redemptionRecordId1 := recordtypes.UserRedemptionRecordKeyFormatter(HostChainId, epochNumber, "receiver")
	redemptionRecordId2 := recordtypes.UserRedemptionRecordKeyFormatter(OsmoChainId, epochNumber, "receiver")

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{
				HostZoneId:            HostChainId,
				StTokenAmount:         sdkmath.NewInt(1_900_000),
				NativeTokenAmount:     sdkmath.NewInt(2_000_000),
				Denom:                 Atom,
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				UserRedemptionRecords: []string{redemptionRecordId1},
			},
			{
				HostZoneId:            OsmoChainId,
				StTokenAmount:         sdkmath.NewInt(2_800_000),
				NativeTokenAmount:     sdkmath.NewInt(3),
				Denom:                 Osmo,
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				UserRedemptionRecords: []string{redemptionRecordId2},
			},
		},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:            redemptionRecordId1,
		HostZoneId:    HostChainId,
		EpochNumber:   epochNumber,
		StTokenAmount: epochUnbondingRecord.HostZoneUnbondings[0].StTokenAmount,
	})

	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:            redemptionRecordId2,
		HostZoneId:    OsmoChainId,
		EpochNumber:   epochNumber,
		StTokenAmount: epochUnbondingRecord.HostZoneUnbondings[1].StTokenAmount,
	})

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    "day",
		EpochNumber:        12,
		NextEpochStartTime: uint64(2661750006000000000), // arbitrary time in the future, year 2056 I believe
		Duration:           uint64(1000000000000),       // 16 min 40 sec
	})
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Successful() {
	// tests that we can successful initiate a host zone unbonding for GAIA and OSMO
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)

	// An event should be emitted for each if they were successful
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_GaiaSuccessful() {
	// Tests that if we initiate unbondings a day where only Gaia is supposed to unbond, it succeeds and Osmo is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 9)

	// An event should only be emitted for Gaia
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_OsmoSuccessful() {
	// Tests that if we initiate unbondings a day where only Osmo is supposed to unbond, it succeeds and Gaia is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 8)

	// An event should only be emitted for Osmo
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_NoneSuccessful() {
	// Tests that if we initiate unbondings a day where none are supposed to unbond, it works successfully
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 10)

	// No event should be emitted for either host
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Failed() {
	// Tests that if Gaia doesn't have enough delegated stake to unbond, it fails
	// but Osmo does and is successful
	s.SetupInitiateAllHostZoneUnbondings()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	hostZone.Validators = []*types.Validator{
		{
			Address:    "cosmos_VALIDATOR",
			Delegation: sdkmath.NewInt(1_000_000),
			Weight:     uint64(10),
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	hostZone, _ = s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)

	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)

	// An event should only be emitted for Osmo
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)

}
