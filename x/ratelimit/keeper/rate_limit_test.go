package keeper_test

import (
	"strconv"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func (s *KeeperTestSuite) createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := 1; i <= 5; i++ {
		rateLimit := types.RateLimit{
			Path: &types.Path{Id: strconv.Itoa(i)},
		}

		rateLimits = append(rateLimits, rateLimit)
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit)
	}
	return rateLimits
}

func (s *KeeperTestSuite) TestGetRateLimit() {
	rateLimits := s.createRateLimits()
	expectedRateLimit := rateLimits[0]

	actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, expectedRateLimit.Path.Id)
	s.Require().True(found, "element should have been found, but was not")
	s.Require().Equal(expectedRateLimit, actualRateLimit)
}

func (s *KeeperTestSuite) TestRemoveRateLimit() {
	rateLimits := s.createRateLimits()
	idToRemove := rateLimits[0].Path.Id

	s.App.RatelimitKeeper.RemoveRateLimit(s.Ctx, idToRemove)
	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, idToRemove)
	s.Require().False(found, "the removed element should not have been found, but it was")
}

func (s *KeeperTestSuite) TestGetAllRateLimits() {
	expectedRateLimits := s.createRateLimits()
	actualRateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
	s.Require().Len(actualRateLimits, len(expectedRateLimits))
	s.Require().ElementsMatch(expectedRateLimits, actualRateLimits, "all rate limits")
}
