package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// GetCmdQueryRateLimit implements a command to query a specific rate limit
func GetCmdQueryRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit [path-id]",
		Short: "Query the rate limit by path id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathId := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRateLimitRequest{
				PathId: pathId,
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
