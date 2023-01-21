package keeper

import (
	"context"
	"fmt"
	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k msgServer) FastUnbond(goCtx context.Context, msg *types.MsgFastUnbond) (*types.MsgFastUnbondResponse, error) {
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

	// construct desired unstaking amount from host zone
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	nativeAmount := sdk.NewDecFromInt(msg.Amount).Mul(hostZone.RedemptionRate).RoundInt()

	if nativeAmount.GT(hostZone.StakedBal) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAmount, "cannot unstake an amount g.t. staked balance on host zone: %v", msg.Amount)
	}

	// safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if !rateIsSafe || (err != nil) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed. hostZone: %s, err: %s", hostZone.String(), err.Error())
		return nil, sdkerrors.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}

	// TODO(TEST-112) bigint safety
	coinString := nativeAmount.String() + stDenom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse inCoin: %s. err: %s", coinString, err.Error())
	}
	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !nativeAmount.IsPositive() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %v", msg.Amount)
	}
	// 	- Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, stDenom)
	k.Logger(ctx).Info(fmt.Sprintf("Redemption issuer IBCDenom balance: %v%s", balance.Amount, balance.Denom))
	k.Logger(ctx).Info(fmt.Sprintf("Redemption requested redemotion amount: %v%s", inCoin.Amount, inCoin.Denom))
	if balance.Amount.LT(msg.Amount) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %v, balance %v: ", msg.Amount, balance.Amount)
	}

	// Find and subtract this amount from a deposit record if it is big enough
	epochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	// Does this use too much gas?
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
	err = k.bankKeeper.SendCoins(ctx, bech32ZoneAddress, sender, sdk.NewCoins(inCoin))
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(inCoins) from account to module")
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't send %v derivative %s tokens from module account. err: %s", inCoin.Amount, inCoin.Denom, err.Error())
	}

	// Escrow user's balance
	redeemCoin := sdk.NewCoins(sdk.NewCoin(stDenom, msg.Amount))
	err = k.bankKeeper.SendCoins(ctx, sender, bech32ZoneAddress, redeemCoin)
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(inCoins) from account to module")
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't send %v derivative %s tokens to module account. err: %s", msg.Amount, hostZone.HostDenom, err.Error())
	}

	k.Logger(ctx).Info(fmt.Sprintf("executed fast unbond: %s", msg.String()))
	return &types.MsgFastUnbondResponse{}, nil
}
