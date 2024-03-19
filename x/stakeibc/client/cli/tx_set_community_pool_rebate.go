package cli

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

func CmdSetCommunityPoolRebate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-rebate [chain-id] [rebate-rate] [liquid-staked-amount]",
		Short: "Registers or updates a community pool rebate",
		Long: strings.TrimSpace(`Registers a community pool rebate by specifying the rebate percentage (as a decimal)
and the amount liquid staked. 
E.g. to specify a 20% rebate, the rebate rate should be 0.2

If a 0.0 rebate or 0 token liquid stake is specified, the rebate will be deleted.
		`),
		Args: cobra.ExactArgs(3),
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

			msg := types.NewMsgSetCommunityPoolRebate(
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
