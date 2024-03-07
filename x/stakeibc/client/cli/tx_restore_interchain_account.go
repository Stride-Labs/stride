package cli

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

func CmdRestoreInterchainAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-interchain-account [chain-id] [connection-id] [account-owner]",
		Short: "Broadcast message restore-interchain-account",
		Long: strings.TrimSpace(
			`Restores a closed channel associated with an interchain account.
Specify the chain ID and account owner - where the owner is the alias for the ICA account

For host zone ICA accounts, the owner is of the form {chainId}.{accountType}
ex:
>>> strided tx restore-interchain-account cosmoshub-4 connection-0 cosmoshub-4.DELEGATION 

For trade route ICA accounts, the owner is of the form:
    {chainId}.{rewardDenom}-{hostDenom}.{accountType}
ex:
>>> strided tx restore-interchain-account dydx-mainnet-1 connection-1 dydx-mainnet-1.uusdc-udydx.CONVERTER_TRADE 
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			chainId := args[0]
			connectionId := args[1]
			accountOwner := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRestoreInterchainAccount(
				clientCtx.GetFromAddress().String(),
				chainId,
				connectionId,
				accountOwner,
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
