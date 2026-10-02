package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Stride-Labs/stride/v35/x/icaoracle/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRestoreOracleICA(),
	)

	return cmd
}

// Restores the oracle ICA channel after a channel closure
func CmdRestoreOracleICA() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-oracle-ica [oracle-chain-id]",
		Short: "Restores an oracle ICA channel",
		Long: strings.TrimSpace(
			fmt.Sprintf(`After a channel closure, creates a new oracle ICA channel and restores the ICA account 

Example:
  $ %[1]s tx %[2]s restore-oracle-ica osmosis
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRestoreOracleICA(
				clientCtx.GetFromAddress().String(),
				chainId,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
