package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/Stride-Labs/stride/v19/x/staketia/types"
)

const (
	FlagInlcudeArchived   = "include-archived"
	FlagAddress           = "address"
	FlagUnbondingRecordId = "unbonding-record-id"
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
		CmdQueryHostZone(),
		CmdQueryDelegationRecords(),
		CmdQueryUnbondingRecords(),
		CmdQueryRedemptionRecord(),
		CmdQueryRedemptionRecords(),
		CmdQuerySlashRecords(),
	)

	return cmd
}

// CmdQueryHostZone implements a command to query the host zone struct
func CmdQueryHostZone() *cobra.Command {
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

// Queries delegation records with an option to include archive records
func CmdQueryDelegationRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation-records",
		Short: "Queries all delegation records",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all delegation records. Optionally include archived records.
Examples:
  $ %[1]s query %[2]s delegation-records
  $ %[1]s query %[2]s delegation-records --include-archived true
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			archiveString, err := cmd.Flags().GetString(FlagInlcudeArchived)
			if err != nil {
				return err
			}
			archiveBool, _ := strconv.ParseBool(archiveString)

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDelegationRecordsRequest{
				IncludeArchived: archiveBool,
			}
			res, err := queryClient.DelegationRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagInlcudeArchived, "", "Include archived records")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Queries unbonding records with an option to include archive records
func CmdQueryUnbondingRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-records",
		Short: "Queries all unbonding records",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all unbonding records. Optionally include archived records.
Example:
  $ %[1]s query %[2]s unbonding-records
  $ %[1]s query %[2]s unbonding-records --include-archived true
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			archiveString, err := cmd.Flags().GetString(FlagInlcudeArchived)
			if err != nil {
				return err
			}
			archiveBool, _ := strconv.ParseBool(archiveString)

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryUnbondingRecordsRequest{
				IncludeArchived: archiveBool,
			}
			res, err := queryClient.UnbondingRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagInlcudeArchived, "", "Include archived records")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Queries a single redemption record
func CmdQueryRedemptionRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redemption-record [epoch-number] [address]",
		Short: "Queries a single redemption record",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries a single redemption record
Example:
  $ %s query %s redemption-record 100 strideXXX
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

			req := &types.QueryRedemptionRecordRequest{
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

// Queries all redemption records with an optional address filter
func CmdQueryRedemptionRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redemption-records",
		Short: "Queries all redemption records with a optional filters",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all redemption records with an optional address or unbonding record ID filters
Examples:
  $ %[1]s query %[2]s redemption-records
  $ %[1]s query %[1]s redemption-records --address strideXXX
  $ %[1]s query %[1]s redemption-records --unbonding-record-id strideXXX
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			address, err := cmd.Flags().GetString(FlagAddress)
			if err != nil {
				return err
			}
			unbondingRecordId, err := cmd.Flags().GetUint64(FlagUnbondingRecordId)
			if err != nil {
				return err
			}

			if address != "" && unbondingRecordId != 0 {
				return errors.New("use redemption-rate query instead of redemption-rates query to filter by both unbonding record id and address")
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRedemptionRecordsRequest{
				Address:           address,
				UnbondingRecordId: unbondingRecordId,
			}
			res, err := queryClient.RedemptionRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(FlagAddress, "", "Filter by redeemer address")
	cmd.Flags().Uint64(FlagUnbondingRecordId, 0, "Filter by unbonding record ID")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Queries all slash records
func CmdQuerySlashRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slash-records",
		Short: "Queries all slash records",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Queries all slash records
Examples:
  $ %s query %s slash-records
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QuerySlashRecordsRequest{}
			res, err := queryClient.SlashRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}
