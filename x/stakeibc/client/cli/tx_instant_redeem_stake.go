package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func CmdInstantRedeemStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instant-redeem-stake [amount] [hostZoneID]",
		Short: "Broadcast message instant-redeem-stake",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argAmount, found := sdk.NewIntFromString(args[0])
			if !found {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
			}
			hostZoneID := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgInstantRedeemStake(
				clientCtx.GetFromAddress().String(),
				argAmount,
				hostZoneID,
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
