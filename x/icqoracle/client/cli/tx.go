package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
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
		CmdAddTokenPrice(),
		CmdRemoveTokenPrice(),
	)

	return cmd
}

func CmdAddTokenPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-token-price [base-denom] [quote-denom] [osmosis-pool-id] [osmosis-base-denom] [osmosis-quote-denom]",
		Short: "Add a token to price tracking",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add a token to price tracking.

Example:
  $ %[1]s tx %[2]s add-token-price uosmo uatom 123 uosmo ibc/... --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterTokenPriceQuery(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				args[2],
				args[3],
				args[4],
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

func CmdRemoveTokenPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-token-price [base-denom] [quote-denom] [osmosis-pool-id]",
		Short: "Remove a token from price tracking",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove a token from price tracking.

Example:
  $ %[1]s tx %[2]s remove-token-price uatom uosmo 123 --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRemoveTokenPriceQuery(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				args[2],
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
