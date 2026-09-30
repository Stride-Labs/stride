package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

var DefaultRelativePacketTimeoutTimestamp = cast.ToUint64((time.Duration(10) * time.Minute).Nanoseconds())

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdClaimUndelegatedTokens())
	cmd.AddCommand(CmdAddValidators())
	cmd.AddCommand(CmdChangeValidatorWeight())
	cmd.AddCommand(CmdChangeMultipleValidatorWeight())
	cmd.AddCommand(CmdDeleteValidator())
	cmd.AddCommand(CmdRestoreInterchainAccount())
	cmd.AddCommand(CmdCloseDelegationChannel())
	cmd.AddCommand(CmdUpdateValidatorSharesExchRate())
	cmd.AddCommand(CmdCalibrateDelegation())
	cmd.AddCommand(CmdUpdateInnerRedemptionRateBounds())

	return cmd
}

func CmdClaimUndelegatedTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-undelegated-tokens [host-zone] [epoch] [receiver]",
		Short: "Broadcast message claimUndelegatedTokens",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostZone := args[0]
			argEpoch, err := cast.ToUint64E(args[1])
			if err != nil {
				return err
			}
			argReceiver := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimUndelegatedTokens(
				clientCtx.GetFromAddress().String(),
				argHostZone,
				argEpoch,
				argReceiver,
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

func CmdAddValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-validators [host-zone] [validator-list-file]",
		Short: "Broadcast message add-validators",
		Long: strings.TrimSpace(
			`Add validators and weights using a JSON file in the following format
	{
		"validator_weights": [
			{"address": "cosmosXXX", "weight": 1},
			{"address": "cosmosXXX", "weight": 2}
		]
	}	
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostZone := args[0]
			validatorListProposalFile := args[1]

			validators, err := parseAddValidatorsFile(validatorListProposalFile)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddValidators(
				clientCtx.GetFromAddress().String(),
				hostZone,
				validators.Validators,
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

// Updates the weight for a single validator
func CmdChangeValidatorWeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-validator-weight [host-zone] [address] [weight]",
		Short: "Broadcast message change-validator-weight to update the weight for a single validator",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostZone := args[0]
			valAddress := args[1]
			weight, err := cast.ToUint64E(args[2])
			if err != nil {
				return err
			}
			weights := []*types.ValidatorWeight{
				{
					Address: valAddress,
					Weight:  weight,
				},
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeValidatorWeights(
				clientCtx.GetFromAddress().String(),
				hostZone,
				weights,
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

// Updates the weight for multiple validators
//
// Accepts a file in the following format:
//
//	{
//		"validator_weights": [
//		     {"address": "cosmosXXX", "weight": 1},
//			 {"address": "cosmosXXX", "weight": 2}
//	    ]
//	}
func CmdChangeMultipleValidatorWeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-validator-weights [host-zone] [validator-weight-file]",
		Short: "Broadcast message change-validator-weights to update the weights for multiple validators",
		Long: strings.TrimSpace(
			`Changes multiple validator weights at once, using a JSON file in the following format
	{
		"validator_weights": [
			{"address": "cosmosXXX", "weight": 1},
			{"address": "cosmosXXX", "weight": 2}
		]
	}	
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			hostZone := args[0]
			validatorWeightChangeFile := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			weights, err := parseChangeValidatorWeightsFile(validatorWeightChangeFile)
			if err != nil {
				return err
			}

			msg := types.NewMsgChangeValidatorWeights(
				clientCtx.GetFromAddress().String(),
				hostZone,
				weights,
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

func CmdDeleteValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-validator [host-zone] [address]",
		Short: "Broadcast message delete-validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostZone := args[0]
			argAddress := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteValidator(
				clientCtx.GetFromAddress().String(),
				argHostZone,
				argAddress,
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

func CmdRestoreInterchainAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-interchain-account [chain-id] [connection-id] [account-owner]",
		Short: "Broadcast message restore-interchain-account",
		Long: strings.TrimSpace(
			`Restores a closed channel associated with an interchain account.
Specify the chain ID and account owner - where the owner is the alias for the ICA account

For host zone ICA accounts, the owner is of the form {chainId}.{accountType}
ex:
>>> strided tx restore-interchain-account cosmoshub-4 connection-0 cosmoshub-4.DELEGATION 

For trade route ICA accounts, the owner is of the form:
    {chainId}.{rewardDenom}-{hostDenom}.{accountType}
ex:
>>> strided tx restore-interchain-account dydx-mainnet-1 connection-1 dydx-mainnet-1.uusdc-udydx.CONVERTER_TRADE 
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			chainId := args[0]
			connectionId := args[1]
			accountOwner := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRestoreInterchainAccount(
				clientCtx.GetFromAddress().String(),
				chainId,
				connectionId,
				accountOwner,
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

func CmdCloseDelegationChannel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-delegation-channel [chain-id]",
		Short: "Broadcast message close-delegation-channel",
		Long: strings.TrimSpace(
			`Closes a delegation ICA channel. This can only be run by the admin

Ex:
>>> strided tx close-delegation-channel cosmoshub-4
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			chainId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCloseDelegationChannel(
				clientCtx.GetFromAddress().String(),
				chainId,
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

func CmdUpdateValidatorSharesExchRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-delegation [chainid] [valoper]",
		Short: "Broadcast message update-delegation (admin only: refreshes a validator's shares-to-tokens rate and applies any slash)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostdenom := args[0]
			argValoper := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateValidatorSharesExchRate(
				clientCtx.GetFromAddress().String(),
				argHostdenom,
				argValoper,
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

func CmdCalibrateDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calibrate-delegation [chainid] [valoper]",
		Short: "Broadcast message calibrate-delegation (admin only: trues up a validator's recorded delegation to the host)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]
			argValoper := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCalibrateDelegation(
				clientCtx.GetFromAddress().String(),
				argChainId,
				argValoper,
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

func CmdUpdateInnerRedemptionRateBounds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-redemption-rate-bounds [chainid] [min-bound] [max-bound]",
		Short: "Broadcast message set-redemption-rate-bounds",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]
			minInnerRedemptionRate := sdkmath.LegacyMustNewDecFromStr(args[1])
			maxInnerRedemptionRate := sdkmath.LegacyMustNewDecFromStr(args[2])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateInnerRedemptionRateBounds(
				clientCtx.GetFromAddress().String(),
				argChainId,
				minInnerRedemptionRate,
				maxInnerRedemptionRate,
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
