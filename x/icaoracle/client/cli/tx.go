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

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
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
		CmdInstantiateOracle(),
		CmdRestoreOracleICA(),
	)

	return cmd
}

// Adds a new oracle given a provided connection and registers the oracle ICA
func CmdAddOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle [connection-id]",
		Short: "Adds an oracle as a destination for metric updates",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Registers a new oracle ICA as a destination for metric updates.
Must provide the ID of an existing connection.

Example:
  $ %[1]s tx %[2]s add-oracle connection-0 
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			connectionId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddOracle(
				clientCtx.GetFromAddress().String(),
				connectionId,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Instantiates an oracle cosmwasm contract
func CmdInstantiateOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instantiate-oracle [oracle-chain-id] [contract-code-id] [transfer-channel-on-oracle]",
		Short: "Instantiates an oracle cosmwasm contract",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submits an ICA to instantiate the oracle cosmwasm contract.
Must provide the codeID of a cosmwasm contract that has already been uploaded to the host chain, 
as well as the transfer channel ID as it lives on the oracle's chain. 

Example:
  $ %[1]s tx %[2]s instantiate-oracle osmosis-1 1000 channel-0
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainId := args[0]
			contractCodeId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			transferChannelOnOracle := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgInstantiateOracle(
				clientCtx.GetFromAddress().String(),
				chainId,
				contractCodeId,
				transferChannelOnOracle,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

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
