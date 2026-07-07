package v33_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v32/app/apptesting"
	v33 "github.com/Stride-Labs/stride/v32/app/upgrades/v33"
	epochstypes "github.com/Stride-Labs/stride/v32/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v32/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v32/x/stakeibc/types"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// Tracked (pre-migration) delegations for the three phantom validators; each is > its on-chain target
var (
	cosmostation = v33.OsmosisPhantomDelegations[0].Address
	chorusOne    = v33.OsmosisPhantomDelegations[1].Address
	pryzm        = v33.OsmosisPhantomDelegations[2].Address

	trackedCosmostation = sdkmath.NewInt(258_638_300000)
	trackedChorusOne    = sdkmath.NewInt(231_057_000000)
	trackedPryzm        = sdkmath.NewInt(185_390_600000)
)

func (s *UpgradeTestSuite) setupOsmosisPhantomState() (trackedTotal, expectedReduction sdkmath.Int) {
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochstypes.STRIDE_EPOCH,
		EpochNumber:     42,
	})

	trackedTotal = trackedCosmostation.Add(trackedChorusOne).Add(trackedPryzm)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          v33.OsmosisChainId,
		HostDenom:        "uosmo",
		TotalDelegations: trackedTotal,
		Validators: []*stakeibctypes.Validator{
			{Address: cosmostation, Delegation: trackedCosmostation},
			{Address: chorusOne, Delegation: trackedChorusOne},
			{Address: pryzm, Delegation: trackedPryzm},
		},
	})

	// reduction = tracked - on-chain target, summed
	expectedReduction = trackedCosmostation.
		Add(trackedChorusOne).
		Add(trackedPryzm.Sub(v33.OsmosisPhantomDelegations[2].OnChainDelegation))
	return trackedTotal, expectedReduction
}

func (s *UpgradeTestSuite) osmoDelegationQueueRecords() []recordstypes.DepositRecord {
	var out []recordstypes.DepositRecord
	for _, r := range s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx) {
		if r.HostZoneId == v33.OsmosisChainId && r.Status == recordstypes.DepositRecord_DELEGATION_QUEUE {
			out = append(out, r)
		}
	}
	return out
}

func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations() {
	trackedTotal, expectedReduction := s.setupOsmosisPhantomState()

	err := v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().NoError(err, "reconciliation should succeed")

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v33.OsmosisChainId)
	s.Require().True(found)

	// Each phantom validator's Delegation is set to its on-chain target
	delegations := map[string]sdkmath.Int{}
	for _, v := range hostZone.Validators {
		delegations[v.Address] = v.Delegation
	}
	s.Require().Equal(sdkmath.ZeroInt(), delegations[cosmostation], "cosmostation reconciled to on-chain 0")
	s.Require().Equal(sdkmath.ZeroInt(), delegations[chorusOne], "chorus_one reconciled to on-chain 0")
	s.Require().Equal(v33.OsmosisPhantomDelegations[2].OnChainDelegation, delegations[pryzm], "pryzmstakedrop reconciled to on-chain")

	// TotalDelegations reduced by exactly the phantom
	s.Require().Equal(trackedTotal.Sub(expectedReduction), hostZone.TotalDelegations, "TotalDelegations reduced by phantom")

	// A single DELEGATION_QUEUE deposit record credits the reconciled amount
	records := s.osmoDelegationQueueRecords()
	s.Require().Len(records, 1, "exactly one deposit record should be created")
	s.Require().Equal(expectedReduction, records[0].Amount, "deposit record credits the reconciled amount")
	s.Require().Equal("uosmo", records[0].Denom, "deposit record denom")

	// Rate-preservation invariant: the credit exactly offsets the TotalDelegations reduction
	s.Require().Equal(trackedTotal.Sub(hostZone.TotalDelegations), records[0].Amount,
		"credit must exactly offset the TotalDelegations reduction (rate preserved)")

	// Internal consistency: TotalDelegations == sum of validator delegations
	sum := sdkmath.ZeroInt()
	for _, v := range hostZone.Validators {
		sum = sum.Add(v.Delegation)
	}
	s.Require().Equal(sum, hostZone.TotalDelegations, "TotalDelegations == sum(validator.Delegation)")
}

func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations_MissingHostZone() {
	// Without the osmosis-1 host zone (e.g. dockernet or a testnet), reconciliation must
	// no-op instead of erroring - an error would fail the upgrade and halt the chain
	err := v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper)
	s.Require().NoError(err, "missing host zone should be skipped, not an error")
	s.Require().Empty(s.osmoDelegationQueueRecords(), "no deposit record should be created")
}

func (s *UpgradeTestSuite) TestReconcileOsmosisDelegations_Idempotent() {
	s.setupOsmosisPhantomState()

	s.Require().NoError(v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper))
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v33.OsmosisChainId)
	firstTotal := hostZone.TotalDelegations

	// Running again must be a no-op: records already match on-chain, so no reduction and no new credit
	s.Require().NoError(v33.ReconcileOsmosisDelegations(s.Ctx, s.App.StakeibcKeeper, s.App.RecordsKeeper))
	hostZoneAfter, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v33.OsmosisChainId)

	s.Require().Equal(firstTotal, hostZoneAfter.TotalDelegations, "second run must not change TotalDelegations")
	s.Require().Len(s.osmoDelegationQueueRecords(), 1, "second run must not create another deposit record")
}
