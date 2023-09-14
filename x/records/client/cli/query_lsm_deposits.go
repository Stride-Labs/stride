package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v14/x/records/types"
)

const (
	FlagHostChainId      = "host-chain-id"
	FlagValidatorAddress = "validator"
	FlagStatus           = "status"
)

func CmdLSMDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsm-deposit [chain-id] [denom]",
		Short: "shows an LSM deposit matching given denom and chain-id",
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
		Use:   "lsm-deposits",
		Short: "shows all lsm-deposits matching optional filters",
		Long: `Shows all LSM deposits with optional filters
Examples:
  $ lsm-deposits
  $ lsm-deposits --host-chain-id=[chain-id]
  $ lsm-deposits --host-chain-id=[chain-id] validator=[validator-address]
  $ lsm-deposits --host-chain-id=[chain-id] --status=[status]
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainId, err := cmd.Flags().GetString(FlagHostChainId)
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

	cmd.Flags().String(FlagHostChainId, "", "The chainId for host zone")
	cmd.Flags().String(FlagValidatorAddress, "", "The validator address")
	cmd.Flags().String(FlagStatus, "", "The status")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
