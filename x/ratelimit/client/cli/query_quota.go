package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// GetCmdQueryQuota implements a command to return the quota by name.
func GetCmdQueryQuota() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota [name]",
		Short: "Query the quota by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			argName := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			quotaReq := &types.QueryQuotaRequest{
				Name: argName,
			}
			res, err := queryClient.Quota(context.Background(), quotaReq)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Quota)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
