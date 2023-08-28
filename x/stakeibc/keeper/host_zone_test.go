package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v14/testutil/keeper"
	"github.com/Stride-Labs/stride/v14/testutil/nullify"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func createNHostZone(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.HostZone {
	items := make([]types.HostZone, n)
	for i := range items {
		items[i].ChainId = strconv.Itoa(i)
		items[i].RedemptionRate = sdk.NewDec(1)
		items[i].LastRedemptionRate = sdk.NewDec(1)
		items[i].MinRedemptionRate = sdk.NewDecWithPrec(5, 1)
		items[i].MaxRedemptionRate = sdk.NewDecWithPrec(15, 1)
		items[i].TotalDelegations = sdkmath.ZeroInt()
		keeper.SetHostZone(ctx, items[i])
	}
	return items
}

func TestHostZoneGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNHostZone(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetHostZone(ctx, item.ChainId)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestHostZoneRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNHostZone(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveHostZone(ctx, item.ChainId)
		_, found := keeper.GetHostZone(ctx, item.ChainId)
		require.False(t, found)
	}
}

func TestHostZoneGetAll(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNHostZone(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllHostZone(ctx)),
	)
}

func TestHostZoneGetAllActiveCase1(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// Case 1: some active some inactive
	numZones := 3
	items := createNHostZone(keeper, ctx, numZones)
	// set the last host zone as halted
	items[numZones-1].Halted = true
	keeper.SetHostZone(ctx, items[numZones-1])

	// only the last host zone is active, so we expect all except that one
	actualActiveHzs := items[:numZones-1]
	getActiveHzResults := keeper.GetAllActiveHostZone(ctx)
	require.ElementsMatch(t,
		nullify.Fill(actualActiveHzs),
		nullify.Fill(getActiveHzResults),
	)
}

func TestHostZoneGetAllActiveCase2(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// Case 2: all active
	numZones := 3
	items := createNHostZone(keeper, ctx, numZones)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllActiveHostZone(ctx)),
	)
}

func TestHostZoneGetAllActiveCase3(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// Case 3: all inactive
	numZones := 3
	items := createNHostZone(keeper, ctx, numZones)
	// set the last host zone as halted
	items[0].Halted = true
	items[1].Halted = true
	items[2].Halted = true
	keeper.SetHostZone(ctx, items[0])
	keeper.SetHostZone(ctx, items[1])
	keeper.SetHostZone(ctx, items[2])
	require.ElementsMatch(t,
		nullify.Fill(types.HostZone{}),
		nullify.Fill(keeper.GetAllActiveHostZone(ctx)),
	)
}

func TestHostZoneGetAllActiveCase4(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)

	// create no zones, check the output is an empty list
	require.ElementsMatch(t,
		nullify.Fill(types.HostZone{}),
		nullify.Fill(keeper.GetAllActiveHostZone(ctx)),
	)
}

func TestGetValidatorFromAddress(t *testing.T) {
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

		require.True(t, found)
		require.Equal(t, expectedValidator, actualValidator)
		require.Equal(t, int64(expectedIndex), actualIndex)
	}

	// Test GetValidatorFromAddress for an validator that doesn't exist
	_, _, found := keeper.GetValidatorFromAddress(validators, "fake_validator")
	require.False(t, found)
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
