package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v10/x/stakeibc/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
				return errorsmod.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
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
