package keeper

import (
	"context"
	"fmt"
	epochtypes "github.com/Stride-Labs/stride/v5/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func (k msgServer) InstantRedeemStake(goCtx context.Context, msg *types.MsgInstantRedeemStake) (*types.MsgInstantRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("fast unbond: %s", msg.String()))

	// get our addresses, make sure they're valid
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "creator address is invalid: %s. err: %s", msg.Creator, err.Error())
	}

	// then make sure host zone is valid
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "host zone is invalid: %s", msg.HostZone)
	}

	// safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if !rateIsSafe || (err != nil) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed. hostZone: %s, err: %s", hostZone.String(), err.Error())
		return nil, sdkerrors.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}

	// construct desired unstaking amount from host zone
	nativeAmount := sdk.NewDecFromInt(msg.Amount).Mul(hostZone.RedemptionRate).RoundInt()
	// Redemption amount must not be greater than staked balance in that zone.
	if nativeAmount.GT(hostZone.StakedBal) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAmount, "cannot unstake an amount g.t. staked balance on host zone: %v", msg.Amount)
	}
	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !nativeAmount.IsPositive() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %v", msg.Amount)
	}

	// get the native Ibc coin to return to user
	ibcDenom := hostZone.GetIbcDenom()
	nativeCoin := sdk.NewCoin(ibcDenom, nativeAmount)

	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	// 	- Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, stDenom)
	k.Logger(ctx).Info(fmt.Sprintf("Redemption issuer IBCDenom balance: %v%s", balance.Amount, balance.Denom))
	if balance.Amount.LT(msg.Amount) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %v, balance %v: ", msg.Amount, balance.Amount)
	}

	// Find and subtract this amount from a deposit record if it is big enough
	epochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	depositRecord, found := k.RecordsKeeper.GetDepositRecordByEpochAndChain(ctx, epochTracker.EpochNumber, hostZone.ChainId)
	if !found {
		k.Logger(ctx).Error("failed to find deposit record")
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, fmt.Sprintf("no deposit record for epoch (%d)", epochTracker.EpochNumber))
	}
	if nativeAmount.GT(depositRecord.Amount) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAmount, "cannot fast unbond an amount %v g.t. pending deposit balance on host zone: %v", nativeAmount, msg.Amount)
	}
	depositRecord.Amount = depositRecord.Amount.Sub(nativeAmount)
	bech32ZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.Address, hostZone.ChainId)
	}
	// Send that amount back to user
	err = k.bankKeeper.SendCoins(ctx, bech32ZoneAddress, sender, sdk.NewCoins(nativeCoin))
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(nativeCoin) back to user")
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't send %v derivative %s tokens from module account. err: %s", nativeCoin.Amount, nativeCoin.Denom, err.Error())
	}

	// Send tokens back to module to burn
	stCoin := sdk.NewCoin(stDenom, msg.Amount)
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		return nil, fmt.Errorf("could not send coins from account %s to module %s. err: %s", sender, types.ModuleName, err.Error())
	}

	// Finally burn the stTokens
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to burn stAssets upon successful unbonding %s", err.Error()))
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't burn %v%s tokens in module account. err: %s", msg.Amount, stDenom, err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("Total supply %s", k.bankKeeper.GetSupply(ctx, stDenom)))

	k.Logger(ctx).Info(fmt.Sprintf("executed fast unbond: %s", msg.String()))
	return &types.MsgInstantRedeemStakeResponse{}, nil
}
