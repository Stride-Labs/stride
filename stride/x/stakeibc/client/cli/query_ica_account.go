package cli

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/Stride-Labs/cosmos-sdk/client"
	"github.com/Stride-Labs/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdShowICAAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-ica-account",
		Short: "shows ICAAccount",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetICAAccountRequest{}

			res, err := queryClient.ICAAccount(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
