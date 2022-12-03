package cli_test

import (
	// "fmt"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"

	strideclitestutil "github.com/Stride-Labs/stride/v4/testutil/cli"

	"github.com/Stride-Labs/stride/v4/testutil/network"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/Stride-Labs/stride/v4/x/claim/client/cli"

	"github.com/Stride-Labs/stride/v4/app"
	cmdcfg "github.com/Stride-Labs/stride/v4/cmd/strided/config"
	"github.com/Stride-Labs/stride/v4/x/claim/types"
	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
)

var addr1 sdk.AccAddress
var addr2 sdk.AccAddress
var distributorMnemonics []string
var distributorAddrs []string

func init() {
	cmdcfg.SetupConfig()
	addr1 = ed25519.GenPrivKey().PubKey().Address().Bytes()
	addr2 = ed25519.GenPrivKey().PubKey().Address().Bytes()
	distributorMnemonics = []string{
		"chronic learn inflict great answer reward evidence stool open moon skate resource arch raccoon decade tell improve stay onion section blouse carry primary fabric",
		"catalog govern other escape eye resemble dirt hundred birth build dirt jacket network blame credit palace similar carry knock auction exotic bus business machine",
	}

	distributorAddrs = []string{
		"stride1ajerf2nmxsg0u728ga7665fmlfguqxcd8e36vf",
		"stride1zkfk3q70ranm3han4lvutvcvetncxg829j972a",
	}
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
			Address:           addr2.String(),
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: claimtypes.DefaultAirdropIdentifier,
		},
	}
	claimGenStateBz := s.cfg.Codec.MustMarshalJSON(claimGenState)
	genState[claimtypes.ModuleName] = claimGenStateBz

	s.cfg.GenesisState = genState
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	// Initiate distributor accounts
	val := s.network.Validators[0]
	for idx := range distributorMnemonics {
		info, _ := val.ClientCtx.Keyring.NewAccount("distributor"+strconv.Itoa(idx), distributorMnemonics[idx], keyring.DefaultBIP39Passphrase, sdk.FullFundraiserPath, hd.Secp256k1)
		distributorAddr := sdk.AccAddress(info.GetPubKey().Address())
		_, err = banktestutil.MsgSendExec(
			val.ClientCtx,
			val.Address,
			distributorAddr,
			sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1020)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			strideclitestutil.DefaultFeeString(s.cfg),
		)
		s.Require().NoError(err)
	}

	// Create a new airdrop
	cmd := cli.CmdCreateAirdrop()
	clientCtx := val.ClientCtx

	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
		claimtypes.DefaultAirdropIdentifier,
		strconv.Itoa(int(time.Now().Unix())),
		strconv.Itoa(int(claimtypes.DefaultAirdropDuration.Seconds())),
		s.cfg.BondDenom,
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, distributorAddrs[0]),
		// common args
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		strideclitestutil.DefaultFeeString(s.cfg),
	})

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
				claimtypes.DefaultAirdropIdentifier,
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

func (s *IntegrationTestSuite) TestCmdTxSetAirdropAllocations() {
	val := s.network.Validators[0]

	claimRecords := []claimtypes.ClaimRecord{
		{
			Address:           "stride1k8g9sagjpdwreqqf0qgqmd46l37595ea5ft9x6",
			Weight:            sdk.NewDecWithPrec(50, 2), // 50%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: claimtypes.DefaultAirdropIdentifier,
		},
		{
			Address:           "stride1av5lwh0msnafn04xkhdyk6mrykxthrawy7uf3d",
			Weight:            sdk.NewDecWithPrec(30, 2), // 30%
			ActionCompleted:   []bool{false, false, false},
			AirdropIdentifier: claimtypes.DefaultAirdropIdentifier,
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
				claimtypes.DefaultAirdropIdentifier,
				fmt.Sprintf("%s,%s", claimRecords[0].Address, claimRecords[1].Address),
				fmt.Sprintf("%s,%s", claimRecords[0].Weight.String(), claimRecords[1].Weight.String()),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, distributorAddrs[0]),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			[]sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(77))),
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(46))),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdSetAirdropAllocations()
			clientCtx := val.ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if claim record is properly set
			cmd = cli.GetCmdQueryClaimRecord()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				claimtypes.DefaultAirdropIdentifier,
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
				claimtypes.DefaultAirdropIdentifier,
				claimRecords[0].Address,
				types.ACTION_FREE.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result1 types.QueryClaimableForActionResponse
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result1))
			s.Require().Equal(tc.expClaimableAmounts[0].String(), result1.Coins.String())
		})
	}
}

func (s *IntegrationTestSuite) TestCmdTxCreateAirdrop() {
	val := s.network.Validators[0]

	airdrop := claimtypes.Airdrop{
		AirdropIdentifier:  "stride-1",
		AirdropStartTime:   time.Now(),
		AirdropDuration:    claimtypes.DefaultAirdropDuration,
		DistributorAddress: distributorAddrs[1],
		ClaimDenom:         claimtypes.DefaultClaimDenom,
	}

	testCases := []struct {
		name       string
		args       []string
		expAirdrop claimtypes.Airdrop
	}{
		{
			"create-airdrop tx",
			[]string{
				"stride-1",
				strconv.Itoa(int(time.Now().Unix())),
				strconv.Itoa(int(claimtypes.DefaultAirdropDuration.Seconds())),
				s.cfg.BondDenom,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, distributorAddrs[1]),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				strideclitestutil.DefaultFeeString(s.cfg),
			},
			airdrop,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdCreateAirdrop()
			clientCtx := val.ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			// Check if airdrop was created properly
			cmd = cli.GetCmdQueryParams()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			})
			s.Require().NoError(err)

			var result types.Params
			s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &result))
			s.Require().Equal(tc.expAirdrop.AirdropDuration, result.Airdrops[1].AirdropDuration)
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
