package cli_test

import (
	// "fmt"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	"github.com/Stride-Labs/stride/testutil/network"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/Stride-Labs/stride/x/claim/client/cli"

	"github.com/Stride-Labs/stride/app"
	cmdcfg "github.com/Stride-Labs/stride/cmd/strided/config"
	"github.com/Stride-Labs/stride/x/claim/types"
	claimtypes "github.com/Stride-Labs/stride/x/claim/types"
)

var addr1 sdk.AccAddress
var addr2 sdk.AccAddress

func init() {
	cmdcfg.SetupConfig()
	addr1 = ed25519.GenPrivKey().PubKey().Address().Bytes()
	addr2 = ed25519.GenPrivKey().PubKey().Address().Bytes()
}

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = network.DefaultConfig()

	genState := app.ModuleBasics.DefaultGenesis(s.cfg.Codec)
	claimGenState := claimtypes.DefaultGenesis()
	claimGenState.ClaimRecords = []types.ClaimRecord{
		{
			Address:         addr1.String(),
			Weight:          sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted: []bool{false, false, false},
		},
		{
			Address:         addr2.String(),
			Weight:          sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted: []bool{false, false, false},
		},
	}
	claimGenStateBz := s.cfg.Codec.MustMarshalJSON(claimGenState)
	genState[claimtypes.ModuleName] = claimGenStateBz

	s.cfg.GenesisState = genState
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	// s.T().Log("tearing down integration test suite")
	// s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestCmdQueryClaimRecord() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
	}{
		{
			"query claim record",
			[]string{
				addr1.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryClaimRecord()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.QueryClaimRecordResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
		})
	}
}

func (s *IntegrationTestSuite) TestCmdQueryClaimableForAction() {
	val := s.network.Validators[0]

	testCases := []struct {
		name  string
		args  []string
		coins sdk.Coins
	}{
		{
			"query claimable-for-action amount",
			[]string{
				addr2.String(),
				types.ActionFree.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5))},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryClaimableForAction()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			var result types.QueryClaimableForActionResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(result.Coins.String(), tc.coins.String())
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
