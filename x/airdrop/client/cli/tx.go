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

	cmd.AddCommand(
		CmdClaimDaily(),
		CmdClaimEarly(),
		CmdClaimAndStake(),
	)

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

// User transaction to claim half of their total amount now, and forfeit the other half to be clawed back
func CmdClaimEarly() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-early [airdrop-id]",
		Short: "Claims rewards immediately, but with a early claim penalty",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claims rewards immediately (including for future days), but with an early
claim penalty causing a portion of the total to be clawed back.

Example:
  $ %[1]s tx %[2]s claim-early airdrop-1 --from user
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimEarly(
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

// User transaction to claim and stake the full airdrop amount
// The rewards will be locked until the end of the distribution period, but will recieve rewards throughout this time
func CmdClaimAndStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-and-stake [airdrop-id] [validator-address]",
		Short: "Claims and stakes the total reward amount",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claims and stakes the full airdrop amount.
The rewards will be locked until the end of the distribution period, but will accrue rewards throughout this time

Example:
  $ %[1]s tx %[2]s claim-early airdrop-1 stridevaloperX --from user
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]
			vaildatorAddress := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimAndStake(
				clientCtx.GetFromAddress().String(),
				airdropId,
				vaildatorAddress,
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
