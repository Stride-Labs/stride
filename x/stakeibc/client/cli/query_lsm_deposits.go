package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

const (
	FlagChainId          = "host-zone"
	FlagValidatorAddress = "validator"
	FlagStatus           = "status"
)

func CmdLSMDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-lsm-deposit [chain-id] [denom]",
		Short: "shows either 0 or 1 deposits which match given denom on chain-id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId := args[0]
			denom := args[1]

			params := &types.QueryLSMDepositRequest{ChainId: chainId, Denom: denom}
			res, err := queryClient.LSMDeposit(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdLSMDeposits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-lsm-deposits --host-zone=[chain_id] --validator=[validator_address] --status=[status]",
		Short: "shows all lsm-deposits filtered by optional flags chain-id validate-address and status",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainId, err := cmd.Flags().GetString(FlagChainId)
			if err != nil {
				return err
			}
			validatorAddress, err := cmd.Flags().GetString(FlagValidatorAddress)
			if err != nil {
				return err
			}
			status, err := cmd.Flags().GetString(FlagStatus)
			if err != nil {
				return err
			}

			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryLSMDepositsRequest{ChainId: chainId, ValidatorAddress: validatorAddress, Status: status}
			res, err := queryClient.LSMDeposits(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagChainId, "", "The chainId for host zone")
	cmd.Flags().String(FlagValidatorAddress, "", "The validator address")
	cmd.Flags().String(FlagStatus, "", "The status")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
