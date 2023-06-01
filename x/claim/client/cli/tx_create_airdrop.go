package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v9/x/claim/types"
)

func CmdCreateAirdrop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-airdrop [identifier] [chain-id] [denom] [start] [duration] [autopilot-enabled]",
		Short: "Broadcast message create-airdrop",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			identifier := args[0]
			chainId := args[1]
			denom := args[2]
			argStartTime, err := strconv.Atoi(args[3])
			if err != nil {
				return err
			}
			argDuration, err := strconv.Atoi(args[4])
			if err != nil {
				return err
			}
			autopilotEnabled, err := strconv.ParseBool(args[5])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			distributor := clientCtx.GetFromAddress().String()
			msg := types.NewMsgCreateAirdrop(
				distributor,
				identifier,
				chainId,
				denom,
				uint64(argStartTime),
				uint64(argDuration),
				autopilotEnabled,
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
