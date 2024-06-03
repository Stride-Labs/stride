package cli

import (
	"fmt"
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
