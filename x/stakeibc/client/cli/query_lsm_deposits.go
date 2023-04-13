package cli

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func CmdLSMDepositsHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-lsm-deposits-host-zone [chain-id]",
		Short: "shows all lsm deposits with the given chain-id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId := args[0]

			params := &types.QueryLSMDepositsRequest{ChainId: chainId}

			res, err := queryClient.LSMDeposits(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdLSMDepositsWithStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-lsm-deposits-with-status [chain-id] [status]",
		Short: "shows all lsm deposits which match the given chain-id and status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId := args[0]
			statusStr := args[1]

			// convert the incoming status string to proper LSMDepositStatus type
			// and then convert LSMDepositStatus to the XStatus type for optional params
			statusInt32, exists := types.LSMDepositStatus_value[statusStr]
			if !exists {
				return errors.New("Unable to recognize requested status!")
			}

			status := types.LSMDepositStatus(statusInt32)
			xStatus := &types.QueryLSMDepositsRequest_Status{Status: status}
			params := &types.QueryLSMDepositsRequest{ChainId: chainId, XStatus: xStatus}
			res, err := queryClient.LSMDeposits(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdLSMDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-lsm-deposit [chain-id] [denom]",
		Short: "shows either 0 or 1 deposits which match given denom on chain-id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId := args[0]
			denomStr := args[1]

			// convert the incoming denom string to the XDenom type for optional params
			xDenom := &types.QueryLSMDepositsRequest_Denom{Denom: denomStr}
			params := &types.QueryLSMDepositsRequest{ChainId: chainId, XDenom: xDenom}
			res, err := queryClient.LSMDeposits(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
