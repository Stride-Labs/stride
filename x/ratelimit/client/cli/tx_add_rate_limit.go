package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// Adds a new rate limit
func CmdAddRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-rate-limit [denom] [channel-id] [max-percent-send] [max-percent-recv] [duration-minutes]",
		Short: "Broadcast message add-rate-limit",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argDenom := args[0]
			argChannelId := args[1]

			argMaxPercentSend, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			argMaxPercentRecv, err := strconv.Atoi(args[3])
			if err != nil {
				return err
			}

			argDurationMinutes, err := strconv.Atoi(args[4])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgAddRateLimit(
				clientCtx.GetFromAddress().String(),
				argDenom,
				argChannelId,
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
