package cli

import (
	"fmt"
	"os"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v11/x/icaoracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Parse the gov proposal file into a proto message
func parseProposalFile(cdc codec.JSONCodec, proposalFile string, proposal proto.Message) error {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return err
	}

	if err = cdc.UnmarshalJSON(contents, proposal); err != nil {
		return err
	}

	return nil
}

// Submits the governance proposal
func submitProposal(clientCtx client.Context, cmd *cobra.Command, proposal govtypes.Content, deposit sdk.Coins) error {
	// Confirm a valid deposit was submitted
	strideDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}
	if len(deposit) != 1 || deposit.GetDenomByIndex(0) != strideDenom {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "Deposit token denom must be %s", strideDenom)
	}

	// Build and validate the proposal
	from := clientCtx.GetFromAddress()
	msg, err := govtypes.NewMsgSubmitProposal(proposal, deposit, from)
	if err != nil {
		return err
	}

	// Finally, broadcast the proposal tx
	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

// Proposal to toggle whether an oracle is active
func CmdToggleOracleProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toggle-oracle [proposal-file]",
		Short: "Submit a proposal to toggle whether an oracle is active",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a toggle-oracle proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal toggle-oracle <path/to/proposal.json> --from=<key_or_address>
Where proposal.json contains:
{
	"title": "Toggle Oracle ...",
    "description": "Deactivate oracle because ...",
    "oracle_chain_id": "osmosis-1",
}
`, version.AppName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalFile := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var proposal types.ToggleOracleProposal
			if err := parseProposalFile(clientCtx.Codec, proposalFile, &proposal); err != nil {
				return err
			}

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

			return submitProposal(clientCtx, cmd, &proposal, deposit)
		},
	}

	cmd.Flags().String(govcli.FlagDeposit, "", "deposit of proposal")

	return cmd
}

// Proposal to remove an oracle
func CmdRemoveOracleProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-oracle [proposal-file]",
		Short: "Submit a proposal to remove an oracle",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a remove-oracle proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal remove-oracle <path/to/proposal.json> --from=<key_or_address>
Where proposal.json contains:
{
	"title": "Remove Oracle ...",
    "description": "Remove oracle because ...",
    "oracle_chain_id": "osmosis-1",
}
`, version.AppName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalFile := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var proposal types.RemoveOracleProposal
			if err := parseProposalFile(clientCtx.Codec, proposalFile, &proposal); err != nil {
				return err
			}

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

			return submitProposal(clientCtx, cmd, &proposal, deposit)
		},
	}

	cmd.Flags().String(govcli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
