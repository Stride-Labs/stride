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

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

const (
	FlagMetricKey     = "metric-key"
	FlagOracleChainId = "oracle-chain-id"
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
		GetCmdQueryMetrics(),
	)

	return cmd
}

// GetCmdQueryOracle implements a command to query a specific oracle using the oracle's chain ID
func GetCmdQueryOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle [chain-id]",
		Short: "Queries a specific oracle",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries a specific oracle using the oracle's chain ID
Example:
  $ %s query %s oracle [chain-id]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainId := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOracleRequest{
				ChainId: chainId,
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

// GetCmdQueryMetrics implements a command to query metrics with optional
// key and/or oracle chain-id filters
func GetCmdQueryMetrics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Queries all metric update ICAs",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all metrics with optional filters
Examples:
  $ %[1]s query %[2]s metrics 
  $ %[1]s query %[2]s metrics --metric-key=[key]
  $ %[1]s query %[2]s metrics --oracle-chain-id=[chain-id]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			metricKey, err := cmd.Flags().GetString(FlagMetricKey)
			if err != nil {
				return err
			}
			oracleChainId, err := cmd.Flags().GetString(FlagOracleChainId)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// If no filters are passed, return all pending metrics
			req := &types.QueryMetricsRequest{
				MetricKey:     metricKey,
				OracleChainId: oracleChainId,
			}
			res, err := queryClient.Metrics(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagMetricKey, "", "The metric key")
	cmd.Flags().String(FlagOracleChainId, "", "The oracle chain ID")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
