package keeper

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v23/x/staketia/types"
)

// Emits a successful liquid stake event, and displays metadata such as the stToken amount
func EmitSuccessfulLiquidStakeEvent(ctx sdk.Context, staker string, hostZone types.HostZone, nativeAmount, stAmount sdkmath.Int) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLiquidStakeRequest,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyLiquidStaker, staker),
			sdk.NewAttribute(types.AttributeKeyHostZone, hostZone.ChainId),
			sdk.NewAttribute(types.AttributeKeyNativeBaseDenom, hostZone.NativeTokenDenom),
			sdk.NewAttribute(types.AttributeKeyNativeIBCDenom, hostZone.NativeTokenIbcDenom),
			sdk.NewAttribute(types.AttributeKeyNativeAmount, nativeAmount.String()),
			sdk.NewAttribute(types.AttributeKeyStTokenAmount, stAmount.String()),
		),
	)
}

// Emits a successful redeem stake event, and displays metadata such as the native amount
func EmitSuccessfulRedeemStakeEvent(ctx sdk.Context, staker string, hostZone types.HostZone, nativeAmount, stAmount sdkmath.Int) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRedeemStakeRequest,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyRedeemer, staker),
			sdk.NewAttribute(types.AttributeKeyHostZone, hostZone.ChainId),
			sdk.NewAttribute(types.AttributeKeyNativeBaseDenom, hostZone.NativeTokenDenom),
			sdk.NewAttribute(types.AttributeKeyNativeIBCDenom, hostZone.NativeTokenIbcDenom),
			sdk.NewAttribute(types.AttributeKeyNativeAmount, nativeAmount.String()),
			sdk.NewAttribute(types.AttributeKeyStTokenAmount, stAmount.String()),
		),
	)
}

// Emits an event indicated the delegation record is correctly marked as done
func EmitSuccessfulConfirmDelegationEvent(ctx sdk.Context, recordId uint64, delegationAmount sdkmath.Int, txHash string, sender string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeConfirmDelegationResponse,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeRecordId, strconv.FormatUint(recordId, 10)),
			sdk.NewAttribute(types.AttributeDelegationNativeAmount, delegationAmount.String()),
			sdk.NewAttribute(types.AttributeTxHash, txHash),
			sdk.NewAttribute(types.AttributeSender, sender),
		),
	)
}

// Emits an event indicated the undelegation record is correctly marked as unbonding_in_progress
func EmitSuccessfulConfirmUndelegationEvent(ctx sdk.Context, recordId uint64, nativeAmount sdkmath.Int, txHash string, sender string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeConfirmUndelegation,
			sdk.NewAttribute(sdk.AttributeKeySender, sender),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeRecordId, strconv.FormatUint(recordId, 10)),
			sdk.NewAttribute(types.AttributeUndelegationNativeAmount, nativeAmount.String()),
			sdk.NewAttribute(types.AttributeTxHash, txHash),
			sdk.NewAttribute(types.AttributeSender, sender),
		),
	)
}

// Emits an event indicated the unbonding record is correctly marked as claimable
func EmitSuccessfulConfirmUnbondedTokenSweepEvent(ctx sdk.Context, recordId uint64, nativeAmount sdkmath.Int, txHash string, sender string) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeConfirmUnbondedTokenSweep,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeRecordId, strconv.FormatUint(recordId, 10)),
			sdk.NewAttribute(types.AttributeUndelegationNativeAmount, nativeAmount.String()),
			sdk.NewAttribute(types.AttributeTxHash, txHash),
			sdk.NewAttribute(types.AttributeSender, sender),
		),
	)
}

// Emits an event indicating a zone was halted
func EmitHaltZoneEvent(ctx sdk.Context, hostZone types.HostZone) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHostZoneHalt,
			sdk.NewAttribute(types.AttributeKeyHostZone, hostZone.ChainId),
			sdk.NewAttribute(types.AttributeKeyRedemptionRate, hostZone.RedemptionRate.String()),
		),
	)
}
