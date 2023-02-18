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

func CmdSubmitOpenBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-open-bid [host-zone] [pool-id] [bid]",
		Short: "Broadcast message submit-open-bid",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argZone := args[0]
			argPoolID := cast.ToUint64(args[1])
			argBid := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitOpenBid(
				clientCtx.GetFromAddress().String(),
				argZone,
				argPoolID,
				argBid,
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
