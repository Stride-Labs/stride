package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	"github.com/Stride-Labs/stride/v28/x/claim/types"
	epochtypes "github.com/Stride-Labs/stride/v28/x/epochs/types"
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
	err := s.App.ClaimKeeper.SetParams(s.Ctx, types.Params{
		Airdrops: []*types.Airdrop{
			{
				AirdropIdentifier: airdropEndedId,
				ClaimedSoFar:      claimedSoFar,
				AirdropStartTime:  s.Ctx.BlockTime().Add(-3 * time.Minute), // started 3 minutes ago
			},
			{
				AirdropIdentifier: airdropInProgressId,
				ClaimedSoFar:      claimedSoFar,
				AirdropStartTime:  s.Ctx.BlockTime().Add(-1 * time.Minute), // started 1 minute ago
			},
		},
	})
	s.Require().NoError(err, "no error expected when setting claims params")

	// Add the corresponding epoch for each airdrop
	epochEnded := epochtypes.EpochInfo{Identifier: epochEndedId}
	epochInProgress := epochtypes.EpochInfo{Identifier: epochInProgressId}
	s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epochEnded)
	s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epochInProgress)

	// Add claim records for each airdrop
	actions := [][]bool{
		{false, false, false},
		{true, false, true},
		{true, true, true},
	}
	addressToAction := map[string][]bool{}
	for i, action := range actions {
		address := addresses[i].String()

		err := s.App.ClaimKeeper.SetClaimRecord(s.Ctx, types.ClaimRecord{
			AirdropIdentifier: airdropEndedId,
			Address:           address,
			ActionCompleted:   action,
		})
		s.Require().NoError(err, "no error expected when setting claims record for airdrop-ended, claim %d", i)

		err = s.App.ClaimKeeper.SetClaimRecord(s.Ctx, types.ClaimRecord{
			AirdropIdentifier: airdropInProgressId,
			Address:           address,
			ActionCompleted:   action,
		})
		s.Require().NoError(err, "no error expected when setting claims record for airdrop-in-progress, claim %d", i)
		addressToAction[address] = action
	}

	// Call AfterEpochEnds with each epoch
	s.App.ClaimKeeper.AfterEpochEnd(s.Ctx, epochEnded)
	s.App.ClaimKeeper.AfterEpochEnd(s.Ctx, epochInProgress)

	// Check that the airdrop that ended had everything reset and the actions were reset
	airdropEnded := s.App.ClaimKeeper.GetAirdropByIdentifier(s.Ctx, airdropEndedId)
	s.Require().Equal(int64(0), airdropEnded.ClaimedSoFar.Int64(), "claimed so far for airdrop that ended")

	actionsReset := []bool{false, false, false}
	endedClaimRecords := s.App.ClaimKeeper.GetClaimRecords(s.Ctx, airdropEndedId)
	s.Require().Len(endedClaimRecords, 3)

	for i, claimRecord := range endedClaimRecords {
		s.Require().Equal(actionsReset, claimRecord.ActionCompleted, "actions for claim record %d, for airdrop %s", i, airdropEndedId)
	}

	// And check that the airdrop that was still in progress has been unchanged
	airdropInProgress := s.App.ClaimKeeper.GetAirdropByIdentifier(s.Ctx, airdropInProgressId)
	s.Require().Equal(claimedSoFar.Int64(), airdropInProgress.ClaimedSoFar.Int64(), "claimed so far for airdrop in progress")

	inProgressClaimRecords := s.App.ClaimKeeper.GetClaimRecords(s.Ctx, airdropInProgressId)
	s.Require().Len(inProgressClaimRecords, 3)

	for i, claimRecord := range inProgressClaimRecords {
		s.Require().Equal(addressToAction[claimRecord.Address], claimRecord.ActionCompleted,
			"actions for claim record %d, for airdrop %s", i, airdropInProgressId)
	}
}
