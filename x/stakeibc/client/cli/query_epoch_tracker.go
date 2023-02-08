package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func CmdListEpochTracker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-epoch-tracker",
		Short: "list all epoch-tracker",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllEpochTrackerRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.EpochTrackerAll(context.Background(), params)
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

func CmdShowEpochTracker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-epoch-tracker [epoch-identifier]",
		Short: "shows a epoch-tracker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argEpochIdentifier := args[0]

			params := &types.QueryGetEpochTrackerRequest{
				EpochIdentifier: argEpochIdentifier,
			}

			res, err := queryClient.EpochTracker(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
