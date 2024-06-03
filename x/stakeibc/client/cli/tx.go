package cli

import (
	"fmt"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

const (
	FlagMinRedemptionRate            = "min-redemption-rate"
	FlagMaxRedemptionRate            = "max-redemption-rate"
	FlagCommunityPoolTreasuryAddress = "community-pool-treasury-address"
	FlagMaxMessagesPerIcaTx          = "max-messages-per-ica-tx"
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
