package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

var _ = strconv.Itoa(0)

func CmdStartAuction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-auction [host-zone] [pool-id]",
		Short: "Broadcast message start-auction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argZone := args[0]
			argPoolID := cast.ToUint64(args[1])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgStartAuction(
				clientCtx.GetFromAddress().String(),
				argZone,
				argPoolID,
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
