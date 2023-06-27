package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

func parseRegisterHostZoneProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.RegisterHostZoneProposal, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}

func CmdRegisterHostZoneProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-host-zone [proposal-file]",
		Short: "Submit a register-host-zone proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a register-host-zone proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal register-host-zone <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
	"title": "Register gaia as a host zone",
    "description": "Proposal to register gaia as host zone.",
	"connection_id": "connection-0",
	"bech32prefix": "cosmos",
	"host_denom": "uatom",
	"ibc_denom": "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
	"transfer_channel_id": "channel-0",
	"unbonding_frequency": 1,
	"min_redemption_rate": "0.5",
	"max_redemption_rate": "5.0",
    "deposit": "64000000ustrd"
}
`, version.AppName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := parseRegisterHostZoneProposalFile(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositFromFlags, err := cmd.Flags().GetString(govcli.FlagDeposit)
			if err != nil {
				return err
			}

			// if deposit from flags is not empty, it overrides the deposit from proposal
			if depositFromFlags != "" {
				proposal.Deposit = depositFromFlags
			}
			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			strideDenom, err := sdk.GetBaseDenom()
			if err != nil {
				return err
			}

			if len(deposit) != 1 || deposit.GetDenomByIndex(0) != strideDenom {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "Deposit token denom must be %s", strideDenom)
			}

			msg, err := govtypes.NewMsgSubmitProposal(&proposal, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
