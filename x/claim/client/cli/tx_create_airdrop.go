package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

func CmdCreateAirdrop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-airdrop [identifier] [start] [duration] [denom]",
		Short: "Broadcast message create-airdrop",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argStartTime, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			argDuration, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgCreateAirdrop(
				clientCtx.GetFromAddress().String(),
				args[0],
				uint64(argStartTime),
				uint64(argDuration),
				args[3],
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
