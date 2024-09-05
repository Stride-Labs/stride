package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/Stride-Labs/stride/v24/x/airdrop/types"
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
		CmdQueryAirdrop(),
		CmdQueryAllAirdrops(),
		CmdQueryUserAllocation(),
		CmdQueryUserAllocations(),
		CmdQueryAllAllocations(),
		CmdQueryUserSummary(),
	)

	return cmd
}

// Queries the configuration for a given airdrop
func CmdQueryAirdrop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airdrop [airdrop-id]",
		Short: "Queries an airdrop's configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAirdropRequest{
				Id: airdropId,
			}
			res, err := queryClient.Airdrop(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// Queries all airdrop configurations
func CmdQueryAllAirdrops() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airdrops",
		Short: "Queries all airdrop configurations",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllAirdropsRequest{}
			res, err := queryClient.AllAirdrops(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// Queries the allocation for a given user for an airdrop
func CmdQueryUserAllocation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-allocation [airdrop-id] [user-address]",
		Short: "Queries a user's allocation for a given airdrop",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]
			address := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryUserAllocationRequest{
				AirdropId: airdropId,
				Address:   address,
			}
			res, err := queryClient.UserAllocation(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// Queries the allocations for a given user across all airdrops
func CmdQueryUserAllocations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-allocations [user-address]",
		Short: "Queries a user's allocations across all airdrops",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryUserAllocationsRequest{
				Address: address,
			}
			res, err := queryClient.UserAllocations(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// Queries all allocations for a given airdrop
func CmdQueryAllAllocations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-allocations [airdrop-id]",
		Short: "Queries all allocations for a given airdrop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllAllocationsRequest{
				AirdropId: airdropId,
			}
			res, err := queryClient.AllAllocations(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// Queries the claim type of an address for an airdrop (daily claim or claim early),
// and the amount claimed and remaining
func CmdQueryUserSummary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-summary [airdrop-id] [user-address]",
		Short: "Queries the summary for a user",
		Long: strings.TrimSpace(
			`Queries airdrop summary info for a user including their total claimed, amount remaining, and claim type`,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]
			address := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryUserSummaryRequest{
				AirdropId: airdropId,
				Address:   address,
			}
			res, err := queryClient.UserSummary(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}
