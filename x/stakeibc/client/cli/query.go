package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group stakeibc queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdShowValidators())
	cmd.AddCommand(CmdListHostZone())
	cmd.AddCommand(CmdShowHostZone())
	cmd.AddCommand(CmdModuleAddress())
	cmd.AddCommand(CmdShowInterchainAccount())
	cmd.AddCommand(CmdListEpochTracker())
	cmd.AddCommand(CmdShowEpochTracker())
	cmd.AddCommand(CmdNextPacketSequence())
	cmd.AddCommand(CmdListTradeRoutes())

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-validators [chain-id]",
		Short: "shows validators",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId := args[0]

			params := &types.QueryGetValidatorsRequest{ChainId: chainId}

			res, err := queryClient.Validators(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdListHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-host-zone",
		Short: "list all HostZone",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllHostZoneRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.HostZoneAll(context.Background(), params)
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

func CmdShowHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-host-zone [chain-id]",
		Short: "shows a HostZone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId := args[0]

			params := &types.QueryGetHostZoneRequest{
				ChainId: chainId,
			}

			res, err := queryClient.HostZone(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
