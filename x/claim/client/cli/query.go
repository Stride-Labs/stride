package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	claimQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	claimQueryCmd.AddCommand(
		GetCmdQueryDistributorAccountBalance(),
		GetCmdQueryParams(),
		GetCmdQueryClaimRecord(),
		GetCmdQueryClaimableForAction(),
		GetCmdQueryTotalClaimable(),
		GetCmdQueryUserVestings(),
	)

	return claimQueryCmd
}

// GetCmdQueryParams implements a command to return the current minting
// parameters.
func GetCmdQueryDistributorAccountBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distributor-account-balance [airdrop-identifier]",
		Short: "Query the current distributor's account balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			argAirdropIdentifier := args[0]
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDistributorAccountBalanceRequest{
				AirdropIdentifier: argAirdropIdentifier,
			}
			res, err := queryClient.DistributorAccountBalance(context.Background(), req)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryParams implements a command to return the current minting
// parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current claims parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryClaimRecord implements the query claim-records command.
func GetCmdQueryClaimRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-record [airdrop-identifier] [address]",
		Args:  cobra.ExactArgs(2),
		Short: "Query the claim record for an account.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the claim record for an account.
This contains an address' initial claimable amounts, and the completed actions.

Example:
$ %s query claim claim-record <address>
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// Query store
			res, err := queryClient.ClaimRecord(context.Background(), &types.QueryClaimRecordRequest{AirdropIdentifier: args[0], Address: args[1]})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryClaimableForAction implements the query claimable for action command.
func GetCmdQueryClaimableForAction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claimable-for-action [airdrop-identifier] [address] [action]",
		Args:  cobra.ExactArgs(3),
		Short: "Query an address' claimable amount for a specific action",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query an address' claimable amount for a specific action

Example:
$ %s query claim claimable-for-action stride1h4astdfzjhcwahtfrh24qtvndzzh49xvqtfftk ActionLiquidStake
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			action, ok := types.Action_value[args[2]]
			if !ok {
				return fmt.Errorf("invalid Action type: %s.  Valid actions are %s, %s, %s", args[2],
					types.ACTION_FREE, types.ACTION_LIQUID_STAKE, types.ACTION_DELEGATE_STAKE)
			}

			// Query store
			res, err := queryClient.ClaimableForAction(context.Background(), &types.QueryClaimableForActionRequest{
				AirdropIdentifier: args[0],
				Address:           args[1],
				Action:            types.Action(action),
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryClaimable implements the query claimables command.
func GetCmdQueryTotalClaimable() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-claimable [airdrop-identifier] [address] [include-claimed]",
		Args:  cobra.ExactArgs(3),
		Short: "Query the total claimable amount remaining for an account.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the total claimable amount remaining for an account.
Example:
$ %s query claim total-claimable stride stride1h4astdfzjhcwahtfrh24qtvndzzh49xvqtfftk true
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			// Query store
			res, err := queryClient.TotalClaimable(context.Background(), &types.QueryTotalClaimableRequest{
				AirdropIdentifier: args[0],
				Address:           args[1],
				IncludeClaimed:    args[2] == "true",
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryUserVestings implements the query user vestings command.
func GetCmdQueryUserVestings() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-vestings [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query user vestings.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query user vestings for an account.
Example:
$ %s query claim user-vestings stride1h4astdfzjhcwahtfrh24qtvndzzh49xvqtfftk
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			// Query store
			res, err := queryClient.UserVestings(context.Background(), &types.QueryUserVestingsRequest{
				Address: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
