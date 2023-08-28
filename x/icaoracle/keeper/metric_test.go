package keeper_test

import (
	"strconv"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

// Helper function to create 5 metric objects with various attributes
func (s *KeeperTestSuite) createMetrics() []types.Metric {
	metrics := []types.Metric{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		metric := types.Metric{
			Key:               "key-" + suffix,
			Value:             "value-" + suffix,
			DestinationOracle: "chain",
			Status:            types.MetricStatus_QUEUED,
		}

		metrics = append(metrics, metric)
		s.App.ICAOracleKeeper.SetMetric(s.Ctx, metric)
	}
	return metrics
}

func (s *KeeperTestSuite) TestGetMetric() {
	metrics := s.createMetrics()

	for _, expected := range metrics {
		metricId := expected.GetMetricID()

		actual, found := s.App.ICAOracleKeeper.GetMetric(s.Ctx, metricId)
		s.Require().True(found, "metric %s should have been found", metricId)
		s.Require().Equal(expected, actual, "metric %s", metricId)
	}
}

func (s *KeeperTestSuite) TestGetAllMetrics() {
	metrics := s.createMetrics()

	actualMetrics := s.App.ICAOracleKeeper.GetAllMetrics(s.Ctx)
	s.Require().Equal(len(actualMetrics), len(metrics), "number of metrics")

	for i, expected := range metrics {
		metricId := expected.GetMetricID()

		actual := actualMetrics[i]
		s.Require().Equal(expected, actual, "metrics %s", metricId)
	}
}

func (s *KeeperTestSuite) TestMetricQueue() {
	metrics := s.createMetrics()

	actualQueuedMetrics := s.App.ICAOracleKeeper.GetAllQueuedMetrics(s.Ctx)
	s.Require().Equal(len(actualQueuedMetrics), len(metrics), "number of queued metrics")

	for i, metric := range metrics {
		metricId := metric.GetMetricID()

		// set the metric to in progres which should remove it from the queue
		s.App.ICAOracleKeeper.UpdateMetricStatus(s.Ctx, metric, types.MetricStatus_IN_PROGRESS)
		metricsInProgress := i + 1
		queuedMetrics := s.App.ICAOracleKeeper.GetAllQueuedMetrics(s.Ctx)

		s.Require().Equal(len(queuedMetrics), len(metrics)-metricsInProgress,
			"number of remaining queued metrics after updating %s", metricId)

		s.Require().ElementsMatch(metrics[metricsInProgress:], queuedMetrics,
			"queued metric after setting %s to in progress", metricId)
	}
}
