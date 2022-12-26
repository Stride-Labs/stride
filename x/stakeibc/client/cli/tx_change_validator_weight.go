package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var _ = strconv.Itoa(0)

func CmdChangeValidatorWeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-validator-weight [host-zone] [address] [weight]",
		Short: "Broadcast message change-validator-weight",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostZone := args[0]
			argName := args[1]
			argWeight, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("Fail to parse arg to sdk.Int (%v)", args[2])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeValidatorWeight(
				clientCtx.GetFromAddress().String(),
				argHostZone,
				argName,
				argWeight,
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
