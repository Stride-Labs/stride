package cli

import (
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

const (
	FlagDistributionStartDate = "distribution-start-date"
	FlagDistributionEndDate   = "distribution-end-date"
	FlagClawbackDate          = "clawback-date"
	FlagClaimTypeDeadlineDate = "claim-type-deadline-date"
	FlagEarlyClaimPenalty     = "early-claim-penalty"
	FlagClaimAndStakeBonus    = "claim-and-stake-bonus"
	FlagDistributionAddress   = "distribution-address"

	FlagRewardDenom    = "reward-denom"
	DefaultRewardDenom = "ustrd"

	DateLayout = "2006-01-02"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdClaimDaily(),
		CmdClaimEarly(),
		CmdCreateAirdrop(),
		CmdUpdateAirdrop(),
		CmdAddAllocations(),
		CmdUpdateUserAllocation(),
		CmdLinkAddresses(),
	)

	return cmd
}

// User transaction to claim all the pending airdrop rewards up to the current day
func CmdClaimDaily() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-daily [airdrop-id]",
		Short: "Claims all the pending airdrop rewards up to the current day",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claims all pending airdrop rewards up to the current day. 
This option is only available if the user has not already elected to claim and stake or claim early

Example:
  $ %[1]s tx %[2]s claim-daily airdrop-1 --from user
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimDaily(
				clientCtx.GetFromAddress().String(),
				airdropId,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// User transaction to claim half of their total amount now, and forfeit the other half to be clawed back
func CmdClaimEarly() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-early [airdrop-id]",
		Short: "Claims rewards immediately, but with a early claim penalty",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claims rewards immediately (including for future days), but with an early
claim penalty causing a portion of the total to be clawed back.

Example:
  $ %[1]s tx %[2]s claim-early airdrop-1 --from user
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimEarly(
				clientCtx.GetFromAddress().String(),
				airdropId,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Admin transaction to create a new airdrop
func CmdCreateAirdrop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-airdrop [airdrop-id]",
		Short: "Creates a new airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Registers a new airdrop

Example:
  $ %[1]s tx %[2]s create-airdrop airdrop-1 \
  	--distribution-start-date  2024-01-01 \
	--distribution-end-date    2024-06-01 \
	--clawback-date            2024-07-01 \
	--claim-type-deadline-date 2024-02-01 \
	--early-claim-penalty      0.5 \
	--claim-and-stake-bonus    0.05 \
	--distribution-address     strideXXX \
	--from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			denom, err := cmd.Flags().GetString(FlagRewardDenom)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse denom")
			}

			distributionStartDateString, err := cmd.Flags().GetString(FlagDistributionStartDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution start date")
			}
			distributionEndDateString, err := cmd.Flags().GetString(FlagDistributionEndDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution end date")
			}
			clawbackDateString, err := cmd.Flags().GetString(FlagClawbackDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse clawback date")
			}
			deadlineDateString, err := cmd.Flags().GetString(FlagClaimTypeDeadlineDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim deadline date date")
			}

			earlyPenaltyString, err := cmd.Flags().GetString(FlagEarlyClaimPenalty)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse early claim penalty address")
			}
			stakeBonusString, err := cmd.Flags().GetString(FlagClaimAndStakeBonus)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim and stake bonus")
			}
			distributionAddress, err := cmd.Flags().GetString(FlagDistributionAddress)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution address")
			}

			distributionStartDate, err := time.Parse(DateLayout, distributionStartDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution start date")
			}
			distributionEndDate, err := time.Parse(DateLayout, distributionEndDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution end date")
			}
			clawbackDate, err := time.Parse(DateLayout, clawbackDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse clawback date")
			}
			deadlineDate, err := time.Parse(DateLayout, deadlineDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim type deadline date")
			}

			earlyClaimPenalty, err := sdk.NewDecFromStr(earlyPenaltyString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse early penalty")
			}
			claimAndStakeBonus, err := sdk.NewDecFromStr(stakeBonusString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim and stake bonus")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateAirdrop(
				clientCtx.GetFromAddress().String(),
				airdropId,
				denom,
				&distributionStartDate,
				&distributionEndDate,
				&clawbackDate,
				&deadlineDate,
				earlyClaimPenalty,
				claimAndStakeBonus,
				distributionAddress,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagRewardDenom, DefaultRewardDenom, "Reward denom for the airdrop")
	cmd.Flags().String(FlagDistributionStartDate, "", "Start date when rewards are distributed")
	cmd.Flags().String(FlagDistributionEndDate, "", "Last date that rewards are distributed")
	cmd.Flags().String(FlagClawbackDate, "", "Date when rewards are clawed back (after distribution end date)")
	cmd.Flags().String(FlagClaimTypeDeadlineDate, "", "Deadline to decide on the claim type")
	cmd.Flags().String(FlagEarlyClaimPenalty, "", "Decimal (0 to 1) representing the penalty for claiming early")
	cmd.Flags().String(FlagClaimAndStakeBonus, "", "Decimal (0 to 1) representing the bonus for claiming and staking")
	cmd.Flags().String(FlagDistributionAddress, "", "Address of the distributor account")

	requiredFlags := []string{
		FlagDistributionStartDate,
		FlagDistributionEndDate,
		FlagClawbackDate,
		FlagClaimTypeDeadlineDate,
		FlagEarlyClaimPenalty,
		FlagClaimAndStakeBonus,
		FlagDistributionAddress,
	}
	for _, flagName := range requiredFlags {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			panic(err)
		}
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Admin transaction to update an existing airdrop
func CmdUpdateAirdrop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-airdrop [airdrop-id]",
		Short: "Updates an existing airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Updates an existing airdrop. All configurations must be provided (even those that will not be changed)

