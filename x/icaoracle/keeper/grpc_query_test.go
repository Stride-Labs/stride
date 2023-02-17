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
			Moniker: expectedOracle.Moniker,
		})
		s.Require().NoError(err, "no error expected when querying oracle %s", expectedOracle.Moniker)
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
		oracle := types.Oracle{Moniker: fmt.Sprintf("moniker-%d", i)}
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
	filterKey := "key2"
	filterMoniker := "moniker2"

	// Add pending metrics across 2 keys and 2 oracles
	updatesByKey := make(map[string][]types.PendingMetricUpdate)
	updatesByMoniker := make(map[string][]types.PendingMetricUpdate)
	allPendingUpdates := []types.PendingMetricUpdate{
		{
			Metric: &types.Metric{Key: "key1"}, OracleMoniker: "moniker1",
		},
		{
			Metric: &types.Metric{Key: "key2"}, OracleMoniker: "moniker1",
		},
		{
			Metric: &types.Metric{Key: "key1"}, OracleMoniker: "moniker2",
		},
		{
			Metric: &types.Metric{Key: "key2"}, OracleMoniker: "moniker2",
		},
	}
	for _, pendingUpdate := range allPendingUpdates {
		key := pendingUpdate.Metric.Key
		moniker := pendingUpdate.OracleMoniker

		updatesByKey[key] = append(updatesByKey[key], pendingUpdate)
		updatesByMoniker[moniker] = append(updatesByMoniker[moniker], pendingUpdate)

		s.App.ICAOracleKeeper.SetMetricUpdateInProgress(s.Ctx, pendingUpdate)
	}

	// First check with no filters
	expectedNoFilters := allPendingUpdates
	queryResponse, err := s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{})
	s.Require().NoError(err, "no error expected when querying pending metric updates with no filter")
	s.Require().ElementsMatch(expectedNoFilters, queryResponse.PendingUpdates)

	// Check with a filter on the key (key == "key2")
	queryResponse, err = s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{
		MetricKey: filterKey,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with key filter")
	s.Require().ElementsMatch(updatesByKey[filterKey], queryResponse.PendingUpdates)

	// Check with a filter on the moniker (moniker == "moniker2")
	queryResponse, err = s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{
		OracleMoniker: filterMoniker,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with key filter")
	s.Require().ElementsMatch(updatesByMoniker[filterMoniker], queryResponse.PendingUpdates)

	// Check with a filter on both the key and moniker (key == "key2", moniker == "moniker2")
	expectedKeyAndMonikerFilter := []types.PendingMetricUpdate{{
		Metric:        &types.Metric{Key: filterKey},
		OracleMoniker: filterMoniker,
	}}
	queryResponse, err = s.QueryClient.PendingMetricUpdates(s.Ctx, &types.QueryPendingMetricUpdatesRequest{
		MetricKey:     filterKey,
		OracleMoniker: filterMoniker,
	})
	s.Require().NoError(err, "no error expected when querying pending metric updates with key filter")
	s.Require().ElementsMatch(expectedKeyAndMonikerFilter, queryResponse.PendingUpdates)
}
