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
// and add them to the queue
func (s *KeeperTestSuite) addPendingMetrics() []types.PendingMetricUpdate {
	// Add 5 metrics each to 2 oracles
	pendingMetrics := []types.PendingMetricUpdate{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		metricUpdate := types.PendingMetricUpdate{
			Metric: &types.Metric{
				Key:   "key-" + suffix,
				Value: "value-" + suffix,
			},
			OracleChainId: "chain-" + suffix,
		}

		pendingMetrics = append(pendingMetrics, metricUpdate)

		s.App.ICAOracleKeeper.SetMetricUpdateInProgress(s.Ctx, metricUpdate)
	}
	return pendingMetrics
}

func (s *KeeperTestSuite) TestGetPendingMetricUpdate() {
	pendingMetrics := s.addPendingMetrics()

	expectedPendingMetric := pendingMetrics[1]
	metricKey := expectedPendingMetric.Metric.Key
	oracleChainId := expectedPendingMetric.OracleChainId

	actualPendingMetric, found := s.App.ICAOracleKeeper.GetPendingMetricUpdate(s.Ctx, metricKey, oracleChainId)
	s.Require().True(found, "metric should have been found, but was not")
	s.Require().Equal(expectedPendingMetric, actualPendingMetric)
}

func (s *KeeperTestSuite) TestGetAllPendingMetricUpdates() {
	expectedPendingMetrics := s.addPendingMetrics()
	actualPendingMetrics := s.App.ICAOracleKeeper.GetAllPendingMetricUpdates(s.Ctx)
	s.Require().Len(actualPendingMetrics, len(expectedPendingMetrics), "number of metrics")
	s.Require().ElementsMatch(expectedPendingMetrics, actualPendingMetrics, "contents of metrics")
}

func (s *KeeperTestSuite) TestSetMetricUpdateComplete() {
	pendingMetrics := s.addPendingMetrics()

	expectedPendingMetric := pendingMetrics[1]
	metricKey := expectedPendingMetric.Metric.Key
	oracleChainId := expectedPendingMetric.OracleChainId

	s.App.ICAOracleKeeper.SetMetricUpdateComplete(s.Ctx, metricKey, oracleChainId)
	_, found := s.App.ICAOracleKeeper.GetPendingMetricUpdate(s.Ctx, metricKey, oracleChainId)
	s.Require().False(found, "the removed metric should not have been found, but it was")
}
