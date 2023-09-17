package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func CmdUpdateInnerRedemptionRateBounds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-tight-bounds [chainid] [min-bound] [max-bound]",
		Short: "Broadcast message update-tight-bounds",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]
			minRedemptionRateStr := args[1]
			maxRedemptionRateStr := args[2]

			minInnerRedemptionRate := sdk.ZeroDec()
			if minRedemptionRateStr != "" {
				minInnerRedemptionRate, err = sdk.NewDecFromStr(minRedemptionRateStr)
				if err != nil {
					return err
				}
			}
			maxInnerRedemptionRate := sdk.ZeroDec()
			if maxRedemptionRateStr != "" {
				maxInnerRedemptionRate, err = sdk.NewDecFromStr(maxRedemptionRateStr)
				if err != nil {
					return err
				}
			}

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
