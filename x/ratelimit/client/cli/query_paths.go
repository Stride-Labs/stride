package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// GetCmdQueryPaths return all available paths
func GetCmdQueryPaths() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paths",
		Short: "Query all paths",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPathsRequest{}
			res, err := queryClient.Paths(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(res.Paths)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
