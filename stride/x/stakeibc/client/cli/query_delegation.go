package cli

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/Stride-Labs/cosmos-sdk/client"
	"github.com/Stride-Labs/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdShowDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-delegation",
		Short: "shows delegation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetDelegationRequest{}

			res, err := queryClient.Delegation(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
