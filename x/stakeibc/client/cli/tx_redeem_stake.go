package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

var _ = strconv.Itoa(0)

func CmdRedeemStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem-stake [amount] [hostZoneID] [receiver]",
		Short: "Broadcast message redeem-stake",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			argAmount, found := sdk.NewIntFromString(args[0])
			if !found {
				return fmt.Errorf("can not convert string to int: %s", types.ErrInvalidType.Error())
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
