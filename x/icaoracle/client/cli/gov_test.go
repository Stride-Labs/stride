package cli_test

import (
	"github.com/Stride-Labs/stride/v5/x/icaoracle/client/cli"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func (s *ClientTestSuite) TestCmdToggleOracleProposal() {
	proposal := types.ToggleOracleProposal{
		Title:         "toggle oracle",
		Description:   "toggle oracle",
		Deposit:       "10ustrd",
		OracleChainId: HostChainId,
		Active:        true,
	}

	cmd := cli.CmdToggleOracleProposal()
	s.ExecuteGovTxAndCheckSuccessful(cmd, &proposal, "toggle oracle")
}

func (s *ClientTestSuite) TestCmdRemoveOracleProposal() {
	proposal := types.RemoveOracleProposal{
		Title:         "remove oracle",
		Description:   "remove oracle",
		Deposit:       "10ustrd",
		OracleChainId: HostChainId,
	}

	cmd := cli.CmdRemoveOracleProposal()
	s.ExecuteGovTxAndCheckSuccessful(cmd, &proposal, "remove oracles")
}
