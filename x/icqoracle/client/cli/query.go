package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/Stride-Labs/stride/v27/x/icqoracle/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryTokenPrice(),
		CmdQueryTokenPrices(),
		CmdQueryParams(),
	)

	return cmd
}

func CmdQueryTokenPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-price [base-denom] [quote-denom] [pool-id]",
		Short: "Query the current price for a specific token",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseDenom := args[0]
			quoteDenom := args[1]
			poolId, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("Error parsing osmosis pool ID as uint64: %w", err)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryTokenPriceRequest{
				BaseDenom:  baseDenom,
				QuoteDenom: quoteDenom,
				PoolId:     poolId,
			}
			res, err := queryClient.TokenPrice(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdQueryTokenPrices() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-prices",
		Short: "Query all token prices",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryTokenPricesRequest{}
			res, err := queryClient.TokenPrices(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Get the parameters",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdQueryTokenPriceForQuoteDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-price-by-quote [base-denom] [quote-denom]",
		Short: "Query the current price for a specific token",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseDenom := args[0]
			quoteDenom := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryTokenPriceForQuoteDenomRequest{
				BaseDenom:  baseDenom,
				QuoteDenom: quoteDenom,
			}
			res, err := queryClient.TokenPriceForQuoteDenom(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}
