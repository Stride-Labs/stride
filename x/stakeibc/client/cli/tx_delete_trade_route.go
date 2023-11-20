package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// The host-denom and reward-denom are not ibc denoms
// these are the native denoms as they appear on their own chains
func CmdDeleteTradeRoute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-trade-route [host-denom] [reward-denom]",
		Short: "Broadcast message delete-trade-route",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostDenom := args[0]
			rewardDenom := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteTradeRoute(
				clientCtx.GetFromAddress().String(),
				hostDenom,
				rewardDenom,
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
