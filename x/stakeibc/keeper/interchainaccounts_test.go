package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v26/x/stakeibc/keeper"
	types "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

// constant number of zero delegations
const numZeroDelegations = 37

func (s *KeeperTestSuite) ClaimAccruedStakingRewardsOnHost() {
	// Create a delegation ICA channel for the ICA submission
	owner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	channelId, portId := s.CreateICAChannel(owner)

	// Create validators
	validators := []*types.Validator{}
	numberGTClaimRewardsBatchSize := int(50)
	for i := 0; i < numberGTClaimRewardsBatchSize; i++ {

		// set most delegations to 5, some to 0
		valDelegation := math.NewInt(5)
		if i > (numberGTClaimRewardsBatchSize - numZeroDelegations) {
			valDelegation = math.NewInt(0)
		}
		validators = append(validators, &types.Validator{
			Address:    fmt.Sprintf("val-%d", i),
			Delegation: valDelegation,
		})
	}

	// Create host zone
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		DelegationIcaAddress: "delegation",
		WithdrawalIcaAddress: "withdrawal",
		Validators:           validators,
	}

	// Create epoch tracker for ICA timeout
	strideEpoch := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // used for timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpoch)

	// Get start sequence number to confirm ICA was set
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, portId, channelId)
	s.Require().True(found)

	// Call claim accrued rewards to submit ICAs
	err := s.App.StakeibcKeeper.ClaimAccruedStakingRewardsOnHost(s.Ctx, hostZone)
	s.Require().NoError(err, "no error expected when accruing rewards")

	// Confirm sequence number incremented by the number of txs
	// where the number of txs is equal:
	// (total_validators - validators_with_zero_delegation) / batch_size
	batchSize := (numberGTClaimRewardsBatchSize - numZeroDelegations) / stakeibckeeper.ClaimRewardsICABatchSize
	expectedEndSequence := startSequence + uint64(batchSize)
	actualEndSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, portId, channelId)
	s.Require().True(found)
	s.Require().Equal(expectedEndSequence, actualEndSequence, "sequence number should have incremented")

	// Attempt to call it with a host zone without a delegation ICA address, it should fail
	invalidHostZone := hostZone
	invalidHostZone.DelegationIcaAddress = ""
	err = s.App.StakeibcKeeper.ClaimAccruedStakingRewardsOnHost(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "ICA account not found")

	// Attempt to call it with a host zone without a withdrawal ICA address, it should fail
	invalidHostZone = hostZone
	invalidHostZone.WithdrawalIcaAddress = ""
	err = s.App.StakeibcKeeper.ClaimAccruedStakingRewardsOnHost(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "ICA account not found")

	// Attempt to call claim with an invalid connection ID on the host zone so the ica fails
	invalidHostZone = hostZone
	invalidHostZone.ConnectionId = ""
	err = s.App.StakeibcKeeper.ClaimAccruedStakingRewardsOnHost(s.Ctx, invalidHostZone)
	s.Require().ErrorContains(err, "Failed to SubmitTxs")
}
