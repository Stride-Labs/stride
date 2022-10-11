package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/claim/types"
)

func CmdDepositAirdrop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-airdrop [airdrop-amount]",
		Short: "Broadcast message deposit-airdrop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAirdropAmount, ok := sdk.NewIntFromString(args[0])
			if !ok {
				return types.ErrFailedToParseDec
			}

			clientCtx, err := client.GetClientTxContext(cmd)

			msg := types.NewMsgDepositAirdrop(
				clientCtx.GetFromAddress().String(),
				sdk.NewCoins(sdk.NewCoin("ustrd", argAirdropAmount)),
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
