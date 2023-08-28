package cli_test

import (
	"fmt"
	"testing"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v14/app"
	cmdcfg "github.com/Stride-Labs/stride/v14/cmd/strided/config"
	strideclitestutil "github.com/Stride-Labs/stride/v14/testutil/cli"
	"github.com/Stride-Labs/stride/v14/testutil/network"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

var (
	HostChainId = "chain-1"
)

type ClientTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	val     *network.Validator
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) SetupSuite() {
	s.T().Log("setting up client test suite")

	s.cfg = network.DefaultConfig()

	genState := app.ModuleBasics.DefaultGenesis(s.cfg.Codec)

	// Add an oracle to the store for the query command
	icaoracleGenstate := types.DefaultGenesis()
	icaoracleGenstate.Oracles = []types.Oracle{{ChainId: HostChainId}}
	icaoracleGenstateBz := s.cfg.Codec.MustMarshalJSON(icaoracleGenstate)
	genState[types.ModuleName] = icaoracleGenstateBz

	s.cfg.GenesisState = genState
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.val = s.network.Validators[0]

	cmdcfg.RegisterDenoms()
}

func (s *ClientTestSuite) ExecuteTxAndCheckSuccessful(cmd *cobra.Command, args []string, description string) {
	defaultFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		strideclitestutil.DefaultFeeString(s.cfg),
	}
	args = append(args, defaultFlags...)

	clientCtx := s.val.ClientCtx
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	var response sdk.TxResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response))
}

func (s *ClientTestSuite) ExecuteQueryAndCheckSuccessful(cmd *cobra.Command, args []string, description string) {
	clientCtx := s.val.ClientCtx
	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
}
