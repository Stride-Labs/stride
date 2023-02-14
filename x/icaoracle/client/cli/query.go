package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/gogo/protobuf/proto"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

const (
	FlagMetricKey     = "metric-key"
	FlagOracleMoniker = "oracle-moniker"
	FlagActive        = "active"
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
		GetCmdQueryOracle(),
		GetCmdQueryOracles(),
		GetCmdQueryPendingMetricUpdates(),
	)

	return cmd
}

// GetCmdQueryOracle implements a command to query a specific oracle using the oracle's moniker
func GetCmdQueryOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "Queries a specific oracle",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries a specific oracle using the oracle's moniker
Example:
  $ %s query %s oracle [moniker]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moniker := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOracleRequest{
				Moniker: moniker,
			}
			res, err := queryClient.Oracle(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// GetCmdQueryOracles implements a command to query all oracles with an optional "active" filter
func GetCmdQueryOracles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracles",
		Short: "Queries all oracles",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all oracles with an optional "active" filter
Examples:
  $ %[1]s query %[2]s oracles
  $ %[1]s query %[2]s oracles --active true
  $ %[1]s query %[2]s oracles --active false
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			activeString, err := cmd.Flags().GetString(FlagActive)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// If no active flag is passed, return all oracles
			var res proto.Message
			if activeString == "" {
				req := &types.QueryAllOraclesRequest{}
				res, err = queryClient.AllOracles(context.Background(), req)
				if err != nil {
					return err
				}
			} else {
				// Otherwise, filter using the active flag
				activeBool, err := strconv.ParseBool(activeString)
				if err != nil {
					return err
				}
				req := &types.QueryActiveOraclesRequest{
					Active: activeBool,
				}
				res, err = queryClient.ActiveOracles(context.Background(), req)
				if err != nil {
					return err
				}
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagActive, "", "Filter only active oracles")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryPendingMetricUpdates implements a command to query pending metric updates with optional
// key and/or oracle moniker filters
func GetCmdQueryPendingMetricUpdates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-metric-updates",
		Short: "Queries all pending metric update ICAs",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all metric update ICAs that have been sent to the oracle but have not received an acknowledgement
Examples:
  $ %[1]s query %[2]s pending-metric-updates 
  $ %[1]s query %[2]s pending-metric-updates --metric-key=[key]
  $ %[1]s query %[2]s pending-metric-updates --oracle-moniker=[moniker]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			metricKey, err := cmd.Flags().GetString(FlagMetricKey)
			if err != nil {
				return err
			}
			oracleMoniker, err := cmd.Flags().GetString(FlagOracleMoniker)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// If no filters are passed, return all pending metrics
			var res proto.Message
			if metricKey == "" && oracleMoniker == "" {
				req := &types.QueryAllPendingMetricUpdatesRequest{}
				res, err = queryClient.AllPendingMetricUpdates(context.Background(), req)
				if err != nil {
					return err
				}
			} else {
				// Otherwise filter by metric key and moniker
				req := &types.QueryPendingMetricUpdatesRequest{
					MetricKey:     metricKey,
					OracleMoniker: oracleMoniker,
				}
				res, err = queryClient.PendingMetricUpdates(context.Background(), req)
				if err != nil {
					return err
				}
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagMetricKey, "", "The metric key")
	cmd.Flags().String(FlagOracleMoniker, "", "The oracle moniker")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
