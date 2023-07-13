package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
)

func CmdWithdrawVotingStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-voting-stake [amount] [denom]",
		Short: "Withdraw staked tokens to Stride for voting",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount, _ := sdk.NewIntFromString(args[0])
			denom := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawVotingStake(
				clientCtx.GetFromAddress().String(),
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
