package v34_test

import (
	sdkmath "cosmossdk.io/math"

	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Seeds the injective-1 host zone with a subset of the reconciled validators, each tracked at
// a round number so the expected post-reconciliation values are obvious.
func (s *UpgradeTestSuite) setupInjectiveHostZone() (tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int) {
	tracked = map[string]sdkmath.Int{}
	trackedTotal = sdkmath.ZeroInt()

	validators := []*stakeibctypes.Validator{}
	for i, entry := range v34.InjectiveDelegationDeltas[:4] {
		delegation := sdkmath.NewInt(int64(1000 + i)).Mul(sdkmath.NewInt(1e18))
		validators = append(validators, &stakeibctypes.Validator{
			Name:       entry.Name,
			Address:    entry.Address,
			Delegation: delegation,
		})
		tracked[entry.Address] = delegation
		trackedTotal = trackedTotal.Add(delegation)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v34.InjectiveChainId,
		HostDenom:        "inj",
		TotalDelegations: trackedTotal,
		Validators:       validators,
	})
	return tracked, trackedTotal
}

func (s *UpgradeTestSuite) TestReconcileInjectiveDelegations() {
	tracked, trackedTotal := s.setupInjectiveHostZone()

	s.Require().NoError(v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper))

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().True(found)

	expectedTotal := trackedTotal
	for _, entry := range v34.InjectiveDelegationDeltas[:4] {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().True(found, "validator %s should still exist", entry.Name)
		s.Require().Equal(tracked[entry.Address].Add(entry.Delta), validator.Delegation,
			"%s delegation should move by its delta", entry.Name)
		expectedTotal = expectedTotal.Add(entry.Delta)
	}
	s.Require().Equal(expectedTotal, hostZone.TotalDelegations, "TotalDelegations adjusted by the applied deltas only")

	// Validators in the constants but not on the host zone are skipped, not errors
	s.Require().Len(hostZone.Validators, 4, "no validators should be added")

	sum := sdkmath.ZeroInt()
	for _, v := range hostZone.Validators {
		sum = sum.Add(v.Delegation)
	}
	s.Require().Equal(sum, hostZone.TotalDelegations, "TotalDelegations == sum(validator.Delegation)")
}

func (s *UpgradeTestSuite) TestReconcileInjectiveDelegations_MissingHostZone() {
	s.Require().NoError(v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper),
		"missing host zone should be skipped, not an error")
}

// A negative delta larger than the tracked delegation means the constant is stale; that
// validator must be skipped (and excluded from the TotalDelegations adjustment) rather than
// driven negative or halting the upgrade.
func (s *UpgradeTestSuite) TestReconcileInjectiveDelegations_SkipsNegativeResult() {
	negative := v34.InjectiveDelegationDeltas[1] // blackpanther, delta ≈ -666 INJ
	s.Require().True(negative.Delta.IsNegative(), "test relies on a negative delta constant")

	tooSmall := sdkmath.NewInt(1e18)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v34.InjectiveChainId,
		HostDenom:        "inj",
		TotalDelegations: tooSmall,
		Validators:       []*stakeibctypes.Validator{{Address: negative.Address, Delegation: tooSmall}},
	})

	s.Require().NoError(v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper))

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().Equal(tooSmall, hostZone.Validators[0].Delegation, "validator must be left untouched")
	s.Require().Equal(tooSmall, hostZone.TotalDelegations, "TotalDelegations must be left untouched")
}

func (s *UpgradeTestSuite) setInjectiveUnbondingRecord(epoch uint64, status recordstypes.HostZoneUnbonding_Status, inProgress uint64) {
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordstypes.EpochUnbondingRecord{
		EpochNumber: epoch,
		HostZoneUnbondings: []*recordstypes.HostZoneUnbonding{{
			HostZoneId:                v34.InjectiveChainId,
			Denom:                     "inj",
			Status:                    status,
			StTokenAmount:             sdkmath.NewInt(100),
			NativeTokenAmount:         sdkmath.NewInt(150),
			NativeTokensToUnbond:      sdkmath.ZeroInt(),
			StTokensToBurn:            sdkmath.ZeroInt(),
			ClaimableNativeTokens:     sdkmath.ZeroInt(),
			UndelegationTxsInProgress: inProgress,
		}},
	})
}

func (s *UpgradeTestSuite) TestRequeueInjectiveUnbondings() {
	requeued := v34.RequeuedUnbondingEpochs
	s.Require().GreaterOrEqual(len(requeued), 3, "test expects at least three re-queued epochs")

	s.setInjectiveUnbondingRecord(requeued[0], recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, 0)       // waiting for sweep
	s.setInjectiveUnbondingRecord(requeued[1], recordstypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS, 0) // sweep ICA in flight
	s.setInjectiveUnbondingRecord(requeued[2], recordstypes.HostZoneUnbonding_CLAIMABLE, 0)                 // already swept

	s.Require().NoError(v34.RequeueInjectiveUnbondings(s.Ctx, s.App.RecordsKeeper))

	for _, epoch := range requeued[:2] {
		record, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epoch, v34.InjectiveChainId)
		s.Require().True(found)
		s.Require().Equal(recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE, record.Status, "epoch %d", epoch)
		s.Require().Equal(record.NativeTokenAmount, record.NativeTokensToUnbond, "full native amount is re-undelegated")
		s.Require().True(record.StTokensToBurn.IsZero(), "stTokens were already burned by the first undelegation")
		s.Require().True(record.ShouldRetryUnbonding(), "record must be picked up by the unbonding flow")
	}

	swept, _ := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, requeued[2], v34.InjectiveChainId)
	s.Require().Equal(recordstypes.HostZoneUnbonding_CLAIMABLE, swept.Status, "record in another status is untouched")
	s.Require().True(swept.NativeTokensToUnbond.IsZero())
}

func (s *UpgradeTestSuite) TestRequeueInjectiveUnbondings_MissingRecords() {
	s.Require().NoError(v34.RequeueInjectiveUnbondings(s.Ctx, s.App.RecordsKeeper),
		"missing records should be skipped, not an error")
}
