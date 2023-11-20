package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// host-denom and reward-denom are to find which trade route should be updated
// if a match is found, then pool-id and the swap amounts will update on that route
func CmdUpdateTradeRoute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-trade-route [host-denom] [reward-denom] [pool-id] [min-swap-amount (optional)] [max-swap-amount (optional)]",
		Short: "Broadcast message update-trade-route",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostDenom := args[0]
			rewardDenom := args[1]
			poolId, err := strconv.ParseUint(args[2], 10, 64)
			minSwapAmount, found := sdk.NewIntFromString(args[3])
			if !found {
				minSwapAmount = sdk.ZeroInt()
			}
			maxSwapAmount, found := sdk.NewIntFromString(args[4])
			if !found {
				const MaxUint = ^uint(0)
				const MaxInt = int64(MaxUint >> 1) 
				maxSwapAmount = sdk.NewInt(MaxInt)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateTradeRoute(
				clientCtx.GetFromAddress().String(),
				hostDenom,
				rewardDenom,
				poolId,
				minSwapAmount,
				maxSwapAmount,
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
