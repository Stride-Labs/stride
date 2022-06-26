package cli

import (
	"context"
	"strconv"

	"github.com/Stride-Labs/stride/x/records/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdListUserRedemptionRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-user-redemption-record",
		Short: "list all userRedemptionRecord",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllUserRedemptionRecordRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.UserRedemptionRecordAll(context.Background(), params)
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

func CmdShowUserRedemptionRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-user-redemption-record [id]",
		Short: "shows a userRedemptionRecord",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			params := &types.QueryGetUserRedemptionRecordRequest{
				Id: id,
			}

			res, err := queryClient.UserRedemptionRecord(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
