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

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

const (
	FlagArchive = "archive"
	FlagAddress = "address"
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
		GetCmdQueryHostZone(),
		GetCmdQueryDelegationRecords(),
		GetCmdQueryUnbondingRecords(),
		GetCmdQueryRedemptionRecord(),
		GetCmdQueryAllRedemptionRecords(),
		GetCmdQuerySlashRecords(),
	)

	return cmd
}

// GetCmdQueryHostZone implements a command to query the host zone struct
func GetCmdQueryHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host-zone",
		Short: "Queries the host zone struct",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries the host zone
Example:
  $ %s query %s host-zone
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryHostZoneRequest{}
			res, err := queryClient.HostZone(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// GetCmdQueryDelegationRecords implements a command to query the delegation records with
// an optional to include archived records
func GetCmdQueryDelegationRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation-records",
		Short: "Queries the delegation records",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries the delegation records. Optionally including the archived records.
Example:
  $ %s query %s delegation-records
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			archiveString, err := cmd.Flags().GetString(FlagArchive)
			if err != nil {
				return err
			}
			archiveBool, err := strconv.ParseBool(archiveString)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDelegationRecordsRequest{
				Archive: archiveBool,
			}
			res, err := queryClient.DelegationRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagArchive, "", "Include archived records")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryUnbondingRecords implements a command to query the unbonding records with
// an optional to include archived records
func GetCmdQueryUnbondingRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-records",
		Short: "Queries the unbonding records",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries the unbonding records. Optionally including the archived records.
Example:
  $ %s query %s unbonidng-records
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			archiveString, err := cmd.Flags().GetString(FlagArchive)
			if err != nil {
				return err
			}
			archiveBool, err := strconv.ParseBool(archiveString)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryUnbondingRecordsRequest{
				Archive: archiveBool,
			}
			res, err := queryClient.UnbondingRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagArchive, "", "Include archived records")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRedemptionRecord implements a command to query a single redemption record
func GetCmdQueryRedemptionRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redemption-record [epoch-number] [address]",
		Short: "Queries a single redemption record",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries a single redemption record
Example:
  $ %s query %s redemption-record [epoch-number] [address]
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			unbondingRecordId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			address := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRedemptionRecord{
				UnbondingRecordId: unbondingRecordId,
				Address:           address,
			}
			res, err := queryClient.RedemptionRecord(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// GetCmdQueryAllRedemptionRecords implements a command to query all redemption records
// with an optional address filter
func GetCmdQueryAllRedemptionRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redemption-records",
		Short: "Queries all redemption records with an optional address filter",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all redemption records with an optional address filter
Examples:
  $ %[1]s query %[2]s redemption-records
  $ %[1]s query %[1]s redemption-records --address strideXXX
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			address, err := cmd.Flags().GetString(FlagAddress)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllRedemptionRecords{
				Address: address,
			}
			res, err := queryClient.AllRedemptionRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}

// GetCmdQuerySlashRecords implements a command to query all slash records
func GetCmdQuerySlashRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slash-records",
		Short: "Queries all slash records",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all slash records
Examples:
  $ %s query %s slash-records
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QuerySlashRecords{}
			res, err := queryClient.SlashRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}
