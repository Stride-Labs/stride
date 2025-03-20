package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

const (
	FlagMinRedemptionRate            = "min-redemption-rate"
	FlagMaxRedemptionRate            = "max-redemption-rate"
	FlagCommunityPoolTreasuryAddress = "community-pool-treasury-address"
	FlagMaxMessagesPerIcaTx          = "max-messages-per-ica-tx"
	FlagLegacy                       = "legacy"
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

	cmd.AddCommand(CmdLiquidStake())
	cmd.AddCommand(CmdLSMLiquidStake())
	cmd.AddCommand(CmdRegisterHostZone())
	cmd.AddCommand(CmdRedeemStake())
	cmd.AddCommand(CmdClaimUndelegatedTokens())
	cmd.AddCommand(CmdRebalanceValidators())
	cmd.AddCommand(CmdAddValidators())
	cmd.AddCommand(CmdChangeValidatorWeight())
	cmd.AddCommand(CmdChangeMultipleValidatorWeight())
	cmd.AddCommand(CmdDeleteValidator())
	cmd.AddCommand(CmdRestoreInterchainAccount())
	cmd.AddCommand(CmdCloseDelegationChannel())
	cmd.AddCommand(CmdUpdateValidatorSharesExchRate())
	cmd.AddCommand(CmdCalibrateDelegation())
	cmd.AddCommand(CmdClearBalance())
	cmd.AddCommand(CmdUpdateInnerRedemptionRateBounds())
	cmd.AddCommand(CmdResumeHostZone())
	cmd.AddCommand(CmdSetCommunityPoolRebate())
	cmd.AddCommand(CmdToggleTradeController())

	return cmd
}

func CmdLiquidStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquid-stake [amount] [hostDenom]",
		Short: "Broadcast message liquid-stake",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAmount, found := sdk.NewIntFromString(args[0])
			if !found {
				return errorsmod.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
			}
			argHostDenom := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgLiquidStake(
				clientCtx.GetFromAddress().String(),
				argAmount,
				argHostDenom,
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

func CmdLSMLiquidStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsm-liquid-stake [amount] [lsm-token-denom]",
		Short: "Broadcast message lsm-liquid-stake",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount, found := sdk.NewIntFromString(args[0])
			if !found {
				return errorsmod.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
			}
			denom := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgLSMLiquidStake(
				clientCtx.GetFromAddress().String(),
				amount,
				denom,
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

func CmdRegisterHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-host-zone [connection-id] [host-denom] [bech32prefix] [ibc-denom] [channel-id] [unbonding-period] [lsm-enabled]",
		Short: "Broadcast message register-host-zone",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			connectionId := args[0]
			hostDenom := args[1]
			bech32prefix := args[2]
			ibcDenom := args[3]
			channelId := args[4]
			unbondingPeriod, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return err
			}
			lsmEnabled, err := strconv.ParseBool(args[6])
			if err != nil {
				return err
			}

			minRedemptionRateStr, err := cmd.Flags().GetString(FlagMinRedemptionRate)
			if err != nil {
				return err
			}
			minRedemptionRate := sdk.ZeroDec()
			if minRedemptionRateStr != "" {
				minRedemptionRate, err = sdk.NewDecFromStr(minRedemptionRateStr)
				if err != nil {
					return err
				}
			}

			maxRedemptionRateStr, err := cmd.Flags().GetString(FlagMaxRedemptionRate)
			if err != nil {
				return err
			}
			maxRedemptionRate := sdk.ZeroDec()
			if maxRedemptionRateStr != "" {
				maxRedemptionRate, err = sdk.NewDecFromStr(maxRedemptionRateStr)
				if err != nil {
					return err
				}
			}

			communityPoolTreasuryAddress, err := cmd.Flags().GetString(FlagCommunityPoolTreasuryAddress)
			if err != nil {
				return err
			}

			maxMessagesPerIcaTx, err := cmd.Flags().GetUint64(FlagMaxMessagesPerIcaTx)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterHostZone(
				clientCtx.GetFromAddress().String(),
				connectionId,
				bech32prefix,
				hostDenom,
				ibcDenom,
				channelId,
				unbondingPeriod,
				minRedemptionRate,
				maxRedemptionRate,
				lsmEnabled,
				communityPoolTreasuryAddress,
				maxMessagesPerIcaTx,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagMinRedemptionRate, "", "minimum redemption rate")
	cmd.Flags().String(FlagMaxRedemptionRate, "", "maximum redemption rate")
	cmd.Flags().String(FlagCommunityPoolTreasuryAddress, "", "community pool treasury address")
	cmd.Flags().Uint64(FlagMaxMessagesPerIcaTx, 0, "maximum number of ICA txs in a given tx")

	return cmd
}

func CmdRedeemStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem-stake [amount] [hostZoneID] [receiver]",
		Short: "Broadcast message redeem-stake",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			argAmount, found := sdk.NewIntFromString(args[0])
			if !found {
				return errorsmod.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
			}
			hostZoneID := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			argReceiver := args[2]

			msg := types.NewMsgRedeemStake(
				clientCtx.GetFromAddress().String(),
				argAmount,
				hostZoneID,
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

func CmdRebalanceValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rebalance-validators [host-zone] [num-to-rebalance]",
		Short: "Broadcast message rebalanceValidators",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argHostZone := args[0]
			argNumValidators, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRebalanceValidators(
				clientCtx.GetFromAddress().String(),
				argHostZone,
				argNumValidators,
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
		Args:  cobra.ExactArgs(2),
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
		Short: "Broadcast message update-delegation",
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
		Short: "Broadcast message calibrate-delegation",
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

func CmdClearBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-balance [chain-id] [amount] [channel-id]",
		Short: "Broadcast message clear-balance",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]
			argAmount, found := sdk.NewIntFromString(args[1])
			if !found {
				return errorsmod.Wrap(sdkerrors.ErrInvalidType, "can not convert string to int")
			}
			argChannelId := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClearBalance(
				clientCtx.GetFromAddress().String(),
				argChainId,
				argAmount,
				argChannelId,
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
			minInnerRedemptionRate := sdk.MustNewDecFromStr(args[1])
			maxInnerRedemptionRate := sdk.MustNewDecFromStr(args[2])

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

func CmdResumeHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume-host-zone [chainid]",
		Short: "Broadcast message resume-host-zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChainId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgResumeHostZone(
				clientCtx.GetFromAddress().String(),
				argChainId,
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

func CmdSetCommunityPoolRebate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-rebate [chain-id] [rebate-rate] [liquid-staked-sttoken-amount]",
		Short: "Registers or updates a community pool rebate",
		Long: strings.TrimSpace(`Registers a community pool rebate by specifying the rebate percentage (as a decimal)
and the amount liquid staked, denominated in the number of stTokens received. 
E.g. to specify a 20% rebate, the rebate rate should be 0.2

If a 0.0 rebate or 0 token liquid stake is specified, the rebate will be deleted.
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			chainId := args[0]
			rebatePercentage, err := sdk.NewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("unable to parse rebate percentage: %s", err.Error())
			}
			liquidStakedStTokenAmount, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return errors.New("unable to parse liquid stake amount")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSetCommunityPoolRebate(
				clientCtx.GetFromAddress().String(),
				chainId,
				rebatePercentage,
				liquidStakedStTokenAmount,
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

func CmdToggleTradeController() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toggle-trade-controller [trade-chain-id] [grant|revoke] [address]",
		Short: "Submits an ICA tx to grant or revoke permissions to trade on behalf of the trade ICA",
		Long: strings.TrimSpace(`Submits an ICA tx to grant or revoke permissions to trade on behalf of the trade ICA
Ex:
>>> strided tx toggle-trade-controller osmosis-1 grant osmoXXX
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			chainId := args[0]
			permissionChangeString := args[1]
			address := args[2]

			permissionChangeInt, ok := types.AuthzPermissionChange_value[strings.ToUpper(permissionChangeString)]
			if !ok {
				return errors.New("invalid permission change, must be either 'grant' or 'revoke'")
			}
			permissionChange := types.AuthzPermissionChange(permissionChangeInt)

			legacy, err := cmd.Flags().GetBool(FlagLegacy)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgToggleTradeController(
				clientCtx.GetFromAddress().String(),
				chainId,
				permissionChange,
				address,
				legacy,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(FlagLegacy, false, "Use legacy osmosis swap message from gamm")

	return cmd
}
