package v35_test

import (
	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	stakeibckeeper "github.com/Stride-Labs/stride/v35/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// Seeds the haqq host zone with every validator in the delta table, each tracked at the value
// the table was measured against (HaqqExpectedTrackedDelegations). The full table is needed
// because the reconciliation refuses to apply a partial one.
func (s *UpgradeTestSuite) setupHaqqHostZone() (tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int) {
	tracked = map[string]sdkmath.Int{}
	trackedTotal = sdkmath.ZeroInt()

	validators := []*stakeibctypes.Validator{}
	for _, entry := range v35.HaqqDelegationDeltas {
		delegation := v35.HaqqExpectedTrackedDelegations[entry.Address]
		validators = append(validators, &stakeibctypes.Validator{
			Name:       entry.Name,
			Address:    entry.Address,
			Delegation: delegation,
		})
		tracked[entry.Address] = delegation
		trackedTotal = trackedTotal.Add(delegation)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, withInBoundsRates(stakeibctypes.HostZone{
		ChainId:          v35.HaqqChainId,
		HostDenom:        "aISLM",
		TotalDelegations: trackedTotal,
		Validators:       validators,
	}))
	return tracked, trackedTotal
}

func (s *UpgradeTestSuite) TestHaqqDelegationDeltas_TableShape() {
	seen := map[string]bool{}
	hasPositive, hasNegative := false, false
	net := sdkmath.ZeroInt()
	for _, entry := range v35.HaqqDelegationDeltas {
		s.Require().False(seen[entry.Address], "duplicate address %s", entry.Address)
		seen[entry.Address] = true
		s.Require().False(entry.Delta.IsZero(), "%s has a zero delta", entry.Name)
		s.Require().Contains(entry.Address, "haqqvaloper1", "%s is not a haqq operator address", entry.Address)
		hasPositive = hasPositive || entry.Delta.IsPositive()
		hasNegative = hasNegative || entry.Delta.IsNegative()
		net = net.Add(entry.Delta)
	}
	s.Require().True(hasPositive && hasNegative, "both signs are applied (spec §5)")
	s.Require().True(net.IsNegative(), "the 2026-09-29 table nets to an over-recording, so TotalDelegations must drop")
}

func (s *UpgradeTestSuite) TestReconcileHaqqDelegations() {
	tracked, trackedTotal := s.setupHaqqHostZone()

	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().True(applied)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().True(found)

	expectedDelta := sdkmath.ZeroInt()
	for _, entry := range v35.HaqqDelegationDeltas {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		s.Require().True(found, "validator %s should still exist", entry.Name)
		s.Require().Equal(tracked[entry.Address].Add(entry.Delta), validator.Delegation, "%s delegation moves by its delta", entry.Name)
		expectedDelta = expectedDelta.Add(entry.Delta)
	}
	s.Require().Equal(expectedDelta, appliedDelta, "returned delta is the sum over the whole table")
	s.Require().Equal(trackedTotal.Add(appliedDelta), hostZone.TotalDelegations, "TotalDelegations adjusted by exactly the net")
	s.Require().True(hostZone.TotalDelegations.LT(trackedTotal), "TotalDelegations drops")

	sum := sdkmath.ZeroInt()
	for _, validator := range hostZone.Validators {
		sum = sum.Add(validator.Delegation)
	}
	s.Require().Equal(sum, hostZone.TotalDelegations, "TotalDelegations == sum(validator.Delegation)")
}

// Dropping one entry from the host zone must skip the whole reconciliation
func (s *UpgradeTestSuite) TestReconcileHaqqDelegations_MissingValidatorSkipsAll() {
	tracked, trackedTotal := s.setupHaqqHostZone()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	hostZone.Validators = hostZone.Validators[1:]
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().False(applied)
	s.Require().True(appliedDelta.IsZero())

	after, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().Equal(trackedTotal, after.TotalDelegations, "nothing written")
	for _, validator := range after.Validators {
		s.Require().Equal(tracked[validator.Address], validator.Delegation, "%s untouched", validator.Name)
	}
}

func (s *UpgradeTestSuite) TestReconcileHaqqDelegations_MissingZone() {
	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().False(applied)
	s.Require().True(appliedDelta.IsZero())
}

func (s *UpgradeTestSuite) TestHaqqExpectedTrackedDelegations_MatchesTable() {
	s.Require().Len(v35.HaqqExpectedTrackedDelegations, len(v35.HaqqDelegationDeltas))
	for _, entry := range v35.HaqqDelegationDeltas {
		expected, found := v35.HaqqExpectedTrackedDelegations[entry.Address]
		s.Require().True(found, "%s has an expected tracked delegation", entry.Name)
		s.Require().False(expected.IsNil() || expected.IsNegative(), "%s expected value is a non-negative int", entry.Name)
	}
}

// A slash booked between table generation and the upgrade changes that row's tracked delegation;
// applying its delta on top would double-apply the slash, so that row is skipped and every other row
// still applies
func (s *UpgradeTestSuite) TestReconcileHaqqDelegations_TrackedMismatchSkipsOnlyThatRow() {
	tracked, trackedTotal := s.setupHaqqHostZone()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	hostZone.Validators[0].Delegation = hostZone.Validators[0].Delegation.SubRaw(1)
	hostZone.TotalDelegations = hostZone.TotalDelegations.SubRaw(1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	mismatched := hostZone.Validators[0]

	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().True(applied)

	expectedDelta := sdkmath.ZeroInt()
	after, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, entry := range v35.HaqqDelegationDeltas {
		validator, _, _ := stakeibckeeper.GetValidatorFromAddress(after.Validators, entry.Address)
		if entry.Address == mismatched.Address {
			s.Require().Equal(tracked[entry.Address].SubRaw(1), validator.Delegation, "%s untouched", entry.Name)
			continue
		}
		s.Require().Equal(tracked[entry.Address].Add(entry.Delta), validator.Delegation, "%s moves by its delta", entry.Name)
		expectedDelta = expectedDelta.Add(entry.Delta)
	}
	s.Require().Equal(expectedDelta, appliedDelta, "returned delta excludes the skipped row")
	s.Require().Equal(trackedTotal.SubRaw(1).Add(expectedDelta), after.TotalDelegations, "TotalDelegations moves by the applied rows only")
}

// When every row's tracked delegation has moved, nothing is applied
func (s *UpgradeTestSuite) TestReconcileHaqqDelegations_AllRowsMismatchedAppliesNothing() {
	tracked, trackedTotal := s.setupHaqqHostZone()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, validator := range hostZone.Validators {
		validator.Delegation = validator.Delegation.AddRaw(1)
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	appliedDelta, applied := v35.ReconcileHaqqDelegations(s.Ctx, s.App.StakeibcKeeper)
	s.Require().False(applied)
	s.Require().True(appliedDelta.IsZero())

	after, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().Equal(trackedTotal, after.TotalDelegations, "nothing written")
	for _, validator := range after.Validators {
		s.Require().Equal(tracked[validator.Address].AddRaw(1), validator.Delegation, "%s untouched", validator.Name)
	}
}
