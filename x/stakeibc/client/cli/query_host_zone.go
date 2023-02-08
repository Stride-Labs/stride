package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func CmdListHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-host-zone",
		Short: "list all HostZone",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllHostZoneRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.HostZoneAll(context.Background(), params)
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

func CmdShowHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-host-zone [chain-id]",
		Short: "shows a HostZone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId := args[0]

			params := &types.QueryGetHostZoneRequest{
				ChainId: chainId,
			}

			res, err := queryClient.HostZone(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
