package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

func CmdUpdateInnerRedemptionRateBounds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-redemption-rate-bounds [chainid] [min-bound] [max-bound]",
		Short: "Broadcast message set-redemption-rate-bounds",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]
			minInnerRedemptionRate := sdk.MustNewDecFromStr(args[1])
			maxInnerRedemptionRate := sdk.MustNewDecFromStr(args[2])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateInnerRedemptionRateBounds(
				clientCtx.GetFromAddress().String(),
				argChainId,
				minInnerRedemptionRate,
				maxInnerRedemptionRate,
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
