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

func (s *KeeperTestSuite) TestQueryAllPendingMetrics() {
	// Create queued metrics
	metrics := s.createMetrics()

	// Query pending metrics - should return 0 metrics
	queryResponse, err := s.QueryClient.AllPendingMetrics(s.Ctx, &types.QueryAllPendingMetricsRequest{})
	s.Require().NoError(err, "no error expected when querying pending metrics")
	s.Require().Empty(queryResponse.Metrics, "no pending metrics at start")

	// Update the first 3 metrics to in progress
	expectedMetrics := []types.Metric{}
	for _, metric := range metrics[:3] {
		s.App.ICAOracleKeeper.UpdateMetricStatus(s.Ctx, metric, types.MetricStatus_METRIC_STATUS_IN_PROGRESS)
		metric.Status = types.MetricStatus_METRIC_STATUS_IN_PROGRESS
		expectedMetrics = append(expectedMetrics, metric)
	}

	// Query metrics again - should return the first 3 metrics
	queryResponse, err = s.QueryClient.AllPendingMetrics(s.Ctx, &types.QueryAllPendingMetricsRequest{})
	s.Require().NoError(err, "no error expected when querying pending metrics")
	s.Require().ElementsMatch(expectedMetrics, queryResponse.Metrics)
}

func (s *KeeperTestSuite) TestQueryPendingMetrics() {
	filterMetricKey := "key-2"
	filterOracleChainId := "chain-2"

	// Add pending metrics across 2 keys and 2 oracles
	updatesByMetric := make(map[string][]types.Metric)
	updatesByOracle := make(map[string][]types.Metric)
	allPendingMetrics := []types.Metric{
		{Key: "key-1", DestinationOracle: "chain-1"},
		{Key: "key-2", DestinationOracle: "chain-1"},
		{Key: "key-1", DestinationOracle: "chain-2"},
		{Key: "key-2", DestinationOracle: "chain-2"},
	}
	for _, metric := range allPendingMetrics {
		key := metric.Key
		chainId := metric.DestinationOracle

		updatesByMetric[key] = append(updatesByMetric[key], metric)
		updatesByOracle[chainId] = append(updatesByOracle[chainId], metric)

		metric.Status = types.MetricStatus_METRIC_STATUS_IN_PROGRESS
		s.App.ICAOracleKeeper.SetMetric(s.Ctx, metric)
	}

	// First check with no filters
	expectedNoFilters := allPendingMetrics
	queryResponse, err := s.QueryClient.PendingMetrics(s.Ctx, &types.QueryPendingMetricsRequest{})
	s.Require().NoError(err, "no error expected when querying pending metric updates with no filter")
	s.Require().ElementsMatch(expectedNoFilters, queryResponse.Metrics)

	// Check with a filter on the metric (metric key == "key-2")
	queryResponse, err = s.QueryClient.PendingMetrics(s.Ctx, &types.QueryPendingMetricsRequest{
		MetricKey: filterMetricKey,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with metric key filter")
	s.Require().ElementsMatch(updatesByMetric[filterMetricKey], queryResponse.Metrics)

	// Check with a filter on the oracle (chain-id == "chain-2")
	queryResponse, err = s.QueryClient.PendingMetrics(s.Ctx, &types.QueryPendingMetricsRequest{
		OracleChainId: filterOracleChainId,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with oracle filter")
	s.Require().ElementsMatch(updatesByOracle[filterOracleChainId], queryResponse.Metrics)

	// Check with a filter on both the metric and oracle (metric key == "key2", chain-id == "chain-2")
	expectedMetricAndOracleFilter := []types.Metric{{
		Key:               filterMetricKey,
		DestinationOracle: filterOracleChainId,
	}}
	queryResponse, err = s.QueryClient.PendingMetrics(s.Ctx, &types.QueryPendingMetricsRequest{
		MetricKey:     filterMetricKey,
		OracleChainId: filterOracleChainId,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with metric key and oracle filter")
	s.Require().ElementsMatch(expectedMetricAndOracleFilter, queryResponse.Metrics)
}
