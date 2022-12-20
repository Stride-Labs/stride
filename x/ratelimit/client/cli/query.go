package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	// Group ratelimit queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryRateLimit(),
		GetCmdQueryRateLimits(),
		GetCmdQueryRateLimitsByChainId(),
	)
	return cmd
}

// GetCmdQueryRateLimit implements a command to query a specific rate limit
func GetCmdQueryRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit [denom] [channel-id]",
		Short: "Query a specific rate limit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			denom := args[0]
			channelId := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRateLimitRequest{
				Denom:     denom,
				ChannelId: channelId,
			}
			res, err := queryClient.RateLimit(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.RateLimit)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRateLimits return all available rate limits.
func GetCmdQueryRateLimits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-rate-limits",
		Short: "Query all rate limits",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRateLimitsRequest{}
			res, err := queryClient.RateLimits(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(res.RateLimits)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRateLimits return all rate limits that exist between Stride
// and the specified ChainId
func GetCmdQueryRateLimitsByChainId() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-rate-limits [chain-id]",
		Short: "Query all rate limits with the given ChainID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainId := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRateLimitsByChainIdRequest{
				ChainId: chainId,
			}
			res, err := queryClient.RateLimitByChainId(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(res.RateLimits)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
