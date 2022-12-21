package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
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

	cmd.AddCommand(CmdAddRateLimit())
	cmd.AddCommand(CmdUpdateRateLimit())
	cmd.AddCommand(CmdRemoveRateLimit())
	cmd.AddCommand(CmdResetRateLimit())
	return cmd
}

// Adds a new rate limit
func CmdAddRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-rate-limit [denom] [channel-id] [max-percent-send] [max-percent-recv] [duration-hours]",
		Short: "Broadcast message add-rate-limit",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			denom := args[0]
			channelId := args[1]

			maxPercentSend, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidType, "can not convert max-percent-send string to sdk.Int")
			}

			maxPercentRecv, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidType, "can not convert max-percent-recv string to sdk.Int")
			}

			durationHours, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgAddRateLimit(
				clientCtx.GetFromAddress().String(),
				denom,
				channelId,
				maxPercentSend,
				maxPercentRecv,
				durationHours,
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

// Update a rate limit
func CmdUpdateRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-rate-limit [denom] [channel-id] [max-percent-send] [max-percet-recv] [duration-hours]",
		Short: "Broadcast message update-rate-limit",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			denom := args[0]
			channelId := args[1]

			maxPercentSend, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidType, "can not convert max-percent-send string to sdk.Int")
			}

			maxPercentRecv, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidType, "can not convert max-percent-recv string to sdk.Int")
			}

			durationHours, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgUpdateRateLimit(
				clientCtx.GetFromAddress().String(),
				denom,
				channelId,
				maxPercentSend,
				maxPercentRecv,
				durationHours,
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

// Remove a rate limit
func CmdRemoveRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-rate-limit [denom] [channel-id]",
		Short: "Broadcast message remove-rate-limit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			denom := args[0]
			channelId := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgRemoveRateLimit(
				clientCtx.GetFromAddress().String(),
				denom,
				channelId,
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

// Reset a rate limit
func CmdResetRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-rate-limit [denom] [channel-id]",
		Short: "Broadcast message reset-rate-limit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			denom := args[0]
			channelId := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgResetRateLimit(
				clientCtx.GetFromAddress().String(),
				denom,
				channelId,
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
