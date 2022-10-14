package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

var _ = strconv.Itoa(0)

func CmdAddValidatorProposal() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "add-validator-proposal [host-zone] [name] [address] [deposit]",
		Short: "Submit an add-validator proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an add-validator proposal along with an initial deposit.
Example:
$ %s tx gov submit-proposal add-validator-proposal juno-1 imperator juno123... --from=<key_or_address>
`, version.AppName),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostZone := args[0]
			argName := args[1]
			argAddress := args[2]
			argDeposit := args[3] // non-standard way to take a deposit but should work

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(argDeposit)
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()

			title := fmt.Sprintf("add-validator(%s) %s %s", argHostZone, argName, argAddress)
			description := fmt.Sprintf("Proposal to add %s validator %s with address %s", argHostZone, argName, argAddress)
			content := types.NewAddValidatorProposal(title, description, argHostZone, argName, argAddress)

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
