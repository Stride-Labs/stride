package cli

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var _ = strconv.Itoa(0)

func CmdRestoreInterchainAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-interchain-account [chain-id] [account-type]",
		Short: "Broadcast message restore-interchain-account",
		Args:  cobra.ExactArgs(2),
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
