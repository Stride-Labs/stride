package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var _ = strconv.Itoa(0)

func CmdRedeemStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem-stake [amount] [hostZoneID] [receiver]",
		Short: "Broadcast message redeem-stake",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			argAmount, err := cast.ToUint64E(args[0])
			if err != nil {
				return err
			}
			hostZoneID := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			argReceiver := args[2]

			msg := types.NewMsgRedeemStake(
				clientCtx.GetFromAddress().String(),
				argAmount,
				hostZoneID,
				argReceiver,
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
