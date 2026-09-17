package v34_test

import (
	sdkmath "cosmossdk.io/math"

	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Seeds the injective-1 host zone with every validator in the delta table, each tracked at a
// round number so the expected post-reconciliation values are obvious. The full table is needed
// because the reconciliation refuses to apply a partial one.
func (s *UpgradeTestSuite) setupInjectiveHostZone() (tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int) {
	tracked = map[string]sdkmath.Int{}
	trackedTotal = sdkmath.ZeroInt()

	validators := []*stakeibctypes.Validator{}
	for i, entry := range v34.InjectiveDelegationDeltas {
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

	appliedDelta := v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().True(found)

	expectedDelta := sdkmath.ZeroInt()
	for _, entry := range v34.InjectiveDelegationDeltas {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().True(found, "validator %s should still exist", entry.Name)
		s.Require().Equal(tracked[entry.Address].Add(entry.Delta), validator.Delegation,
			"%s delegation should move by its delta", entry.Name)
		expectedDelta = expectedDelta.Add(entry.Delta)
	}
	s.Require().True(expectedDelta.IsPositive(), "test relies on the table netting positive")
	s.Require().Equal(expectedDelta, appliedDelta, "returned delta is the sum over the whole table")
	s.Require().Equal(trackedTotal.Add(appliedDelta), hostZone.TotalDelegations,
		"TotalDelegations adjusted by exactly the returned delta")
	s.Require().Len(hostZone.Validators, len(v34.InjectiveDelegationDeltas), "no validators should be added")

	sum := sdkmath.ZeroInt()
	for _, v := range hostZone.Validators {
		sum = sum.Add(v.Delegation)
	}
	s.Require().Equal(sum, hostZone.TotalDelegations, "TotalDelegations == sum(validator.Delegation)")
}

// Dropping one entry from the host zone must skip the whole reconciliation: applying the rest
// would return a partial sum, and that partial sum is what gets undelegated.
func (s *UpgradeTestSuite) TestReconcileInjectiveDelegations_MissingValidatorSkipsAll() {
	tracked, trackedTotal := s.setupInjectiveHostZone()

	// Remove the largest negative entry so a partial application would overshoot the true excess
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	removed := v34.InjectiveDelegationDeltas[1] // blackpanther, delta ≈ -666 INJ
	s.Require().True(removed.Delta.IsNegative(), "test relies on removing a negative delta")
	_, index, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, removed.Address)
	s.Require().True(found)
	hostZone.Validators = append(hostZone.Validators[:index], hostZone.Validators[index+1:]...)
	hostZone.TotalDelegations = trackedTotal.Sub(tracked[removed.Address])
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	appliedDelta := v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().True(appliedDelta.IsZero(), "nothing is applied when any entry is missing")

	hostZone, _ = s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().Equal(trackedTotal.Sub(tracked[removed.Address]), hostZone.TotalDelegations, "TotalDelegations must be left untouched")
	for _, validator := range hostZone.Validators {
		s.Require().Equal(tracked[validator.Address], validator.Delegation, "%s must be left untouched", validator.Name)
	}
}

func (s *UpgradeTestSuite) TestReconcileInjectiveDelegations_MissingHostZone() {
	appliedDelta := v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().True(appliedDelta.IsZero(), "nothing applied without a host zone")

	_, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().False(found, "the reconciliation should not create the host zone")
}

// A negative delta larger than the tracked delegation means the constant is stale; the whole
// reconciliation must be skipped with nothing written rather than drive the delegation negative
// or apply the other entries.
func (s *UpgradeTestSuite) TestReconcileInjectiveDelegations_NegativeResultSkipsAll() {
	tracked, trackedTotal := s.setupInjectiveHostZone()

	negative := v34.InjectiveDelegationDeltas[1] // blackpanther, delta ≈ -666 INJ
	s.Require().True(negative.Delta.IsNegative(), "test relies on a negative delta constant")

	// Shrink that validator's tracked delegation below the delta's magnitude
	tooSmall := sdkmath.NewInt(1e18)
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	_, index, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, negative.Address)
	s.Require().True(found)
	hostZone.Validators[index].Delegation = tooSmall
	hostZone.TotalDelegations = trackedTotal.Sub(tracked[negative.Address]).Add(tooSmall)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	expectedTotal := hostZone.TotalDelegations

	appliedDelta := v34.ReconcileInjectiveDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().True(appliedDelta.IsZero(), "nothing is applied when any entry would go negative")

	hostZone, _ = s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.InjectiveChainId)
	s.Require().Equal(expectedTotal, hostZone.TotalDelegations, "TotalDelegations must be left untouched")
	s.Require().Equal(tooSmall, hostZone.Validators[index].Delegation, "the failing validator must be left untouched")
	for _, validator := range hostZone.Validators {
		if validator.Address == negative.Address {
			continue
		}
		s.Require().Equal(tracked[validator.Address], validator.Delegation, "%s must be left untouched", validator.Name)
	}
}
