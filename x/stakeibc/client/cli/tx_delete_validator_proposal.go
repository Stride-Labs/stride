package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func CmdDeleteValidatorProposal() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "delete-validator-proposal [host-zone] [address] [deposit]",
		Short: "Submit an delete-validator proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an delete-validator proposal along with an initial deposit.
Example:
$ %s tx gov submit-proposal delete-validator-proposal juno-1 juno123... --from=<key_or_address>
`, version.AppName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostZone := args[0]
			argAddress := args[1]
			argDeposit := args[2] // non-standard way to take a deposit but should work

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(argDeposit)
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()

			title := fmt.Sprintf("delete-validator(%s) %s", argHostZone, argAddress)
			description := fmt.Sprintf("Proposal to remove %s validator with address %s", argHostZone, argAddress)
			content := types.NewDeleteValidatorProposal(title, description, argHostZone, argAddress)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
