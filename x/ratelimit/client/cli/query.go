package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Stride-Labs/stride/v9/x/ratelimit/types"
)

const (
	FlagDenom = "denom"
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
		GetCmdQueryAllRateLimits(),
		GetCmdQueryRateLimitsByChainId(),
	)
	return cmd
}

// GetCmdQueryRateLimit implements a command to query rate limits by channel-id and denom
func GetCmdQueryRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit [channel-id]",
		Short: "Query rate limits by channel-id and denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query rate limits by channel-id and denom.

Example:
  $ %s query %s rate-limit [channel-id]
  $ %s query %s rate-limit [channel-id] --denom=[denom]
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelId := args[0]
			denom, err := cmd.Flags().GetString(FlagDenom)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			if denom == "" {
				req := &types.QueryRateLimitsByChannelIdRequest{
					ChannelId: channelId,
				}
				res, err := queryClient.RateLimitsByChannelId(context.Background(), req)
				if err != nil {
					return err
				}

				return clientCtx.PrintObjectLegacy(res.RateLimits)
			}

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

	cmd.Flags().String(FlagDenom, "", "The denom identifying a specific rate limit")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAllRateLimits return all available rate limits.
func GetCmdQueryAllRateLimits() *cobra.Command {
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

			req := &types.QueryAllRateLimitsRequest{}
			res, err := queryClient.AllRateLimits(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(res.RateLimits)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRateLimitsByChainId return all rate limits that exist between Stride
// and the specified ChainId
func GetCmdQueryRateLimitsByChainId() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limits-by-chain [chain-id]",
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
			res, err := queryClient.RateLimitsByChainId(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(res.RateLimits)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
