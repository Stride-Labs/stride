package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// GetCmdQueryQuota implements a command to return all available quotas.
func GetCmdQueryQuotas() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quotas",
		Short: "Query the quota by name",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			quotasReq := &types.QueryQuotasRequest{}
			res, err := queryClient.Quotas(context.Background(), quotasReq)

			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(res.Quotas)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
