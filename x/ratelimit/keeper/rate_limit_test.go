package keeper_test

import (
	"strconv"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func (s *KeeperTestSuite) createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		rateLimit := types.RateLimit{
			Path: &types.Path{Denom: "denom-" + suffix, ChannelId: "channel-" + suffix},
		}

		rateLimits = append(rateLimits, rateLimit)
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit)
	}
	return rateLimits
}

func (s *KeeperTestSuite) TestGetRateLimit() {
	rateLimits := s.createRateLimits()

	expectedRateLimit := rateLimits[0]
	denom := expectedRateLimit.Path.Denom
	channelId := expectedRateLimit.Path.ChannelId

	actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found, "element should have been found, but was not")
	s.Require().Equal(expectedRateLimit, actualRateLimit)
}

func (s *KeeperTestSuite) TestRemoveRateLimit() {
	rateLimits := s.createRateLimits()

	rateLimitToRemove := rateLimits[0]
	denomToRemove := rateLimitToRemove.Path.Denom
	channelIdToRemove := rateLimitToRemove.Path.ChannelId

	s.App.RatelimitKeeper.RemoveRateLimit(s.Ctx, denomToRemove, channelIdToRemove)
	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denomToRemove, channelIdToRemove)
	s.Require().False(found, "the removed element should not have been found, but it was")
}

func (s *KeeperTestSuite) TestGetAllRateLimits() {
	expectedRateLimits := s.createRateLimits()
	actualRateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
	s.Require().Len(actualRateLimits, len(expectedRateLimits))
	s.Require().ElementsMatch(expectedRateLimits, actualRateLimits, "all rate limits")
}
