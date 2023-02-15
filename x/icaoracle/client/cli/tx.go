package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
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
		CmdAddOracle(),
		CmdRestoreOracleICA(),
	)

	return cmd
}

// Adds a new oracle given a provided connection and cosmwasm contract
func CmdAddOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle [moniker] [connection-id] [contract-code-id]",
		Short: "Adds an oracle as a destination for metric updates",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Registers a new cosmwasm oracle to mark it as a destination for metric updates.
Must provide the ID of an existing connection and a cosmwasm contract that has been uploaded to the host chain.

Example:
  $ %[1]s tx %[2]s add-oracle osmosis connection-0 10
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			moniker := args[0]
			connectionId := args[1]
			contractCodeId, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddOracle(
				clientCtx.GetFromAddress().String(),
				moniker,
				connectionId,
				contractCodeId,
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

// Restores the oracle ICA channel after a channel closure
func CmdRestoreOracleICA() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-oracle-ica [oracle-moniker]",
		Short: "Restores an oracle ICA channel",
		Long: strings.TrimSpace(
			fmt.Sprintf(`After a channel closure, creates a new oracle ICA channel and restores the ICA account 

Example:
  $ %[1]s tx %[2]s restore-oracle-ica osmosis
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moniker := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRestoreOracleICA(
				clientCtx.GetFromAddress().String(),
				moniker,
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
