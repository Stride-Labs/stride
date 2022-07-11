package cli

import (
    "context"
	
    "github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
    "github.com/Stride-Labs/stride/x/stakeibc/types"
)

func CmdListPendingClaims() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-pending-claims",
		Short: "list all pending-claims",
		RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx := client.GetClientContextFromCmd(cmd)

            pageReq, err := client.ReadPageRequest(cmd.Flags())
            if err != nil {
                return err
            }

            queryClient := types.NewQueryClient(clientCtx)

            params := &types.QueryAllPendingClaimsRequest{
                Pagination: pageReq,
            }

            res, err := queryClient.PendingClaimsAll(context.Background(), params)
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

func CmdShowPendingClaims() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-pending-claims [sequence]",
		Short: "shows a pending-claims",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
            clientCtx := client.GetClientContextFromCmd(cmd)

            queryClient := types.NewQueryClient(clientCtx)

             argSequence := args[0]
            
            params := &types.QueryGetPendingClaimsRequest{
                Sequence: argSequence,
                
            }

            res, err := queryClient.PendingClaims(context.Background(), params)
            if err != nil {
                return err
            }

            return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

    return cmd
}
