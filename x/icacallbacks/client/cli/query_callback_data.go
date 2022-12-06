package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
)

func CmdListCallbackData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-callback-data",
		Short: "list all callback-data",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllCallbackDataRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.CallbackDataAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowCallbackData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-callback-data [callback-key]",
		Short: "shows a callback-data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argCallbackKey := args[0]

			params := &types.QueryGetCallbackDataRequest{
				CallbackKey: argCallbackKey,
			}

			res, err := queryClient.CallbackData(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
