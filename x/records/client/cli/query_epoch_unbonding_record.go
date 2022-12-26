package cli

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/records/types"
)

func CmdListEpochUnbondingRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-epoch-unbonding-record",
		Short: "list all EpochUnbondingRecord",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllEpochUnbondingRecordRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.EpochUnbondingRecordAll(context.Background(), params)
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

func CmdShowEpochUnbondingRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-epoch-unbonding-record [id]",
		Short: "shows a EpochUnbondingRecord",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			epochNumber, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			params := &types.QueryGetEpochUnbondingRecordRequest{
				EpochNumber: epochNumber,
			}

			res, err := queryClient.EpochUnbondingRecord(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
