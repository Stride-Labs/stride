package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var _ = strconv.Itoa(0)

func CmdUpdateValidatorSharesExchRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-delegation [chainid] [valoper]",
		Short: "Broadcast message update-delegation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostdenom := args[0]
			argValoper := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateValidatorSharesExchRate(
				clientCtx.GetFromAddress().String(),
				argHostdenom,
				argValoper,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
