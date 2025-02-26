package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/Stride-Labs/stride/v25/x/strdburner/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryStrdBurnerAddress(),
		CmdQueryStrdBurnerTotalBurned(),
	)

	return cmd
}

func CmdQueryStrdBurnerAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "Query the address of the stride burner module",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryStrdBurnerAddressRequest{}
			res, err := queryClient.StrdBurnerAddress(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdQueryStrdBurnerTotalBurned() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-burned",
		Short: "Query the total amount of STRD the was burned using x/strdburner",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryTotalStrdBurnedRequest{}
			res, err := queryClient.TotalStrdBurned(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}
