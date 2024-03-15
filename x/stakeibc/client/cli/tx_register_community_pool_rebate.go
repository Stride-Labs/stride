package cli

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

func CmdRegisterCommunityPoolRebate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-rebate [chain-id] [rebate-percentage] [liquid-stake-amount]",
		Short: "Broadcast message register-rebate",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			chainId := args[0]
			rebatePercentage, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("unable to parse rebate percentage: %s", err.Error())
			}
			liquidStakeAmount, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return errors.New("unable to parse liquid stake amount")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterCommunityPoolRebate(
				clientCtx.GetFromAddress().String(),
				chainId,
				rebatePercentage,
				liquidStakeAmount,
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
