package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (s *IntegrationTestSuite) TestCmdTxAddRateLimit() {
	val := s.network.Validators[0]
	quota := types.Quota{
		MaxPercentSend: 10,
		MaxPercentRecv: 20,
		DurationHours:  1,
	}
	path := types.Path{
		Denom:     sdk.DefaultBondDenom,
		ChannelId: "channel-0",
	}

	testCases := []struct {
		name     string
		args     []string
		expQuota types.Quota
	}{
		{
			"add-rate-limit tx",
			[]string{
				sdk.DefaultBondDenom,                    // denom
				path.ChannelId,                          // channelId
				fmt.Sprintf("%d", quota.MaxPercentSend), // maxPercentSend
				fmt.Sprintf("%d", quota.MaxPercentRecv), // maxPercentRecv
				fmt.Sprintf("%d", quota.DurationHours),  // durationHours
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
			cmd := cli.CmdAddRateLimit()
			clientCtx := val.ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if ratelimit was set properly
			cmd = cli.GetCmdQueryRateLimit()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				path.Denom,
				path.ChannelId,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result types.RateLimit
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.expQuota, *result.Quota)
		})
	}
}

func (s *IntegrationTestSuite) TestCmdTxUpdateRateLimit() {
	val := s.network.Validators[0]
	quota := types.Quota{
		MaxPercentSend: 20,
		MaxPercentRecv: 30,
		DurationHours:  1,
	}
	path := types.Path{
		Denom:     sdk.DefaultBondDenom,
		ChannelId: "channel-1",
	}

	testCases := []struct {
		name     string
		args     []string
		expQuota types.Quota
	}{
		{
			"update-rate-limit tx",
			[]string{
				sdk.DefaultBondDenom,                    // denom
				path.ChannelId,                          // channelId
				fmt.Sprintf("%d", quota.MaxPercentSend), // maxPercentSend
				fmt.Sprintf("%d", quota.MaxPercentRecv), // maxPercentRecv
				fmt.Sprintf("%d", quota.DurationHours),  // durationHours
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
			cmd := cli.CmdAddRateLimit()
			args := make([]string, len(tc.args))
			copy(args, tc.args)
			args[2] = "10"
			args[3] = "20"
			clientCtx := val.ClientCtx

			// Add a new ratelimit
			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			s.Require().NoError(err)

			// Update the rate limit
			cmd = cli.CmdUpdateRateLimit()
			_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if ratelimit was updated properly
			cmd = cli.GetCmdQueryRateLimit()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				path.Denom,
				path.ChannelId,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result types.RateLimit
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.expQuota, *result.Quota)
		})
	}
}

func (s *IntegrationTestSuite) TestCmdTxResetRateLimit() {
	val := s.network.Validators[0]
	quota := types.Quota{
		MaxPercentSend: 20,
		MaxPercentRecv: 30,
		DurationHours:  1,
	}
	path := types.Path{
		Denom:     sdk.DefaultBondDenom,
		ChannelId: "channel-2",
	}
	flow := types.Flow{
		Inflow:       0,
		Outflow:      0,
		ChannelValue: s.cfg.StakingTokens.Uint64(),
	}

	testCases := []struct {
		name    string
		args    []string
		expFlow types.Flow
	}{
		{
			"reset-rate-limit tx",
			[]string{
				sdk.DefaultBondDenom, // denom
				path.ChannelId,       // channelId
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			flow,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdAddRateLimit()
			clientCtx := val.ClientCtx

			// Add a new ratelimit
			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				sdk.DefaultBondDenom,                    // denom
				path.ChannelId,                          // channelId
				fmt.Sprintf("%d", quota.MaxPercentSend), // maxPercentSend
				fmt.Sprintf("%d", quota.MaxPercentRecv), // maxPercentRecv
				fmt.Sprintf("%d", quota.DurationHours),  // durationHours
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			})
			s.Require().NoError(err)

			// Reset the rate limit
			cmd = cli.CmdResetRateLimit()
			_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if ratelimit was reset properly
			cmd = cli.GetCmdQueryRateLimit()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				path.Denom,
				path.ChannelId,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result types.RateLimit
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.expFlow, *result.Flow)
		})
	}
}

func (s *IntegrationTestSuite) TestCmdTxRemoveRateLimit() {
	val := s.network.Validators[0]

	rateLimit := types.RateLimit{}
	denom := sdk.DefaultBondDenom
	channelId := "channel-0"

	testCases := []struct {
		name         string
		args         []string
		expRateLimit types.RateLimit
	}{
		{
			"remove-ratelimit tx",
			[]string{
				denom,     // denom
				channelId, // channel id
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			rateLimit,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdRemoveRateLimit()
			clientCtx := val.ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if ratelimit was removed properly
			cmd = cli.GetCmdQueryRateLimit()
			_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				denom,
				channelId,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().Error(err)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
