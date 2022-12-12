package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func CmdAddQuota() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-quota [name] [max-percent-send] [max-percent-recv] [duration-minutes]",
		Short: "Broadcast message add-quota",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argName := args[0]
			argMaxPercentSend, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			argMaxPercentRecv, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			argDurationMinutes, err := strconv.Atoi(args[3])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgAddQuota(
				clientCtx.GetFromAddress().String(),
				argName,
				uint64(argMaxPercentSend),
				uint64(argMaxPercentRecv),
				uint64(argDurationMinutes),
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
