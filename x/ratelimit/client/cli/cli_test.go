package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	strideclitestutil "github.com/Stride-Labs/stride/v4/testutil/cli"

	"github.com/Stride-Labs/stride/v4/testutil/network"

	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/client/cli"

	"github.com/Stride-Labs/stride/v4/app"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = network.DefaultConfig()

	genState := app.ModuleBasics.DefaultGenesis(s.cfg.Codec)
	ratelimitGenState := types.DefaultGenesis()
	ratelimitGenStateBz := s.cfg.Codec.MustMarshalJSON(ratelimitGenState)
	genState[types.ModuleName] = ratelimitGenStateBz

	s.cfg.GenesisState = genState
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestCmdTxAddQuota() {
	val := s.network.Validators[0]

	quota := types.Quota{
		Name:            "quota",
		MaxPercentSend:  10,
		MaxPercentRecv:  20,
		DurationMinutes: 30,
	}

	testCases := []struct {
		name     string
		args     []string
		expQuota types.Quota
	}{
		{
			"add-quota tx",
			[]string{
				"quota", // name
				"10",    // maxPercentSend
				"20",    // maxPercentRecv
				"30",    // durationMinutes
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			quota,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdAddQuota()
			clientCtx := val.ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if quota was set properly
			cmd = cli.GetCmdQueryQuota()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				"quota",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result types.Quota
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.expQuota.Name, result.Name)
		})
	}
}

func (s *IntegrationTestSuite) TestCmdTxRemoveQuota() {
	val := s.network.Validators[0]

	quota := types.Quota{}

	testCases := []struct {
		name     string
		args     []string
		expQuota types.Quota
	}{
		{
			"remove-quota tx",
			[]string{
				"quota", // name
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			quota,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdRemoveQuota()
			clientCtx := val.ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if quota was removed properly
			cmd = cli.GetCmdQueryQuota()
			_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				"quota",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().Error(err)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
