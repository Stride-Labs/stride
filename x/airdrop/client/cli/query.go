package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
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
		Use:   "airdrop",
		Short: "Queries an airdrop's configuration",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries an airdrop configuration
Example:
  $ %s query %s airdrop [airdrop-id]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
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
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all airdrop configurations
Example:
  $ %s query %s airdrops
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
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
		Use:   "user-allocation",
		Short: "Queries a user allocation for a given airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries a user allocation for a given airdrop
Example:
  $ %s query %s user-allocation [airdrop-id] [user-address]
`, version.AppName, types.ModuleName),
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
		Use:   "user-allocations",
		Short: "Queries user allocations across all airdrops",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries user allocations across all airdrops
Example:
  $ %s query %s user-allocation [user-address]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
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
		Use:   "all-allocations",
		Short: "Queries all allocations for a given airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all allocations for a given airdrop
Example:
  $ %s query %s all-allocations [airdrop-id]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
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

// Queries the claim type of an address for an airdrop (daily claim, claim & stake, early),
// and the amount claimed and remaining
func CmdQueryUserSummary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-summary",
		Short: "Queries the summary for a user",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries a user's claimed and remaining reward amounts, as well as their
claim status (daily, claim & stake, claim early)

Example:
  $ %s query %s user-summary [address]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryUserSummaryRequest{
				Address: address,
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
