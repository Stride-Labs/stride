package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
)

func CmdDepositVotingStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-voting-stake [stdenom] [amount]",
		Short: "Deposit staked tokens to Stride for voting",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			stdenom := args[0]
			amount, _ := sdk.NewIntFromString(args[1])
			
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDepositVotingStake(
				clientCtx.GetFromAddress().String(),
				stdenom,
				amount,
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
