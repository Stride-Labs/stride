package keeper_test

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/Stride-Labs/stride/v29/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v29/x/stakeibc/types"
)

func (s *KeeperTestSuite) createNHostZone(n int) []types.HostZone {
	items := make([]types.HostZone, n)
	for i := range items {
		items[i].ChainId = strconv.Itoa(i)
		items[i].RedemptionRate = sdkmath.LegacyNewDec(1)
		items[i].LastRedemptionRate = sdkmath.LegacyNewDec(1)
		items[i].MinRedemptionRate = sdkmath.LegacyNewDecWithPrec(5, 1)
		items[i].MaxRedemptionRate = sdkmath.LegacyNewDecWithPrec(15, 1)
		items[i].MinInnerRedemptionRate = sdkmath.LegacyNewDecWithPrec(5, 1)
		items[i].MaxInnerRedemptionRate = sdkmath.LegacyNewDecWithPrec(15, 1)
		items[i].TotalDelegations = sdkmath.ZeroInt()
		items[i].Validators = []*types.Validator{}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, items[i])
	}
	return items
}

func (s *KeeperTestSuite) TestHostZoneGet() {
	items := s.createNHostZone(10)
	for _, item := range items {
		got, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, item.ChainId)
		s.Require().True(found)
		s.Require().Equal(
			&item,
			&got,
		)
	}
}

func (s *KeeperTestSuite) TestGetHostZoneFromHostDenom() {
	// Create host zones with different denoms
	hostZones := s.createNHostZone(5)
	denoms := []string{}
	for i, hostZone := range hostZones {
		denom := fmt.Sprintf("denom-%d", i)
		denoms = append(denoms, denom)
		hostZone.HostDenom = denom

		hostZones[i] = hostZone
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Search each host zone by the denom
	for i, denom := range denoms {
		hostZone, err := s.App.StakeibcKeeper.GetHostZoneFromHostDenom(s.Ctx, denom)
		s.Require().NoError(err)
		s.Require().Equal(hostZones[i], *hostZone, "host zone")
	}

	// Confirm non-existent denom errors
	_, err := s.App.StakeibcKeeper.GetHostZoneFromHostDenom(s.Ctx, "non-existent")
	s.Require().ErrorContains(err, "No HostZone for non-existent denom found")
}

func (s *KeeperTestSuite) TestGetHostZoneFromIBCDenom() {
	// Create host zones with different denoms
	hostZones := s.createNHostZone(5)
	denoms := []string{}
	for i, hostZone := range hostZones {
		denom := fmt.Sprintf("denom-%d", i)
		denoms = append(denoms, denom)
		hostZone.IbcDenom = denom

		hostZones[i] = hostZone
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Search each host zone by the denom
	for i, denom := range denoms {
		hostZone, err := s.App.StakeibcKeeper.GetHostZoneFromIBCDenom(s.Ctx, denom)
		s.Require().NoError(err)
		s.Require().Equal(hostZones[i], *hostZone, "host zone")
	}

	// Confirm non-existent denom errors
	_, err := s.App.StakeibcKeeper.GetHostZoneFromIBCDenom(s.Ctx, "non-existent")
	s.Require().ErrorContains(err, "No HostZone for non-existent found")
}

func (s *KeeperTestSuite) TestHostZoneRemove() {
	items := s.createNHostZone(10)
	for _, item := range items {
		s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, item.ChainId)
		_, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, item.ChainId)
		s.Require().False(found)
	}
}

func (s *KeeperTestSuite) TestHostZoneGetAll() {
	items := s.createNHostZone(10)
	s.Require().ElementsMatch(
		items,
		s.App.StakeibcKeeper.GetAllHostZone(s.Ctx),
	)
}

func (s *KeeperTestSuite) TestHostZoneGetAllActiveCase1() {
	// Case 1: some active some inactive
	numZones := 3
	items := s.createNHostZone(numZones)
	// set the last host zone as halted
	items[numZones-1].Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, items[numZones-1])

	// only the last host zone is active, so we expect all except that one
	actualActiveHzs := items[:numZones-1]
	getActiveHzResults := s.App.StakeibcKeeper.GetAllActiveHostZone(s.Ctx)
	s.Require().ElementsMatch(
		actualActiveHzs,
		getActiveHzResults,
	)
}

