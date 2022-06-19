package cli

import (
    "context"
	
    "github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
    "github.com/Stride-Labs/stride/x/stakeibc/types"
)

func CmdListControllerBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-controller-balances",
		Short: "list all controllerBalances",
		RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx := client.GetClientContextFromCmd(cmd)

            pageReq, err := client.ReadPageRequest(cmd.Flags())
            if err != nil {
                return err
            }

            queryClient := types.NewQueryClient(clientCtx)

            params := &types.QueryAllControllerBalancesRequest{
                Pagination: pageReq,
            }

            res, err := queryClient.ControllerBalancesAll(context.Background(), params)
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

func CmdShowControllerBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-controller-balances [index]",
		Short: "shows a controllerBalances",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
            clientCtx := client.GetClientContextFromCmd(cmd)

            queryClient := types.NewQueryClient(clientCtx)

             argIndex := args[0]
            
            params := &types.QueryGetControllerBalancesRequest{
                Index: argIndex,
                
            }

            res, err := queryClient.ControllerBalances(context.Background(), params)
            if err != nil {
                return err
            }

            return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

    return cmd
}
