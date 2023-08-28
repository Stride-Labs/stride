package keeper_test

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func (s *KeeperTestSuite) TestQueryOracle() {
	allOracles := s.CreateTestOracles()
	for _, expectedOracle := range allOracles {
		queryResponse, err := s.QueryClient.Oracle(context.Background(), &types.QueryOracleRequest{
			ChainId: expectedOracle.ChainId,
		})
		s.Require().NoError(err, "no error expected when querying oracle %s", expectedOracle.ChainId)
		s.Require().Equal(expectedOracle, *queryResponse.Oracle)
	}
}

func (s *KeeperTestSuite) TestQueryAllOracles() {
	expectedOracles := s.CreateTestOracles()
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

func (s *KeeperTestSuite) TestQueryMetrics() {
	filterMetricKey := "key-2"
	filterOracleChainId := "chain-2"

	// Add metrics across 2 keys and 2 oracles
	updatesByMetric := make(map[string][]types.Metric)
	updatesByOracle := make(map[string][]types.Metric)
	allMetrics := []types.Metric{
		{Key: "key-1", DestinationOracle: "chain-1", Status: types.MetricStatus_IN_PROGRESS},
		{Key: "key-2", DestinationOracle: "chain-1", Status: types.MetricStatus_IN_PROGRESS},
		{Key: "key-1", DestinationOracle: "chain-2", Status: types.MetricStatus_IN_PROGRESS},
		{Key: "key-2", DestinationOracle: "chain-2", Status: types.MetricStatus_IN_PROGRESS},
	}
	for _, metric := range allMetrics {
		key := metric.Key
		chainId := metric.DestinationOracle

		updatesByMetric[key] = append(updatesByMetric[key], metric)
		updatesByOracle[chainId] = append(updatesByOracle[chainId], metric)

		s.App.ICAOracleKeeper.SetMetric(s.Ctx, metric)
	}

	// First check with no filters
	expectedNoFilters := allMetrics
	queryResponse, err := s.QueryClient.Metrics(s.Ctx, &types.QueryMetricsRequest{})
	s.Require().NoError(err, "no error expected when querying pending metric updates with no filter")
	s.Require().ElementsMatch(expectedNoFilters, queryResponse.Metrics, "no filter")

	// Check with a filter on the metric (metric key == "key-2")
	queryResponse, err = s.QueryClient.Metrics(s.Ctx, &types.QueryMetricsRequest{
		MetricKey: filterMetricKey,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with metric key filter")
	s.Require().ElementsMatch(updatesByMetric[filterMetricKey], queryResponse.Metrics, "metric key filter")

	// Check with a filter on the oracle (chain-id == "chain-2")
	queryResponse, err = s.QueryClient.Metrics(s.Ctx, &types.QueryMetricsRequest{
		OracleChainId: filterOracleChainId,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with oracle filter")
	s.Require().ElementsMatch(updatesByOracle[filterOracleChainId], queryResponse.Metrics, "chain-id filter")

	// Check with a filter on both the metric and oracle (metric key == "key2", chain-id == "chain-2")
	expectedMetricAndOracleFilter := []types.Metric{{
		Key:               filterMetricKey,
		DestinationOracle: filterOracleChainId,
		Status:            types.MetricStatus_IN_PROGRESS,
	}}
	queryResponse, err = s.QueryClient.Metrics(s.Ctx, &types.QueryMetricsRequest{
		MetricKey:     filterMetricKey,
		OracleChainId: filterOracleChainId,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with metric key and oracle filter")
	s.Require().ElementsMatch(expectedMetricAndOracleFilter, queryResponse.Metrics, "metric and chain filter")
}
