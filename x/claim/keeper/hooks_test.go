package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	"github.com/Stride-Labs/stride/v9/x/claim/types"
	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
)

func (s *KeeperTestSuite) TestAfterEpochEnd() {
	addresses := apptesting.CreateRandomAccounts(3)

	airdropEndedId := "ended"
	airdropInProgressId := "in-progress"

	epochEndedId := "airdrop-" + airdropEndedId
	epochInProgressId := "airdrop-" + airdropInProgressId

	claimedSoFar := sdkmath.NewInt(1000)

	// Add two airdrops - one that ended, and one that's in progress
	types.DefaultVestingInitialPeriod = time.Minute * 2 // vesting period of 2 minutes
	err := s.app.ClaimKeeper.SetParams(s.ctx, types.Params{
		Airdrops: []*types.Airdrop{
			{
				AirdropIdentifier: airdropEndedId,
				ClaimedSoFar:      claimedSoFar,
				AirdropStartTime:  s.ctx.BlockTime().Add(-3 * time.Minute), // started 3 minutes ago
			},
			{
				AirdropIdentifier: airdropInProgressId,
				ClaimedSoFar:      claimedSoFar,
				AirdropStartTime:  s.ctx.BlockTime().Add(-1 * time.Minute), // started 1 minute ago
			},
		},
	})
	s.Require().NoError(err, "no error expected when setting claims params")

	// Add the corresponding epoch for each airdrop
	epochEnded := epochtypes.EpochInfo{Identifier: epochEndedId}
	epochInProgress := epochtypes.EpochInfo{Identifier: epochInProgressId}
	s.app.EpochsKeeper.SetEpochInfo(s.ctx, epochEnded)
	s.app.EpochsKeeper.SetEpochInfo(s.ctx, epochInProgress)

	// Add claim records for each airdrop
	actions := [][]bool{
		{false, false, false},
		{true, false, true},
		{true, true, true},
	}
	addressToAction := map[string][]bool{}
	for i, action := range actions {
		address := addresses[i].String()

		err := s.app.ClaimKeeper.SetClaimRecord(s.ctx, types.ClaimRecord{
			AirdropIdentifier: airdropEndedId,
			Address:           address,
			ActionCompleted:   action,
		})
		s.Require().NoError(err, "no error expected when setting claims record for airdrop-ended, claim %d", i)

		err = s.app.ClaimKeeper.SetClaimRecord(s.ctx, types.ClaimRecord{
			AirdropIdentifier: airdropInProgressId,
			Address:           address,
			ActionCompleted:   action,
		})
		s.Require().NoError(err, "no error expected when setting claims record for airdrop-in-progress, claim %d", i)
		addressToAction[address] = action
	}

	// Call AfterEpochEnds with each epoch
	s.app.ClaimKeeper.AfterEpochEnd(s.ctx, epochEnded)
	s.app.ClaimKeeper.AfterEpochEnd(s.ctx, epochInProgress)

	// Check that the airdrop that ended had everything reset and the actions were reset
	airdropEnded := s.app.ClaimKeeper.GetAirdropByIdentifier(s.ctx, airdropEndedId)
	s.Require().Equal(int64(0), airdropEnded.ClaimedSoFar.Int64(), "claimed so far for airdrop that ended")

	actionsReset := []bool{false, false, false}
	endedClaimRecords := s.app.ClaimKeeper.GetClaimRecords(s.ctx, airdropEndedId)
	s.Require().Len(endedClaimRecords, 3)

	for i, claimRecord := range endedClaimRecords {
		s.Require().Equal(actionsReset, claimRecord.ActionCompleted, "actions for claim record %d, for airdrop %s", i, airdropEndedId)
	}

	// And check that the airdrop that was still in progress has been unchanged
	airdropInProgress := s.app.ClaimKeeper.GetAirdropByIdentifier(s.ctx, airdropInProgressId)
	s.Require().Equal(claimedSoFar.Int64(), airdropInProgress.ClaimedSoFar.Int64(), "claimed so far for airdrop in progress")

	inProgressClaimRecords := s.app.ClaimKeeper.GetClaimRecords(s.ctx, airdropInProgressId)
	s.Require().Len(inProgressClaimRecords, 3)

	for i, claimRecord := range inProgressClaimRecords {
		s.Require().Equal(addressToAction[claimRecord.Address], claimRecord.ActionCompleted,
			"actions for claim record %d, for airdrop %s", i, airdropInProgressId)
	}
}
