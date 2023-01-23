package keeper_test

import (
	"context"
	"fmt"
	"time"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v5/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"

	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

// Add three rate limits on different channels
// Each should have a different chainId
func (s *KeeperTestSuite) setupQueryRateLimitTests() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := int64(0); i <= 2; i++ {
		clientId := fmt.Sprintf("07-tendermint-%d", i)
		chainId := fmt.Sprintf("chain-%d", i)
		connectionId := fmt.Sprintf("connection-%d", i)
		channelId := fmt.Sprintf("channel-%d", i)

		// First register the client, connection, and channel (so we can map back to chainId)
		// Nothing in the client state matters besides the chainId
		clientState := ibctmtypes.NewClientState(
			chainId, ibctmtypes.Fraction{}, time.Duration(0), time.Duration(0), time.Duration(0), clienttypes.Height{}, nil, nil, true, true,
		)
		connection := connectiontypes.ConnectionEnd{ClientId: clientId}
		channel := channeltypes.Channel{ConnectionHops: []string{connectionId}}

		s.App.IBCKeeper.ClientKeeper.SetClientState(s.Ctx, clientId, clientState)
		s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, connectionId, connection)
		s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, channelId, channel)

		// Then add the rate limit
		rateLimit := types.RateLimit{
			Path: &types.Path{Denom: "denom", ChannelId: channelId},
		}
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit)
		rateLimits = append(rateLimits, rateLimit)
	}
	return rateLimits
}

func (s *KeeperTestSuite) TestQueryAllRateLimits() {
	expectedRateLimits := s.setupQueryRateLimitTests()
	queryResponse, err := s.QueryClient.AllRateLimits(context.Background(), &types.QueryAllRateLimitsRequest{})
	s.Require().NoError(err)
	s.Require().ElementsMatch(expectedRateLimits, queryResponse.RateLimits)
}

func (s *KeeperTestSuite) TestQueryRateLimit() {
	allRateLimits := s.setupQueryRateLimitTests()
	for _, expectedRateLimit := range allRateLimits {
		queryResponse, err := s.QueryClient.RateLimit(context.Background(), &types.QueryRateLimitRequest{
			Denom:     expectedRateLimit.Path.Denom,
			ChannelId: expectedRateLimit.Path.ChannelId,
		})
		s.Require().NoError(err, "no error expected when querying rate limit on channel: %s", expectedRateLimit.Path.ChannelId)
		s.Require().Equal(expectedRateLimit, *queryResponse.RateLimit)
	}
}

func (s *KeeperTestSuite) TestQueryRateLimitsByChainId() {
	allRateLimits := s.setupQueryRateLimitTests()
	for i, expectedRateLimit := range allRateLimits {
		chainId := fmt.Sprintf("chain-%d", i)
		queryResponse, err := s.QueryClient.RateLimitsByChainId(context.Background(), &types.QueryRateLimitsByChainIdRequest{
			ChainId: chainId,
		})
		s.Require().NoError(err, "no error expected when querying rate limit on chain: %s", chainId)
		s.Require().Len(queryResponse.RateLimits, 1)
		s.Require().Equal(expectedRateLimit, queryResponse.RateLimits[0])
	}
}

func (s *KeeperTestSuite) TestQueryRateLimitsByChannelId() {
	allRateLimits := s.setupQueryRateLimitTests()
	for i, expectedRateLimit := range allRateLimits {
		channelId := fmt.Sprintf("channel-%d", i)
		queryResponse, err := s.QueryClient.RateLimitsByChannelId(context.Background(), &types.QueryRateLimitsByChannelIdRequest{
			ChannelId: channelId,
		})
		s.Require().NoError(err, "no error expected when querying rate limit on channel: %s", channelId)
		s.Require().Len(queryResponse.RateLimits, 1)
		s.Require().Equal(expectedRateLimit, queryResponse.RateLimits[0])
	}
}
