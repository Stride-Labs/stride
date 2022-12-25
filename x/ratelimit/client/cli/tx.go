package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
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
	strideDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}

	if len(deposit) != 1 || deposit.GetDenomByIndex(0) != strideDenom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "Deposit token denom must be %s", strideDenom)
	}

	from := clientCtx.GetFromAddress()
	msg, err := govtypes.NewMsgSubmitProposal(proposal, deposit, from)
	if err != nil {
		return err
	}
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

// Adds a new rate limit proposal
func CmdAddRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-rate-limit [proposal-file]",
		Short: "Submit a add-rate-limit proposal",
		Args:  cobra.ExactArgs(1),
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

// Update a rate limit
func CmdUpdateRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-rate-limit [proposal-file]",
		Short: "Submit a update-rate-limit proposal",
		Args:  cobra.ExactArgs(1),
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

// Remove a rate limit
func CmdRemoveRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-rate-limit [proposal-file]",
		Short: "Submit a remove-rate-limit proposal",
		Args:  cobra.ExactArgs(1),
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

// Reset a rate limit
func CmdResetRateLimitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-rate-limit [proposal-file]",
		Short: "Submit a reset-rate-limit proposal",
		Args:  cobra.ExactArgs(1),
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
