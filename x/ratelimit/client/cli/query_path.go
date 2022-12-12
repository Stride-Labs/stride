package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// GetCmdQueryPath implements a command to query a specific path
func GetCmdQueryPath() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path [path-id]",
		Short: "Query the path by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathId := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPathRequest{
				Id: pathId,
			}
			res, err := queryClient.Path(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Path)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
