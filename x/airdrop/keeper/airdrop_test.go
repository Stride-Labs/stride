package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v26/x/airdrop/types"
)

func (s *KeeperTestSuite) addAirdrops() (airdrops []types.Airdrop) {
	for i := 0; i <= 4; i++ {
		airdrop := types.Airdrop{
			Id:                fmt.Sprintf("airdrop-%d", i),
			EarlyClaimPenalty: sdkmath.LegacyZeroDec(),
		}
		airdrops = append(airdrops, airdrop)
		s.App.AirdropKeeper.SetAirdrop(s.Ctx, airdrop)
	}
	return airdrops
}

func (s *KeeperTestSuite) TestGetAirdrop() {
	airdrops := s.addAirdrops()

	for i := 0; i < len(airdrops); i++ {
		expectedAirdrop := airdrops[i]
		actualAirdrop, found := s.App.AirdropKeeper.GetAirdrop(s.Ctx, expectedAirdrop.Id)
		s.Require().True(found, "airdrop %s should have been found", expectedAirdrop.Id)
		s.Require().Equal(expectedAirdrop, actualAirdrop, "airdrop %s", expectedAirdrop.Id)
	}
}

func (s *KeeperTestSuite) TestRemoveAirdrop() {
	airdrops := s.addAirdrops()

	for removedIndex := 0; removedIndex < len(airdrops); removedIndex++ {
		// Remove from removed index
		removedId := airdrops[removedIndex].Id
		s.App.AirdropKeeper.RemoveAirdrop(s.Ctx, removedId)

		// Confirm removed
		_, found := s.App.AirdropKeeper.GetAirdrop(s.Ctx, removedId)
		s.Require().False(found, "airdrop %d should have been removed", removedId)

		// Check all other airdrops are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(airdrops); checkedIndex++ {
			checkedId := airdrops[checkedIndex].Id
			_, found := s.App.AirdropKeeper.GetAirdrop(s.Ctx, checkedId)
			s.Require().True(found, "airdrop %d should still exist after removal of %s",
				checkedId, removedId)
		}
	}
}

func (s *KeeperTestSuite) TestGetAllAirdrops() {
	expectedAirdrops := s.addAirdrops()
	actualAirdrops := s.App.AirdropKeeper.GetAllAirdrops(s.Ctx)
	s.Require().Equal(len(expectedAirdrops), len(actualAirdrops), "number of airdrops")
	s.Require().ElementsMatch(expectedAirdrops, actualAirdrops)
}
