package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

const (
	FlagMinRedemptionRate = "min-redemption-rate"
	FlagMaxRedemptionRate = "max-redemption-rate"
)

var _ = strconv.Itoa(0)

// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
func CmdRegisterHostZone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-host-zone [connection-id] [host-denom] [bech32prefix] [ibc-denom] [channel-id] [unbonding-frequency]",
		Short: "Broadcast message register-host-zone",
		Args:  cobra.ExactArgs(6),
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
			unbondingFrequency, err := strconv.ParseUint(args[5], 10, 64)
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

			msg := types.NewMsgRegisterHostZone(
				clientCtx.GetFromAddress().String(),
				connectionId,
				bech32prefix,
				hostDenom,
				ibcDenom,
				channelId,
				unbondingFrequency,
				minRedemptionRate,
				maxRedemptionRate,
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

	return cmd
}
