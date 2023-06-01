package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// Exchanges a user's native tokens for stTokens using the current redemption rate
// The native tokens must live on Stride with an IBC denomination before this function is called
// The typical flow consists, first, of a transfer of native tokens from the host zone to Stride,
//    and then the invocation of this LiquidStake function
//
// WARNING: This function is invoked from the begin/end blocker in a way that does not revert partial state when
//    an error is thrown (i.e. the execution is non-atomic).
//    As a result, it is important that the validation steps are positioned at the top of the function,
//    and logic that creates state changes (e.g. bank sends, mint) appear towards the end of the function
func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the host zone from the base denom in the message (e.g. uatom)
	hostZone, err := k.GetHostZoneFromHostDenom(ctx, msg.HostDenom)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "no host zone found for denom (%s)", msg.HostDenom)
	}

	// Error immediately if the host zone is halted
	if hostZone.Halted {
		return nil, errorsmod.Wrapf(types.ErrHaltedHostZone, "halted host zone found for denom (%s)", msg.HostDenom)
	}

	// Get user and module account addresses
	liquidStakerAddress, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "user's address is invalid")
	}
	hostZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, *hostZone)
	if !rateIsSafe || (err != nil) {
		return nil, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, "HostZone: %s, err: %s", hostZone.ChainId, err.Error())
	}

	// Grab the deposit record that will be used for record keeping
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	depositRecord, found := k.RecordsKeeper.GetTransferDepositRecordByEpochAndChain(ctx, strideEpochTracker.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no deposit record for epoch (%d)", strideEpochTracker.EpochNumber)
	}

	// The tokens that are sent to the protocol are denominated in the ibc hash of the native token on stride (e.g. ibc/xxx)
	nativeDenom := hostZone.IbcDenom
	nativeCoin := sdk.NewCoin(nativeDenom, msg.Amount)
	if !types.IsIBCToken(nativeDenom) {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "denom is not an IBC token (%s)", nativeDenom)
	}

	// Confirm the user has a sufficient balance to execute the liquid stake
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, nativeDenom)
	if balance.IsLT(nativeCoin) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stAmount := (sdk.NewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}

	// Transfer the native tokens from the user to module account
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneAddress, sdk.NewCoins(nativeCoin)); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Mint the stTokens and transfer them to the user
	stDenom := types.StAssetDenomFromHostZoneDenom(msg.HostDenom)
	stCoin := sdk.NewCoin(stDenom, stAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to mint coins")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account", stCoin.String())
	}

	// Update the liquid staked amount on the deposit record
	depositRecord.Amount = depositRecord.Amount.Add(msg.Amount)
	k.RecordsKeeper.SetDepositRecord(ctx, *depositRecord)

	// Emit liquid stake event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLiquidStakeRequest,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyLiquidStaker, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyHostZone, hostZone.ChainId),
			sdk.NewAttribute(types.AttributeKeyNativeBaseDenom, msg.HostDenom),
			sdk.NewAttribute(types.AttributeKeyNativeIBCDenom, hostZone.IbcDenom),
			sdk.NewAttribute(types.AttributeKeyNativeAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyStTokenAmount, stAmount.String()),
		),
	)

	k.hooks.AfterLiquidStake(ctx, liquidStakerAddress)
	return &types.MsgLiquidStakeResponse{}, nil
}
