package cli

import (
	"fmt"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v28/x/auction/types"
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
		Use:   "place-bid [auction-name] [selling-token-amount] [payment-token-amount]",
		Short: "Place a bid on an auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place a bid on an auction for a specific token.

Example:
  $ %[1]s tx %[2]s place-bid auctionName 123 1000000 --from mykey
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sellingTokenAmount, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("cannot parse sellingTokenAmount as sdkmath.Int from '%s'", args[2])
			}

			paymentTokenAmount, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("cannot parse paymentTokenAmount as sdkmath.Int from '%s'", args[2])
			}

			msg := types.NewMsgPlaceBid(
				clientCtx.GetFromAddress().String(),
				args[0],
				sellingTokenAmount,
				paymentTokenAmount,
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
		Use:   "create-auction [name] [selling-denom] [payment-denom] [enabled] [min-price-multiplier] [min-bid-amount] [beneficiary]",
		Short: "Create a new auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new auction for a specific token.

Example:
  $ %[1]s tx %[2]s create-auction my-auction ibc/DEADBEEF ustrd true 0.95 1000000 strideXXX --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			enabled := args[3] == "true"

			minBidAmount, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse minBidAmount as uint64 from '%s': %w", args[5], err)
			}

			msg := types.NewMsgCreateAuction(
				clientCtx.GetFromAddress().String(),
				args[0],
				types.AuctionType_AUCTION_TYPE_FCFS,
				args[1],
				args[2],
				enabled,
				args[4],
				minBidAmount,
				args[6],
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
		Use:   "update-auction [name] [enabled] [min-price-multiplier] [min-bid-amount] [beneficiary]",
		Short: "Update an existing auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update an existing auction's parameters.

Example:
  $ %[1]s tx %[2]s update-auction auctionName true 0.97 500000 strideXXX --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			enabled, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("cannot parse enabled as bool from '%s': %w", args[1], err)
			}

			minBidAmount, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("cannot parse minBidAmount as uint64 from '%s': %w", args[3], err)
			}

			msg := types.NewMsgUpdateAuction(
				clientCtx.GetFromAddress().String(),
				args[0],
				types.AuctionType_AUCTION_TYPE_FCFS,
				enabled,
				args[2],
				minBidAmount,
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
