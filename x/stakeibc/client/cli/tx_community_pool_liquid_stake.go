package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func CmdCommunityPoolLiquidStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool-liquid-stake [community-pool-host-zone] [amount] [token-denom]",
		Short: "IBC transfer and liquid stake tokens from community pool host zone",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			communityPoolHostZoneId := args[0]
			amount, found := sdk.NewIntFromString(args[1])
			if !found {
				return errorsmod.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
			}
			denom := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCommunityPoolLiquidStake(
				clientCtx.GetFromAddress().String(),				
				communityPoolHostZoneId,
				amount,
				denom,
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
