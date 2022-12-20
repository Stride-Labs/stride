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
	val     *network.Validator
}

var (
	path = types.Path{
		Denom:     sdk.DefaultBondDenom,
		ChannelId: "channel-0",
	}

	initialQuota = types.Quota{
		MaxPercentSend: 10,
		MaxPercentRecv: 20,
		DurationHours:  1,
	}

	initialFlow = types.Flow{
		Inflow:       0,
		Outflow:      0,
		ChannelValue: 500000000, // channel value == the network initial supply
	}

	updatedQuota = types.Quota{
		MaxPercentSend: 30,
		MaxPercentRecv: 50,
		DurationHours:  2,
	}
)

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

	s.val = s.network.Validators[0]
}

func (s *IntegrationTestSuite) addRateLimit() {
	args := []string{
		path.Denom,     // denom
		path.ChannelId, // channelId
		fmt.Sprintf("%d", initialQuota.MaxPercentSend), // maxPercentSend
		fmt.Sprintf("%d", initialQuota.MaxPercentRecv), // maxPercentRecv
		fmt.Sprintf("%d", initialQuota.DurationHours),  // durationHours
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.val.Address.String()),
		// common args
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		strideclitestutil.DefaultFeeString(s.cfg),
	}

	cmd := cli.CmdAddRateLimit()
	clientCtx := s.val.ClientCtx

	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) checkRateLimit(expectedQuota types.Quota, expectedFlow types.Flow) {
	clientCtx := s.val.ClientCtx

	cmd := cli.GetCmdQueryRateLimit()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
		path.Denom,
		path.ChannelId,
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	})
	s.Require().NoError(err)

	var rateLimit types.RateLimit
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &rateLimit))
	s.Require().Equal(path, *rateLimit.Path)
	s.Require().Equal(expectedQuota, *rateLimit.Quota)
	s.Require().Equal(expectedFlow, *rateLimit.Flow)
}

func (s *IntegrationTestSuite) TestCmdTxAddRateLimit() {
	// Add a rate limit
	s.addRateLimit()

	// Check that it was added properly
	s.checkRateLimit(initialQuota, initialFlow)
}

func (s *IntegrationTestSuite) TestCmdTxUpdateRateLimit() {
	// Add a rate limit
	s.addRateLimit()
	s.checkRateLimit(initialQuota, initialFlow)

	// Then update the rate limit
	updateArgs := []string{
		path.Denom,     // denom
		path.ChannelId, // channelId
		fmt.Sprintf("%d", updatedQuota.MaxPercentSend), // maxPercentSend
		fmt.Sprintf("%d", updatedQuota.MaxPercentRecv), // maxPercentRecv
		fmt.Sprintf("%d", updatedQuota.DurationHours),  // durationHours
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.val.Address.String()),
		// common args
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		strideclitestutil.DefaultFeeString(s.cfg),
	}

	cmd := cli.CmdUpdateRateLimit()
	_, err := clitestutil.ExecTestCLICmd(s.val.ClientCtx, cmd, updateArgs)
	s.Require().NoError(err)

	// Check if ratelimit was updated properly
	s.checkRateLimit(updatedQuota, initialFlow)
}

func (s *IntegrationTestSuite) TestCmdTxResetRateLimit() {
	// Add a rate limit
	s.addRateLimit()
	s.checkRateLimit(initialQuota, initialFlow)

	// Then reset the rate limit
	resetArgs := []string{
		path.Denom,
		path.ChannelId,
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.val.Address.String()),
		// common args
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		strideclitestutil.DefaultFeeString(s.cfg),
	}

	// Reset the rate limit
	cmd := cli.CmdResetRateLimit()
	_, err := clitestutil.ExecTestCLICmd(s.val.ClientCtx, cmd, resetArgs)
	s.Require().NoError(err)

	// Check if ratelimit was reset properly
	s.checkRateLimit(initialQuota, initialFlow)
}

func (s *IntegrationTestSuite) TestCmdTxRemoveRateLimit() {
	// Add a rate limit
	s.addRateLimit()
	s.checkRateLimit(initialQuota, initialFlow)

	// Then remove the rate limit
	removeArgs := []string{
		path.Denom,
		path.ChannelId,
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.val.Address.String()),
		// common args
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		strideclitestutil.DefaultFeeString(s.cfg),
	}

	cmd := cli.CmdRemoveRateLimit()
	_, err := clitestutil.ExecTestCLICmd(s.val.ClientCtx, cmd, removeArgs)
	s.Require().NoError(err)

	// Check if ratelimit was removed properly (get command should error)
	cmd = cli.GetCmdQueryRateLimit()
	_, err = clitestutil.ExecTestCLICmd(s.val.ClientCtx, cmd, []string{
		path.Denom,
		path.ChannelId,
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	})
	s.Require().Error(err)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