Example:
  $ %[1]s tx %[2]s update-airdrop airdrop-1 \
  	--distribution-start-date  2024-01-01 \
	--distribution-end-date    2024-06-01 \
	--clawback-date            2024-07-01 \
	--claim-type-deadline-date 2024-02-01 \
	--early-claim-penalty      0.5 \
	--claim-and-stake-bonus    0.05 \
	--distribution-address     strideXXX \
	--from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]

			denom, err := cmd.Flags().GetString(FlagRewardDenom)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse denom")
			}

			distributionStartDateString, err := cmd.Flags().GetString(FlagDistributionStartDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution start date")
			}
			distributionEndDateString, err := cmd.Flags().GetString(FlagDistributionEndDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution end date")
			}
			clawbackDateString, err := cmd.Flags().GetString(FlagClawbackDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse clawback date")
			}
			deadlineDateString, err := cmd.Flags().GetString(FlagClaimTypeDeadlineDate)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim deadline date date")
			}

			earlyPenaltyString, err := cmd.Flags().GetString(FlagEarlyClaimPenalty)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse early claim penalty address")
			}
			stakeBonusString, err := cmd.Flags().GetString(FlagClaimAndStakeBonus)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim and stake bonus")
			}
			distributionAddress, err := cmd.Flags().GetString(FlagDistributionAddress)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution address")
			}

			distributionStartDate, err := time.Parse(DateLayout, distributionStartDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution start date")
			}
			distributionEndDate, err := time.Parse(DateLayout, distributionEndDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse distribution end date")
			}
			clawbackDate, err := time.Parse(DateLayout, clawbackDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse clawback date")
			}
			deadlineDate, err := time.Parse(DateLayout, deadlineDateString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim type deadline date")
			}

			earlyClaimPenalty, err := sdk.NewDecFromStr(earlyPenaltyString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse early penalty")
			}
			claimAndStakeBonus, err := sdk.NewDecFromStr(stakeBonusString)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse claim and stake bonus")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateAirdrop(
				clientCtx.GetFromAddress().String(),
				airdropId,
				denom,
				&distributionStartDate,
				&distributionEndDate,
				&clawbackDate,
				&deadlineDate,
				earlyClaimPenalty,
				claimAndStakeBonus,
				distributionAddress,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagRewardDenom, DefaultRewardDenom, "Reward denom for the airdrop")

	cmd.Flags().String(FlagDistributionStartDate, "", "Start date when rewards are distributed")
	cmd.Flags().String(FlagDistributionEndDate, "", "Last date that rewards are distributed")
	cmd.Flags().String(FlagClawbackDate, "", "Date when rewards are clawed back (after distribution end date)")
	cmd.Flags().String(FlagClaimTypeDeadlineDate, "", "Deadline to decide on the claim type")
	cmd.Flags().String(FlagEarlyClaimPenalty, "", "Decimal (0 to 1) representing the penalty for claiming early")
	cmd.Flags().String(FlagClaimAndStakeBonus, "", "Decimal (0 to 1) representing the bonus for claiming and staking")
	cmd.Flags().String(FlagDistributionAddress, "", "Address of the distributor account")

	requiredFlags := []string{
		FlagDistributionStartDate,
		FlagDistributionEndDate,
		FlagClawbackDate,
		FlagClaimTypeDeadlineDate,
		FlagEarlyClaimPenalty,
		FlagClaimAndStakeBonus,
		FlagDistributionAddress,
	}
	for _, flagName := range requiredFlags {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			panic(err)
		}
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Admin transaction to add multiple user allocations for a given airdrop
func CmdAddAllocations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-allocations [airdrop-id] [allocations-csv-file]",
		Short: "Adds multiple user allocations for a given airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Adds multiple user allocations for a given airdrop using a CSV file

Example CSV:
 strideXXX,0,10,10,20,30,40,...
 strideYYY,0,10,10,20,30,40,...

Example Command:
  $ %[1]s tx %[2]s add-allocations airdrop-1 allocations.csv --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]
			allocationsFileName := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			allocations, err := ParseMultipleUserAllocations(allocationsFileName)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse allocations csv")
			}

			msg := types.NewMsgAddAllocations(
				clientCtx.GetFromAddress().String(),
				airdropId,
				allocations,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Admin transaction to update a user's allocation to an airdrop
func CmdUpdateUserAllocation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-user-allocation [airdrop-id] [user-address] [allocations-file]",
		Short: "Update's a single user allocation's for a given airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update's a single user allocation's for a given airdrop

Example file:
 0,10,10,20,30,40,...

Example Command:
  $ %[1]s tx %[2]s update-user-allocations airdrop-1 allocations.csv --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]
			address := args[1]
			allocationsFileName := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			allocations, err := ParseSingleUserAllocations(allocationsFileName)
			if err != nil {
				return errorsmod.Wrapf(err, "unable to parse allocations csv")
			}

			msg := types.NewMsgUpdateUserAllocation(
				clientCtx.GetFromAddress().String(),
				airdropId,
				address,
				allocations,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Admin address to link a stride and non-stride address, merging their allocations
func CmdLinkAddresses() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-addresses [airdrop-id] [stride-address] [host-address]",
		Short: "Links a stride and non-stride address, merging their allocations",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Links a stride and non-stride address, merging their allocations

Example Command:
  $ %[1]s tx %[2]s link-addresses airdrop-1 strideXXX dymXXX --from admin
`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			airdropId := args[0]
			strideAddress := args[1]
			hostAddress := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgLinkAddresses(
				clientCtx.GetFromAddress().String(),
				airdropId,
				strideAddress,
				hostAddress,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
