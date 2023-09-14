package cli_test

import "github.com/Stride-Labs/stride/v14/x/icaoracle/client/cli"

func (s *ClientTestSuite) TestCmdQueryOracle() {
	args := []string{
		HostChainId,
	}

	cmd := cli.GetCmdQueryOracle()
	s.ExecuteQueryAndCheckSuccessful(cmd, args, "query oracle")
}

func (s *ClientTestSuite) TestCmdQueryOracles() {
	cmd := cli.GetCmdQueryOracles()
	s.ExecuteQueryAndCheckSuccessful(cmd, []string{}, "query oracles")
	s.ExecuteQueryAndCheckSuccessful(cmd, []string{"--active", "true"}, "query active oracles")
	s.ExecuteQueryAndCheckSuccessful(cmd, []string{"--active", "false"}, "query inactive oracles")
}

func (s *ClientTestSuite) TestCmdQueryPendingMetricUpdates() {
	byMetric := []string{"--metric-key", "key"}
	byOracle := []string{"--oracle-chain-id", HostChainId}

	cmd := cli.GetCmdQueryMetrics()
	s.ExecuteQueryAndCheckSuccessful(cmd, []string{}, "query pending metric updates")
	s.ExecuteQueryAndCheckSuccessful(cmd, byMetric, "query pending metric updates by metric")
	s.ExecuteQueryAndCheckSuccessful(cmd, byOracle, "query pending metric updates by oracle")
	s.ExecuteQueryAndCheckSuccessful(cmd, append(byMetric, byOracle...), "query pending metric updates by metric and oracle")
}
