package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Stride-Labs/stride/v4/x/interchainquery/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	// Group lockup queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	return cmd
}

// GetCmdQueries provides a list of all query objects
func GetCmdListQueries() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-queries",
		Short: "Query all queries",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query interchainquery list-queries`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryServiceClient(clientCtx)

			req := &types.QueryAllQueriesRequest{}

			res, err := queryClient.AllQueries(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueries provides a list of all pending queries
// (queries that have not have been requested but have not received a response)
func GetCmdListPendingQueries() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-pending-queries",
		Short: "Query all pending queries",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query interchainquery list-pending-queries`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryServiceClient(clientCtx)

			req := &types.QueryPendingQueriesRequest{}

			res, err := queryClient.PendingQueries(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
