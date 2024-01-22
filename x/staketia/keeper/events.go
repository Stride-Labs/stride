package keeper

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
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
