package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v35/utils"
	epochtypes "github.com/Stride-Labs/stride/v35/x/epochs/types"
	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// LiquidStake exchanges native tokens for stTokens at the current redemption rate.
//
// The user-facing MsgLiquidStake handler was removed in the v35 wind-down; this keeper
// method is what autopilot (inbound IBC liquid stakes), the community pool staking flow and
// the reward collector still call. The native tokens must live on Stride with an IBC
// denomination before this function is called.
//
// WARNING: This function is invoked from the begin/end blocker in a way that does not revert
// partial state when an error is thrown (i.e. the execution is non-atomic). As a result, the
// validation steps are positioned at the top of the function, and logic that creates state
// changes (bank sends, mint) appears towards the end.
func (k Keeper) LiquidStake(ctx sdk.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
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
	hostZoneDepositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
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
	stAmount := (sdkmath.LegacyNewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}

	// Transfer the native tokens from the user to module account
	// Note: checkBlockedAddr=false because hostZoneDepositAddress is a module
	if err := utils.SafeSendCoins(false, k.bankKeeper, ctx, liquidStakerAddress, hostZoneDepositAddress, sdk.NewCoins(nativeCoin)); err != nil {
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
	EmitSuccessfulLiquidStakeEvent(ctx, msg, *hostZone, stAmount)

	k.hooks.AfterLiquidStake(ctx, liquidStakerAddress)
	return &types.MsgLiquidStakeResponse{StToken: stCoin}, nil
}
