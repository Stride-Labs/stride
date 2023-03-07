package keeper_test

import (
	"strconv"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// Helper function to create 5 metric objects with various attributes
// and add them to the queue
func (s *KeeperTestSuite) addMetricsToQueue() []types.Metric {
	metrics := []types.Metric{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		metric := types.Metric{
			Key:   "key-" + suffix,
			Value: "value-" + suffix,
		}

		metrics = append(metrics, metric)
		s.App.ICAOracleKeeper.QueueMetricUpdate(s.Ctx, metric)
	}
	return metrics
}

func (s *KeeperTestSuite) TestGetMetricFromQueue() {
	metrics := s.addMetricsToQueue()

	expectedMetric := metrics[1]
	metricKey := expectedMetric.Key

	actualMetric, found := s.App.ICAOracleKeeper.GetMetricFromQueue(s.Ctx, metricKey)
	s.Require().True(found, "metric should have been found, but was not")
	s.Require().Equal(expectedMetric, actualMetric)

	_, found = s.App.ICAOracleKeeper.GetMetricFromQueue(s.Ctx, "fake_key")
	s.Require().False(found, "metric does not exist and should not have been found")
}

func (s *KeeperTestSuite) TestGetAllMetricsFromQueue() {
	expectedMetrics := s.addMetricsToQueue()
	actualMetrics := s.App.ICAOracleKeeper.GetAllMetricsFromQueue(s.Ctx)
	s.Require().Len(actualMetrics, len(expectedMetrics), "number of metrics")
	s.Require().ElementsMatch(expectedMetrics, actualMetrics, "contents of metrics")
}

func (s *KeeperTestSuite) TestRemoveMetricFromQueue() {
	metrics := s.addMetricsToQueue()

	metricToRemove := metrics[1]
	metricKey := metricToRemove.Key

	s.App.ICAOracleKeeper.RemoveMetricFromQueue(s.Ctx, metricKey)
	_, found := s.App.ICAOracleKeeper.GetMetricFromQueue(s.Ctx, metricKey)
	s.Require().False(found, "the removed metric should not have been found, but it was")
}

// Helper function to create 5 metric objects with various attributes
// and add them to the pending store
func (s *KeeperTestSuite) addPendingUpdates() []types.PendingMetricUpdate {
	// Add 5 metrics each across 5 oracles
	pendingMetrics := []types.PendingMetricUpdate{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		metricUpdate := types.PendingMetricUpdate{
			Metric: &types.Metric{
				Key:        "key-" + suffix,
				Value:      "value-" + suffix,
				UpdateTime: uint64(i),
			},
			OracleChainId: "chain-" + suffix,
		}

		pendingMetrics = append(pendingMetrics, metricUpdate)

		s.App.ICAOracleKeeper.SetMetricUpdateInProgress(s.Ctx, metricUpdate)
	}
	return pendingMetrics
}

// Helper function to create 5 metric objects with the same key and oracle, but various times
// Returns the list of metrics (instead of metric updates)
func (s *KeeperTestSuite) addPendingUpdatesWithSameKey() []types.Metric {
	// Add 5 metrics each across 5 oracles
	metrics := []types.Metric{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		metric := types.Metric{
			Key:        "key1",
			Value:      "value-" + suffix,
			UpdateTime: uint64(i),
		}
		metricUpdate := types.PendingMetricUpdate{
			Metric:        &metric,
			OracleChainId: HostChainId,
		}

		metrics = append(metrics, metric)

		s.App.ICAOracleKeeper.SetMetricUpdateInProgress(s.Ctx, metricUpdate)
	}
	return metrics
}

func (s *KeeperTestSuite) TestGetPendingMetricUpdate() {
	pendingUpdates := s.addPendingUpdates()

	expectedPendingUpdates := pendingUpdates[1]
	metricKey := expectedPendingUpdates.Metric.Key
	oracleChainId := expectedPendingUpdates.OracleChainId
	updateTime := expectedPendingUpdates.Metric.UpdateTime

	expectedPendingUpdates, found := s.App.ICAOracleKeeper.GetPendingMetricUpdate(s.Ctx, metricKey, oracleChainId, updateTime)
	s.Require().True(found, "metric should have been found, but was not")
	s.Require().Equal(expectedPendingUpdates, expectedPendingUpdates)

	_, found = s.App.ICAOracleKeeper.GetPendingMetricUpdate(s.Ctx, "fake_key", oracleChainId, updateTime)
	s.Require().False(found, "metric does not exist and should not have been found")
}

func (s *KeeperTestSuite) TestGetAllPendingMetricUpdates() {
	expectedPendingUpdates := s.addPendingUpdates()
	actualPendingUpdates := s.App.ICAOracleKeeper.GetAllPendingMetricUpdates(s.Ctx)
	s.Require().Len(actualPendingUpdates, len(expectedPendingUpdates), "number of metrics")
	s.Require().ElementsMatch(expectedPendingUpdates, actualPendingUpdates, "contents of metrics")
}

func (s *KeeperTestSuite) TestGetPendingMetrics() {
	expectedMetrics := s.addPendingUpdatesWithSameKey()

	// Add additional metrics with the same key
	expectedMetric := expectedMetrics[0]
	metricKey := expectedMetric.Key

	actualMetrics := s.App.ICAOracleKeeper.GetPendingMetrics(s.Ctx, metricKey, HostChainId)
	s.Require().Len(actualMetrics, len(expectedMetrics), "number of metrics")
	s.Require().ElementsMatch(expectedMetrics, actualMetrics, "contents of metrics")
}

func (s *KeeperTestSuite) TestSetMetricUpdateComplete() {
	pendingMetrics := s.addPendingUpdates()

	expectedPendingMetric := pendingMetrics[1]
	metricKey := expectedPendingMetric.Metric.Key
	oracleChainId := expectedPendingMetric.OracleChainId
	updateTime := expectedPendingMetric.Metric.UpdateTime

	s.App.ICAOracleKeeper.SetMetricUpdateComplete(s.Ctx, metricKey, oracleChainId, updateTime)
	_, found := s.App.ICAOracleKeeper.GetPendingMetricUpdate(s.Ctx, metricKey, oracleChainId, updateTime)
	s.Require().False(found, "the removed metric should not have been found, but it was")
}
