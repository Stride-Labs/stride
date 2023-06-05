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

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func parseAddValidatorsProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.AddValidatorsProposal, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	proposal.Title = fmt.Sprintf("Add validators to %s", proposal.HostZone)

	return proposal, nil
}

func CmdAddValidatorsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-validators [proposal-file]",
		Short: "Submit an add-validator proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an add-validators proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal add-validators <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
    "description": "Proposal to add Imperator because they contribute in XYZ ways!",
    "hostZone": "GAIA",
    "validators": [
		{
			"name": "Imperator",
    		"address": "cosmosvaloper1v5y0tg0jllvxf5c3afml8s3awue0ymju89frut",
		},
	],
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

			proposal, err := parseAddValidatorsProposalFile(clientCtx.Codec, args[0])
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
