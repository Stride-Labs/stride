package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v9/x/ratelimit/types"
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
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	// Finally, broadcast the proposal tx
	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

// Adds a new rate limit proposal
func CmdAddRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-rate-limit [proposal-file]",
		Short: "Submit a add-rate-limit proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an add-rate-limit proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal add-rate-limit <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
	"title": "Add Rate Limit to ...",
    "description": "Proposal to enable rate limiting on...",
    "denom": "ustrd",
    "channel_id": "channel-0",
    "max_percent_send": "10",
	"max_percent_recv": "10",
	"duration_hours": "24", 
    "deposit": "10000000ustrd"
}
`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			proposalFile := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var proposal types.AddRateLimitProposal
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

// Update a rate limit
func CmdUpdateRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-rate-limit [proposal-file]",
		Short: "Submit a update-rate-limit proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an update-rate-limit proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal update-rate-limit <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
	"title": "Update Rate Limit ...",
    "description": "Proposal to update rate limit...",
    "denom": "ustrd",
    "channel_id": "channel-0",
    "max_percent_send": "10",
	"max_percent_recv": "20",
	"duration_hours": "24", 
    "deposit": "10000000ustrd"
}
`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			proposalFile := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var proposal types.UpdateRateLimitProposal
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

// Remove a rate limit
func CmdRemoveRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-rate-limit [proposal-file]",
		Short: "Submit a remove-rate-limit proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an remove-rate-limit proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal remove-rate-limit <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
	"title": "Remove Rate Limit ...",
    "description": "Proposal to remove rate limiting on...",
    "denom": "ustrd",
    "channel_id": "channel-0",
    "deposit": "10000000ustrd"
}
`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			proposalFile := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var proposal types.RemoveRateLimitProposal
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

// Reset a rate limit
func CmdResetRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-rate-limit [proposal-file]",
		Short: "Submit a reset-rate-limit proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an reset-rate-limit proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal reset-rate-limit <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
	"title": "Reset Rate Limit ...",
    "description": "Proposal to reset the rate limit...",
    "denom": "ustrd",
    "channel_id": "channel-0",
    "deposit": "10000000ustrd"
}
`, version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			proposalFile := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var proposal types.ResetRateLimitProposal
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
