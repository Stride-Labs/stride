package cli_test

import (
	// "fmt"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"

	strideclitestutil "github.com/Stride-Labs/stride/testutil/cli"

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
var distributorMnemonic string
var distributorAddr string

func init() {
	cmdcfg.SetupConfig()
	addr1 = ed25519.GenPrivKey().PubKey().Address().Bytes()
	addr2 = ed25519.GenPrivKey().PubKey().Address().Bytes()
	distributorMnemonic = "chronic learn inflict great answer reward evidence stool open moon skate resource arch raccoon decade tell improve stay onion section blouse carry primary fabric"
	distributorAddr = "stride1ajerf2nmxsg0u728ga7665fmlfguqxcd8e36vf"
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
	claimGenState.Params.DistributorAddress = distributorAddr
	claimGenState.Params.AirdropStartTime = time.Now()
	claimGenState.Params.ClaimDenom = s.cfg.BondDenom
	claimGenState.ClaimRecords = []types.ClaimRecord{
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

	// Initiate distributor account
	val := s.network.Validators[0]
	info, _ := val.ClientCtx.Keyring.NewAccount("distributor", distributorMnemonic, keyring.DefaultBIP39Passphrase, sdk.FullFundraiserPath, hd.Secp256k1)
	distributorAddr := sdk.AccAddress(info.GetPubKey().Address())
	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		distributorAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 100000000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		strideclitestutil.DefaultFeeString(s.cfg),
	)
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

func (s *IntegrationTestSuite) TestCmdTxDepositAirdrop() {
	val := s.network.Validators[0]
	depositAmount := sdk.NewInt(1000)
	testCases := []struct {
		name               string
		args               []string
		expDepositedAmount sdk.Coins
	}{
		{
			"deposit-airdrop tx",
			[]string{
				fmt.Sprintf("%d%s", depositAmount.Int64(), s.cfg.BondDenom),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, distributorAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			sdk.Coins{sdk.NewCoin(s.cfg.BondDenom, depositAmount)},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdDepositAirdrop()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if claim record is properly set
			cmd = cli.GetCmdQueryModuleAccountBalance()
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result types.QueryModuleAccountBalanceResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(result.ModuleAccountBalance.String(), tc.expDepositedAmount.String())
		})
	}
}

func (s *IntegrationTestSuite) TestCmdTxSetAirdropAllocations() {
	val := s.network.Validators[0]

	claimRecords := []claimtypes.ClaimRecord{
		{
			Address:         "stride1k8g9sagjpdwreqqf0qgqmd46l37595ea5ft9x6",
			Weight:          sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted: []bool{false, false, false},
		},
		{
			Address:         "stride1av5lwh0msnafn04xkhdyk6mrykxthrawy7uf3d",
			Weight:          sdk.NewDecWithPrec(30, 2), // 30%
			ActionCompleted: []bool{false, false, false},
		},
	}

	testCases := []struct {
		name                string
		args                []string
		expClaimableAmounts []sdk.Coins
	}{
		{
			"set-airdrop-allocations tx",
			[]string{
				fmt.Sprintf("%s,%s", claimRecords[0].Address, claimRecords[1].Address),
				fmt.Sprintf("%s,%s", claimRecords[0].Weight.String(), claimRecords[1].Weight.String()),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, distributorAddr),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			[]sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(125))),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(75))),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdSetAirdropAllocations()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if claim record is properly set
			cmd = cli.GetCmdQueryClaimRecord()
			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				claimRecords[0].Address,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result types.QueryClaimRecordResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(result.ClaimRecord.String(), claimRecords[0].String())

			// Check if claimable amount for actions is correct
			cmd = cli.GetCmdQueryClaimableForAction()
			clientCtx = val.ClientCtx

			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				claimRecords[0].Address,
				types.ActionFree.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})

			var result1 types.QueryClaimableForActionResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result1))
			s.Require().Equal(tc.expClaimableAmounts[0].String(), result1.Coins.String())
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
