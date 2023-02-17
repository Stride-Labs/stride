package keeper_test

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func (s *KeeperTestSuite) TestQueryOracle() {
	allOracles := s.createOracles()
	for _, expectedOracle := range allOracles {
		queryResponse, err := s.QueryClient.Oracle(context.Background(), &types.QueryOracleRequest{
			ChainId: expectedOracle.ChainId,
		})
		s.Require().NoError(err, "no error expected when querying oracle %s", expectedOracle.ChainId)
		s.Require().Equal(expectedOracle, *queryResponse.Oracle)
	}
}

func (s *KeeperTestSuite) TestQueryAllOracles() {
	expectedOracles := s.createOracles()
	queryResponse, err := s.QueryClient.AllOracles(context.Background(), &types.QueryAllOraclesRequest{})
	s.Require().NoError(err, "no error expected when querying all oracles")
	s.Require().ElementsMatch(expectedOracles, queryResponse.Oracles)
}

func (s *KeeperTestSuite) TestQueryActiveOracles() {
	// Add 6 oracles, alternating each oracle between active and inactive
	activeOracles := []types.Oracle{}
	inActiveOracles := []types.Oracle{}
	for i := 1; i <= 6; i++ {
		oracle := types.Oracle{ChainId: fmt.Sprintf("chain-%d", i)}
		if i%2 == 0 {
			oracle.Active = true
			activeOracles = append(activeOracles, oracle)
		} else {
			oracle.Active = false
			inActiveOracles = append(inActiveOracles, oracle)
		}
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)
	}

	// Query active oracles
	activeOraclesResponse, err := s.QueryClient.ActiveOracles(context.Background(), &types.QueryActiveOraclesRequest{
		Active: true,
	})
	s.Require().NoError(err, "no error expected when querying active oracles")
	s.Require().ElementsMatch(activeOracles, activeOraclesResponse.Oracles)

	// Query inactive oracles
	inActiveOraclesResponse, err := s.QueryClient.ActiveOracles(context.Background(), &types.QueryActiveOraclesRequest{
		Active: false,
	})
	s.Require().NoError(err, "no error expected when querying inactive oracles")
	s.Require().ElementsMatch(inActiveOracles, inActiveOraclesResponse.Oracles)
}

func (s *KeeperTestSuite) TestQueryAllPendingMetricUpdates() {
	expectedPendingMetrics := s.addPendingMetrics()
	queryResponse, err := s.QueryClient.AllPendingMetricUpdates(s.Ctx, &types.QueryAllPendingMetricUpdatesRequest{})
	s.Require().NoError(err, "no error expected when querying pending metric updates")
	s.Require().ElementsMatch(expectedPendingMetrics, queryResponse.PendingUpdates)
}

func (s *KeeperTestSuite) TestQueryPendingMetricUpdates() {
	filterMetricKey := "key-2"
	filterOracleChainId := "chain-2"

	// Add pending metrics across 2 keys and 2 oracles
	updatesByMetric := make(map[string][]types.PendingMetricUpdate)
	updatesByOracle := make(map[string][]types.PendingMetricUpdate)
	allPendingUpdates := []types.PendingMetricUpdate{
		{
			Metric: &types.Metric{Key: "key-1"}, OracleChainId: "chain-1",
		},
		{
			Metric: &types.Metric{Key: "key-2"}, OracleChainId: "chain-1",
		},
		{
			Metric: &types.Metric{Key: "key-1"}, OracleChainId: "chain-2",
		},
		{
			Metric: &types.Metric{Key: "key-2"}, OracleChainId: "chain-2",
		},
	}
	for _, pendingUpdate := range allPendingUpdates {
		key := pendingUpdate.Metric.Key
		chainId := pendingUpdate.OracleChainId

		updatesByMetric[key] = append(updatesByMetric[key], pendingUpdate)
		updatesByOracle[chainId] = append(updatesByOracle[chainId], pendingUpdate)

		s.App.ICAOracleKeeper.SetMetricUpdateInProgress(s.Ctx, pendingUpdate)
	}

	// First check with no filters
	expectedNoFilters := allPendingUpdates
	queryResponse, err := s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{})
	s.Require().NoError(err, "no error expected when querying pending metric updates with no filter")
	s.Require().ElementsMatch(expectedNoFilters, queryResponse.PendingUpdates)

	// Check with a filter on the metric (metric key == "key-2")
	queryResponse, err = s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{
		MetricKey: filterMetricKey,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with metric key filter")
	s.Require().ElementsMatch(updatesByMetric[filterMetricKey], queryResponse.PendingUpdates)

	// Check with a filter on the oracle (chain-id == "chain-2")
	queryResponse, err = s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{
		OracleChainId: filterOracleChainId,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with oracle filter")
	s.Require().ElementsMatch(updatesByOracle[filterOracleChainId], queryResponse.PendingUpdates)

	// Check with a filter on both the metric and oracle (metric key == "key2", chain-id == "chain-2")
	expectedMetricAndOracleFilter := []types.PendingMetricUpdate{{
		Metric:        &types.Metric{Key: filterMetricKey},
		OracleChainId: filterOracleChainId,
	}}
	queryResponse, err = s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{
		MetricKey:     filterMetricKey,
		OracleChainId: filterOracleChainId,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with metric key and oracle filter")
	s.Require().ElementsMatch(expectedMetricAndOracleFilter, queryResponse.PendingUpdates)
}
