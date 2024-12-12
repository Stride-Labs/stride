package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v24/x/auction/types"
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
		CmdPlaceBid(),
		CmdCreateAuction(),
		CmdUpdateAuction(),
	)

	return cmd
}

func CmdPlaceBid() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-bid [utokenAmount] [ustrdAmount]",
		Short: "Place a bid on an auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a bid on an auction for a specific token.

Example:
  $ %[1]s tx %[2]s place-bid 123ibc/DEADBEEF 1000000 --from mykey
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("cannot parse token amount and denom from '%s': %w", args[0], err)
			}

			ustrdAmount, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse ustrdAmount as uint64 from '%s': %w", args[2], err)
			}

			msg := types.NewMsgPlaceBid(
				clientCtx.GetFromAddress().String(),
				coin.Denom,
				coin.Amount.Uint64(),
				ustrdAmount,
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

func CmdCreateAuction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-auction [denom] [enabled] [price-multiplier] [min-bid-amount] [beneficiary]",
		Short: "Create a new auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new auction for a specific token.

Example:
  $ %[1]s tx %[2]s create-auction ibc/DEADBEEF true 0.95 1000000 --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			enabled := args[1] == "true"

			minBidAmount, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse minBidAmount as uint64 from '%s': %w", args[3], err)
			}

			msg := types.NewMsgCreateAuction(
				clientCtx.GetFromAddress().String(),
				types.AuctionType_AUCTION_TYPE_FCFS, // auction type
				args[0],                             // denom
				enabled,                             // enabled
				args[2],                             // price multiplier
				minBidAmount,                        // min bid amount
				args[4],                             // beneficiary
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

func CmdUpdateAuction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-auction [denom] [enabled] [price-multiplier] [min-bid-amount] [beneficiary]",
		Short: "Update an existing auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update an existing auction's parameters.

Example:
  $ %[1]s tx %[2]s update-auction ibc/DEADBEEF true 0.97 500000 --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			enabled := args[1] == "true"

			minBidAmount, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse minBidAmount as uint64 from '%s': %w", args[3], err)
			}

			msg := types.NewMsgUpdateAuction(
				clientCtx.GetFromAddress().String(),
				types.AuctionType_AUCTION_TYPE_FCFS, // auction type
				args[0],                             // denom
				enabled,                             // enabled
				args[2],                             // price multiplier
				minBidAmount,                        // min bid amount
				args[4],                             // beneficiary

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