func (s *KeeperTestSuite) TestHostZoneGetAllActiveCase2() {
	// Case 2: all active
	numZones := 3
	items := s.createNHostZone(numZones)
	s.Require().ElementsMatch(
		items,
		s.App.StakeibcKeeper.GetAllActiveHostZone(s.Ctx),
	)
}

func (s *KeeperTestSuite) TestHostZoneGetAllActiveCase3() {
	// Case 3: all inactive
	numZones := 3
	items := s.createNHostZone(numZones)
	// set the last host zone as halted
	items[0].Halted = true
	items[1].Halted = true
	items[2].Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, items[0])
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, items[1])
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, items[2])
	s.Require().ElementsMatch(
		types.HostZone{},
		s.App.StakeibcKeeper.GetAllActiveHostZone(s.Ctx),
	)
}

func (s *KeeperTestSuite) TestHostZoneGetAllActiveCase4() {
	// create no zones, check the output is an empty list
	s.Require().ElementsMatch(
		types.HostZone{},
		s.App.StakeibcKeeper.GetAllActiveHostZone(s.Ctx),
	)
}

func (s *KeeperTestSuite) TestGetValidatorFromAddress() {
	numValidators := 3

	// Create list of validators
	addresses := []string{}
	validators := []*types.Validator{}
	for i := 1; i <= numValidators; i++ {
		address := fmt.Sprintf("val-%d", i)

		addresses = append(addresses, address)
		validators = append(validators, &types.Validator{Address: address})
	}

	// For each validator that was just added, test GetValidatorFromAddress
	for expectedIndex, address := range addresses {
		expectedValidator := *validators[expectedIndex]
		actualValidator, actualIndex, found := keeper.GetValidatorFromAddress(validators, address)

		s.Require().True(found)
		s.Require().Equal(expectedValidator, actualValidator)
		s.Require().Equal(int64(expectedIndex), actualIndex)
	}

	// Test GetValidatorFromAddress for an validator that doesn't exist
	_, _, found := keeper.GetValidatorFromAddress(validators, "fake_validator")
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestGetHostZoneFromTransferChannelID() {
	// Store 5 host zones
	expectedHostZones := map[string]types.HostZone{}
	for i := 0; i < 5; i++ {
		chainId := fmt.Sprintf("chain-%d", i)
		channelId := fmt.Sprintf("channel-%d", i)

		hostZone := types.HostZone{
			ChainId:           chainId,
			TransferChannelId: channelId,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
		expectedHostZones[channelId] = hostZone
	}

	// Look up each host zone by the channel ID
	for i := 0; i < 5; i++ {
		channelId := fmt.Sprintf("channel-%d", i)

		expectedHostZone := expectedHostZones[channelId]
		actualHostZone, found := s.App.StakeibcKeeper.GetHostZoneFromTransferChannelID(s.Ctx, channelId)

		s.Require().True(found, "found host zone %d", i)
		s.Require().Equal(expectedHostZone.ChainId, actualHostZone.ChainId, "host zone %d chain-id", i)
	}

	// Lookup a non-existent host zone - should not be found
	_, found := s.App.StakeibcKeeper.GetHostZoneFromTransferChannelID(s.Ctx, "fake_channel")
	s.Require().False(found, "fake channel should not be found")
}

// Helper function to check the validator's slash query progress and checkpoint after it was incremented
func (s *KeeperTestSuite) checkValidatorSlashQueryProgress(address string, expectedProgress, expectedCheckpoint sdkmath.Int) {
	actualHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().Len(actualHostZone.Validators, 3, "host zone should still have 3 validators")

	actualValidator := types.Validator{}
	for _, validator := range actualHostZone.Validators {
		if validator.Address == address {
			actualValidator = *validator
		}
	}
	s.Require().NotEmpty(actualValidator.Address, "validator address not found")
	s.Require().Equal(expectedProgress.Int64(), actualValidator.SlashQueryProgressTracker.Int64(), "slash query progress")
	s.Require().Equal(expectedCheckpoint.Int64(), actualValidator.SlashQueryCheckpoint.Int64(), "slash query checkpoint")
}

func (s *KeeperTestSuite) TestIncrementValidatorSlashQueryProgress() {
	// Slash query progress for validator B is as follows:
	//  Initial Checkpoint: 1000 (from previous TVL)
	//  Current TVL: 10k, Threshold: 11% => New Checkpoint of 1100
	//  Old Progress: 7800 => Old Interval: 7800 / 1000 = Interval #7
	//  New Stake #1: 180 => New Interval: 8001 / 1000 = Interval #8
	incrementedValidator := "valB"
	threshold := uint64(11)
	totalStakeAmount := sdkmath.NewInt(10_000)

	initialCheckpoint := sdkmath.NewInt(1000)
	expectedCheckpoint := sdkmath.NewInt(1100)

	initialProgress := sdkmath.NewInt(7800)
	firstStakeAmount := sdkmath.NewInt(180)
	progressAfterFirstStake := sdkmath.NewInt(7980)
	secondStakeAmount := sdkmath.NewInt(100)
	progressAfterSecondStake := sdkmath.NewInt(8080)

	// Store a host zone with 3 validators and 1 in progress
	initialHostZone := types.HostZone{
		ChainId: HostChainId,
		Validators: []*types.Validator{
			{Address: "valA"},
			{Address: incrementedValidator, SlashQueryProgressTracker: initialProgress, SlashQueryCheckpoint: initialCheckpoint},
			{Address: "valC"},
		},
		TotalDelegations: totalStakeAmount,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, initialHostZone)

	// Set params with 10% threshold
	params := types.DefaultParams()
	params.ValidatorSlashQueryThreshold = threshold
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	// Increment the progress for valB by an amount that falls short of the checkpoint
	err := s.App.StakeibcKeeper.IncrementValidatorSlashQueryProgress(
		s.Ctx,
		HostChainId,
		incrementedValidator,
		firstStakeAmount,
	)
	s.Require().NoError(err, "no error expected when incrementing slash query progress")

	// Check progress was updated and checkpoint was not
	s.checkValidatorSlashQueryProgress(incrementedValidator, progressAfterFirstStake, initialCheckpoint)

	// Increment the progress again - this time it should increment the checkpoint
	err = s.App.StakeibcKeeper.IncrementValidatorSlashQueryProgress(
		s.Ctx,
		HostChainId,
		incrementedValidator,
		secondStakeAmount,
	)
	s.Require().NoError(err, "no error expected when incrementing slash query progress")

	// Check progress and checkpoint were updated
	s.checkValidatorSlashQueryProgress(incrementedValidator, progressAfterSecondStake, expectedCheckpoint)

	// Try to increment from a non-existed host chain - it should fail
	err = s.App.StakeibcKeeper.IncrementValidatorSlashQueryProgress(s.Ctx, "fake_host", incrementedValidator, firstStakeAmount)
	s.Require().ErrorContains(err, "host zone not found")

	// Try to increment from a non-existed validator - it should fail
	err = s.App.StakeibcKeeper.IncrementValidatorSlashQueryProgress(s.Ctx, HostChainId, "fake_val", firstStakeAmount)
	s.Require().ErrorContains(err, "validator not found")
}

// Tests Increment/DecrementValidatorDelegationsChangesInProgress
func (s *KeeperTestSuite) TestUpdateValidatorDelegationChangesInProgress() {
	hostZone := &types.HostZone{
		Validators: []*types.Validator{
			{Address: "other_val1", DelegationChangesInProgress: 1},
			{Address: ValAddress, DelegationChangesInProgress: 2},
			{Address: "other_val3", DelegationChangesInProgress: 3},
		},
		TotalDelegations: sdkmath.NewInt(6000),
	}
	updatedIndex := 1
	start := int(2)

	// Increment once - should end at 3
	err := s.App.StakeibcKeeper.IncrementValidatorDelegationChangesInProgress(hostZone, ValAddress)
	s.Require().NoError(err, "no error expected when incremented ")
	s.Require().Equal(start+1, int(hostZone.Validators[updatedIndex].DelegationChangesInProgress),
		"delegation change after increment")

	// Increment 10 more times - should end at 13
	for i := 0; i < 10; i++ {
		err := s.App.StakeibcKeeper.IncrementValidatorDelegationChangesInProgress(hostZone, ValAddress)
		s.Require().NoError(err, "no error expected when incrementing loop %d", i)
	}
	s.Require().Equal(start+11, int(hostZone.Validators[updatedIndex].DelegationChangesInProgress),
		"delegation change after increment loop")

	// Confirm the other validators did not change
	s.Require().Equal(1, int(hostZone.Validators[0].DelegationChangesInProgress),
		"delegation change val1 after increment")
	s.Require().Equal(3, int(hostZone.Validators[2].DelegationChangesInProgress),
		"delegation change val3 after increment")

	// Decrement - should end at 12
	err = s.App.StakeibcKeeper.DecrementValidatorDelegationChangesInProgress(hostZone, ValAddress)
	s.Require().NoError(err, "no error expected when decrementing")
	s.Require().Equal(start+10, int(hostZone.Validators[updatedIndex].DelegationChangesInProgress),
		"delegation change after decrement")

	// Decrement 12 more times - it should end at 0
	for i := 0; i < 12; i++ {
		err := s.App.StakeibcKeeper.DecrementValidatorDelegationChangesInProgress(hostZone, ValAddress)
		s.Require().NoError(err, "no error expected when decrementing loop %d", i)
	}
	s.Require().Equal(0, int(hostZone.Validators[updatedIndex].DelegationChangesInProgress),
		"delegation change after decrement loop")

	// Attempt to decrement again, it should fail
	err = s.App.StakeibcKeeper.DecrementValidatorDelegationChangesInProgress(hostZone, ValAddress)
	s.Require().ErrorContains(err, "cannot decrement the number of delegation updates")

	// Attempt to increment a non-existent validator - it should fail
	err = s.App.StakeibcKeeper.IncrementValidatorDelegationChangesInProgress(hostZone, "fake_val")
	s.Require().ErrorContains(err, "validator not found")

	// Attempt to decrement a non-existent validator - it should fail
	err = s.App.StakeibcKeeper.DecrementValidatorDelegationChangesInProgress(hostZone, "fake_val")
	s.Require().ErrorContains(err, "validator not found")
}

func (s *KeeperTestSuite) TestAddDelegationToValidator() {
	hostZone := &types.HostZone{
		Validators: []*types.Validator{
			{Address: "other_val1", Delegation: sdkmath.NewInt(1000)},
			{Address: ValAddress, Delegation: sdkmath.NewInt(2000)},
			{Address: "other_val2", Delegation: sdkmath.NewInt(3000)},
		},
		TotalDelegations: sdkmath.NewInt(6000),
	}
	updatedIndex := 1

	// Add 500 to the validator
	err := s.App.StakeibcKeeper.AddDelegationToValidator(s.Ctx, hostZone, ValAddress, sdkmath.NewInt(500), "")
	s.Require().NoError(err, "no error expected when adding delegation to validator")
	s.Require().Equal(int64(2500), hostZone.Validators[updatedIndex].Delegation.Int64(), "delegation after addition")
	s.Require().Equal(int64(6500), hostZone.TotalDelegations.Int64(), "total delegations after addition")

	// Subtract 250 from the validator
	err = s.App.StakeibcKeeper.AddDelegationToValidator(s.Ctx, hostZone, ValAddress, sdkmath.NewInt(-250), "")
	s.Require().NoError(err, "no error expected when subtracting delegation from validator")
	s.Require().Equal(int64(2250), hostZone.Validators[updatedIndex].Delegation.Int64(), "delegation after subtraction")
	s.Require().Equal(int64(6250), hostZone.TotalDelegations.Int64(), "total delegations after subtraction")

	// Confirm other validators were not modified
	s.Require().Equal(int64(1000), hostZone.Validators[0].Delegation.Int64(), "validator at index 0 should not have changed")
	s.Require().Equal(int64(3000), hostZone.Validators[2].Delegation.Int64(), "validator at index 2 should not have changed")

	// Attempt to subtract more than the validator has - it should fail
	err = s.App.StakeibcKeeper.AddDelegationToValidator(s.Ctx, hostZone, ValAddress, sdkmath.NewInt(-3000), "")
	s.Require().ErrorContains(err, "Delegation change (3000) is greater than validator")

	// Attempt to modify a validator that doesn't exist - it should fail
	err = s.App.StakeibcKeeper.AddDelegationToValidator(s.Ctx, hostZone, "does_not_exist", sdkmath.NewInt(1000), "")
	s.Require().ErrorContains(err, "validator not found")

	// Attempt to subtract more than the total delegations on the host - it should fail
	// Here, w set the validator's delegation to be much higher than the TotalDelegation
	//   (which should not be possible in practice)
	hostZone.Validators[updatedIndex].Delegation = sdkmath.NewInt(10000)
	err = s.App.StakeibcKeeper.AddDelegationToValidator(s.Ctx, hostZone, ValAddress, sdkmath.NewInt(-7000), "")
	s.Require().ErrorContains(err, "Delegation change (7000) is greater than total delegation amount on host")
}

func (s *KeeperTestSuite) TestCheckValidatorWeightsBelowCap() {
	testCases := []struct {
		name       string
		weightCap  uint64
		validators []*types.Validator
		exceedsCap bool
	}{
		{
			name:      "not enough validators",
			weightCap: 10,
			validators: []*types.Validator{
				{Address: "val1", Weight: 1},
				{Address: "val2", Weight: 1},
				{Address: "val3", Weight: 1},
				{Address: "val4", Weight: 1},
				{Address: "val5", Weight: 1},
				{Address: "val6", Weight: 1},
				{Address: "val7", Weight: 1},
				{Address: "val8", Weight: 1},
				{Address: "val9", Weight: 1},
			},
			exceedsCap: false,
		},
		{
			name:      "zero total weight",
			weightCap: 10,
			validators: []*types.Validator{
				{Address: "val1", Weight: 0},
				{Address: "val2", Weight: 0},
				{Address: "val3", Weight: 0},
				{Address: "val4", Weight: 0},
				{Address: "val5", Weight: 0},
				{Address: "val6", Weight: 0},
				{Address: "val7", Weight: 0},
				{Address: "val8", Weight: 0},
				{Address: "val9", Weight: 0},
				{Address: "val10", Weight: 0},
			},
			exceedsCap: false,
		},
		{
			name:      "10pct splits below threshold",
			weightCap: 11,
			validators: []*types.Validator{
				{Address: "val1", Weight: 1},
				{Address: "val2", Weight: 1},
				{Address: "val3", Weight: 1},
				{Address: "val4", Weight: 1},
				{Address: "val5", Weight: 1},
				{Address: "val6", Weight: 1},
				{Address: "val7", Weight: 1},
				{Address: "val8", Weight: 1},
				{Address: "val9", Weight: 1},
				{Address: "val10", Weight: 1},
			},
			exceedsCap: false,
		},
		{
			name:      "10pct splits at threshold",
			weightCap: 10,
			validators: []*types.Validator{
				{Address: "val1", Weight: 1},
				{Address: "val2", Weight: 1},
				{Address: "val3", Weight: 1},
				{Address: "val4", Weight: 1},
				{Address: "val5", Weight: 1},
				{Address: "val6", Weight: 1},
				{Address: "val7", Weight: 1},
				{Address: "val8", Weight: 1},
				{Address: "val9", Weight: 1},
				{Address: "val10", Weight: 1},
			},
			exceedsCap: false,
		},
		{
			name:      "10pct splits exceeds threshold",
			weightCap: 9,
			validators: []*types.Validator{
				{Address: "val1", Weight: 1},
				{Address: "val2", Weight: 1},
				{Address: "val3", Weight: 1},
				{Address: "val4", Weight: 1},
				{Address: "val5", Weight: 1},
				{Address: "val6", Weight: 1},
				{Address: "val7", Weight: 1},
				{Address: "val8", Weight: 1},
				{Address: "val9", Weight: 1},
				{Address: "val10", Weight: 1},
			},
			exceedsCap: true,
		},
		{
			name:      "One val exceeds cap",
			weightCap: 10,
			validators: []*types.Validator{
				{Address: "val1", Weight: 1},
				{Address: "val2", Weight: 1},
				{Address: "val3", Weight: 1},
				{Address: "val4", Weight: 1},
				{Address: "val5", Weight: 1},
				{Address: "val6", Weight: 1},
				{Address: "val7", Weight: 1},
				{Address: "val8", Weight: 2},
				{Address: "val9", Weight: 1},
				{Address: "val10", Weight: 1},
			},
			exceedsCap: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			params := s.App.StakeibcKeeper.GetParams(s.Ctx)
			params.ValidatorWeightCap = tc.weightCap
			s.App.StakeibcKeeper.SetParams(s.Ctx, params)

			hostZone := types.HostZone{
				ChainId:    HostChainId,
				Validators: tc.validators,
			}
			s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

			err := s.App.StakeibcKeeper.CheckValidatorWeightsBelowCap(s.Ctx, HostChainId)
			if !tc.exceedsCap {
				s.Require().NoError(err, "set should not have exceeded cap")
			} else {
				s.Require().Error(err, "set should have exceeded cap")
			}
		})
	}
}

// TODO [cleanup]: Remove after v17 upgrade
func (s *KeeperTestSuite) TestDisableHubTokenization() {
	chainId := "cosmoshub-4"

	// Create the host zone and delegation channel
	owner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_DELEGATION)
	channelId, portId := s.CreateICAChannel(owner)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:      chainId,
		ConnectionId: ibctesting.FirstConnectionID,
	})

	// Call the disable function and confirm the sequence number incremented (indicating an ICA was submitted)
	s.CheckICATxSubmitted(portId, channelId, func() error {
		s.App.StakeibcKeeper.DisableHubTokenization(s.Ctx)
		return nil
	})
}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone() {
	validators := []*types.Validator{
		{Address: "val1", Weight: 20},
		{Address: "val2", Weight: 40},
		{Address: "val3", Weight: 30},
		{Address: "val6", Weight: 5},
		{Address: "val5", Weight: 0},
		{Address: "val4", Weight: 5},
	}

	// Get targets with an even 100 total delegated - no overflow to last validator
	totalDelegation := sdkmath.NewInt(100)
	hostZone := types.HostZone{ChainId: HostChainId, Validators: validators}
	actualTargets, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, totalDelegation)
	s.Require().NoError(err, "no error expected when getting target weights for total delegation of 100")

	// Confirm target - should equal the validator's weight
	for _, validator := range validators {
		s.Require().Equal(int64(validator.Weight), actualTargets[validator.Address].Int64(),
			"validator %s target for total delegation of 100", validator.Address)
	}

	// Get targets with an uneven amount delegated - 77 - over flow to last validator
	totalDelegation = sdkmath.NewInt(77)
	expectedTargets := map[string]int64{
		"val5": 0,  // 0%  of 77 = 0
		"val4": 3,  // 5%  of 77 = 3.85 -> 3
		"val6": 3,  // 5%  of 77 = 3.85 -> 3
		"val1": 15, // 20% of 77 = 15.4 -> 15
		"val3": 23, // 30% of 77 = 23.1 -> 23
		"val2": 33, // Gets all overflow: 77 - 3 - 3 - 15 - 23 = 33
	}
	actualTargets, err = s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, totalDelegation)
	s.Require().NoError(err, "no error expected when getting target weights for total delegation of 77")

	// Confirm target amounts again
	for validatorAddress, expectedTarget := range expectedTargets {
		s.Require().Equal(expectedTarget, actualTargets[validatorAddress].Int64(),
			"validator %s target for total delegation of 77", validatorAddress)
	}

	// Check zero delegations throws an error
	_, err = s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, sdkmath.ZeroInt())
	s.Require().ErrorContains(err, "Cannot calculate target delegation if final amount is less than or equal to zero")

	// Check zero weights throws an error
	_, err = s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, types.HostZone{}, sdkmath.NewInt(1))
	s.Require().ErrorContains(err, "No non-zero validators found for host zone")
}

func (s *KeeperTestSuite) TestEnableRedemptions() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:            HostChainId,
		RedemptionsEnabled: false,
	})

	err := s.App.StakeibcKeeper.EnableRedemptions(s.Ctx, HostChainId)
	s.Require().NoError(err)

	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().True(hostZone.RedemptionsEnabled, "redemptions should have been enabled")
}
