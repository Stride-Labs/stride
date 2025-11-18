package cli

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v30/x/strdburner/types"
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
		CmdBurn(),
		CmdLink(),
	)

	return cmd
}

// Burn STRD
func CmdBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn [amount]",
		Short: "Burn STRD",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Burn STRD tokens

Example:
  $ %[1]s tx %[2]s burn 10000
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, ok := sdkmath.NewIntFromString(args[0])
			if !ok {
				return errors.New("unable to parse amount")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgBurn(
				clientCtx.GetFromAddress().String(),
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

// Link non-stride address
func CmdLink() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [linked-address]",
		Short: "Link non-stride adress",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Link non-stride adress
WARNING: No address validation is done in the CLI, it is YOUR responsibility to enter a correct address

Example:
  $ %[1]s tx %[2]s link 0x1234...
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			linkedAddress := args[0]
			msg := types.NewMsgLink(
				clientCtx.GetFromAddress().String(),
				linkedAddress,
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
