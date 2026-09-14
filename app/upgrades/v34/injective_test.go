package v34_test

import (
	sdkmath "cosmossdk.io/math"

	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
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

	appliedDelta, err := v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().True(found)

	expectedDelta := sdkmath.ZeroInt()
	for _, entry := range v34.InjectiveDelegationDeltas[:4] {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().True(found, "validator %s should still exist", entry.Name)
		s.Require().Equal(tracked[entry.Address].Add(entry.Delta), validator.Delegation,
			"%s delegation should move by its delta", entry.Name)
		expectedDelta = expectedDelta.Add(entry.Delta)
	}
	s.Require().True(expectedDelta.IsPositive(), "test relies on the seeded deltas netting positive")
	s.Require().Equal(expectedDelta, appliedDelta, "returned delta is the sum over the seeded validators only")
	s.Require().Equal(trackedTotal.Add(appliedDelta), hostZone.TotalDelegations,
		"TotalDelegations adjusted by exactly the returned delta")

	// Validators in the constants but not on the host zone are skipped, not errors
	s.Require().Len(hostZone.Validators, 4, "no validators should be added")

	sum := sdkmath.ZeroInt()
	for _, v := range hostZone.Validators {
		sum = sum.Add(v.Delegation)
	}
	s.Require().Equal(sum, hostZone.TotalDelegations, "TotalDelegations == sum(validator.Delegation)")
}

func (s *UpgradeTestSuite) TestReconcileInjectiveDelegations_MissingHostZone() {
	appliedDelta, err := v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().NoError(err, "missing host zone should be skipped, not an error")
	s.Require().True(appliedDelta.IsZero(), "nothing applied without a host zone")

	_, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().False(found, "the reconciliation should not create the host zone")
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

	appliedDelta, err := v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().NoError(err)
	s.Require().True(appliedDelta.IsZero(), "the skipped validator's delta must be excluded from the returned total")

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().Equal(tooSmall, hostZone.Validators[0].Delegation, "validator must be left untouched")
	s.Require().Equal(tooSmall, hostZone.TotalDelegations, "TotalDelegations must be left untouched")
}
