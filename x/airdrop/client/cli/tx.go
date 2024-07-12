package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
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

	cmd.AddCommand()

	return cmd
}

// User transaction to claim all the pending airdrop rewards up to the current day
func CmdClaimDaily() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-daily [airdrop-id]",
		Short: "Claims all the pending airdrop rewards up to the current day",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claims all pending airdrop rewards up to the current day. 
This option is only available if the user has not already elected to claim and stake or claim early

Example:
  $ %[1]s tx %[2]s claim-daily airdrop-1 --from user
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimDaily(
				clientCtx.GetFromAddress().String(),
				airdropId,
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
