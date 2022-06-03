package cli

import (
	"strconv"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)


// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
func CmdRegisterHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "register-host-zone [connection-id] [base-denom] [local-denom]",
		Short: "Broadcast message register-host-zone",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			connectionId := args[0]
			baseDenom := args[1]
			localDenom := args[2]
			msg := types.NewMsgRegisterHostZone(
				clientCtx.GetFromAddress().String(),
				connectionId,
				baseDenom,
				localDenom,
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
