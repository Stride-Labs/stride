package v35_test

import (
	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Seeds the haqq host zone with every validator in the delta table, each tracked at a round
// number of whole ISLM so the expected post-reconciliation values are obvious. The full table
// is needed because the reconciliation refuses to apply a partial one.
func (s *UpgradeTestSuite) setupHaqqHostZone() (tracked map[string]sdkmath.Int, trackedTotal sdkmath.Int) {
	tracked = map[string]sdkmath.Int{}
	trackedTotal = sdkmath.ZeroInt()

	validators := []*stakeibctypes.Validator{}
	for i, entry := range v35.HaqqDelegationDeltas {
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
		ChainId:          v35.HaqqChainId,
		HostDenom:        "aISLM",
		TotalDelegations: trackedTotal,
		Validators:       validators,
	})
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
