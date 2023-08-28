package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func CmdRestoreInterchainAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-interchain-account [chain-id] [account-type]",
		Short: "Broadcast message restore-interchain-account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Restores a closed channel associated with an interchain account.
Specify the interchain account type as either: %s, %s, %s, or %s`,
				types.ICAAccountType_DELEGATION,
				types.ICAAccountType_WITHDRAWAL,
				types.ICAAccountType_REDEMPTION,
				types.ICAAccountType_FEE)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]
			argAccountType := args[1]

			accountType, found := types.ICAAccountType_value[argAccountType]
			if !found {
				return errors.New("Invalid account type.")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRestoreInterchainAccount(
				clientCtx.GetFromAddress().String(),
				argChainId,
				types.ICAAccountType(accountType),
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
