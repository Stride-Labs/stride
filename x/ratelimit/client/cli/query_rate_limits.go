package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// GetCmdQueryRateLimits return all available rate limits.
func GetCmdQueryRateLimits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limits",
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
