package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/Stride-Labs/stride/v26/x/auction/types"
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
		CmdQueryAuction(),
		CmdQueryAuctions(),
	)

	return cmd
}

func CmdQueryAuction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auction [name]",
		Short: "Query the auction info for a specific auction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAuctionRequest{
				Name: args[0],
			}
			res, err := queryClient.Auction(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}

func CmdQueryAuctions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auctions",
		Short: "Get all auctions",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAuctionsRequest{}
			res, err := queryClient.Auctions(context.Background(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	return cmd
}
