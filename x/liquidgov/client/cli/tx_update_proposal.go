package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
)

func CmdUpdateProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-proposal [chain-id] [proposal-id]",
		Short: "Trigger ICQ to the given chain-id for the proposal info",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			host_zone_id := args[0]
			proposal_id, _ := strconv.ParseUint(args[1], 10, 64)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateProposal(
				clientCtx.GetFromAddress().String(),
				host_zone_id,
				proposal_id,
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
