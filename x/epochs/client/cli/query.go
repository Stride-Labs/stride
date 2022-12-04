package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/epochs/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group epochs queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdEpochsInfos(),
		GetCmdCurrentEpoch(),
		GetCmdSecondsRemaining(),
	)

	return cmd
}

// GetCmdEpochsInfos provide running epochInfos
func GetCmdEpochsInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-infos",
		Short: "Query running epochInfos",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query epochs epoch-infos`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryEpochsInfoRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.EpochInfos(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdCurrentEpoch provides current epoch by specified identifier
func GetCmdCurrentEpoch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-epoch",
		Short: "Query current epoch by specified identifier",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query epochs current-epoch week`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.CurrentEpoch(cmd.Context(), &types.QueryCurrentEpochRequest{
				Identifier: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdSecondsRemaining provides seconds-remaining by specified identifier
func GetCmdSecondsRemaining() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seconds-remaining",
		Short: "Query seconds remaining by specified identifier",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query epochs seconds-remaining week`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.EpochInfo(cmd.Context(), &types.QueryEpochInfoRequest{
				Identifier: args[0],
			})
			if err != nil {
				return err
			}

			// duration: seconds
			duration, err := cast.ToInt64E(res.Epoch.Duration.Seconds())
			if err != nil {
				return err
			}
			// current epoch start time
			startTime := res.Epoch.CurrentEpochStartTime.Unix()
			// diff in seconds
			currTime := time.Now().Unix()
			// 60 - (100 - 80)
			remaining := duration - (currTime - startTime)

			return clientCtx.PrintString(strconv.FormatInt(remaining, 10))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
