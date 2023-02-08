package keeper

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"fmt"
	recordstypes "github.com/Stride-Labs/stride/v5/x/records/types"
	"github.com/spf13/cast"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func (k msgServer) InstantRedeemStake(goCtx context.Context, msg *types.MsgInstantRedeemStake) (*types.MsgInstantRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("instant redeem stake: %s", msg.String()))

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

	// Determine the instant redemption commission rate to the relevant portion can be sent to the community pool for now
	params := k.GetParams(ctx)
	instantRedemptionCommissionInt, err := cast.ToInt64E(params.InstantRedemptionCommission)
	if err != nil {
		return nil, err
	}

	// check that instant redemption commission is between 0 and 1
	instantRedemptionCommission := sdk.NewDec(instantRedemptionCommissionInt).Quo(sdk.NewDec(10000))
	if instantRedemptionCommission.LT(sdk.ZeroDec()) || instantRedemptionCommission.GT(sdk.OneDec()) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Aborting instant redemption -- instant redemption commission must be between 0 and 1!")
	}

	// construct desired unstaking amount from host zone
	nativeAmount := sdk.NewDecFromInt(msg.Amount).Mul(hostZone.RedemptionRate).RoundInt()
	// safety checks on the coin
	// 	- Redemption amount must be positive
	if !nativeAmount.IsPositive() {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be greater than 0. found: %v", msg.Amount)
	}
	// Compute the fee amount from the by using the redemption commission of the native amount to return to user
	feeAmount := instantRedemptionCommission.Mul(sdk.NewDecFromInt(nativeAmount)).TruncateInt()
	redemptionAmount := nativeAmount.Sub(feeAmount)

	// get the native Ibc coin to return to user, and feeCoin for community pool
	ibcDenom := hostZone.GetIbcDenom()
	redemptionCoin := sdk.NewCoin(ibcDenom, redemptionAmount)
	feeCoin := sdk.NewCoin(ibcDenom, feeAmount)

	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	// 	- Creator owns at least "amount" stAssets
	balance := k.bankKeeper.GetBalance(ctx, sender, stDenom)
	k.Logger(ctx).Info(fmt.Sprintf("Redemption issuer IBCDenom balance: %v%s", balance.Amount, balance.Denom))
	if balance.Amount.LT(msg.Amount) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than redemption amount. redemption amount: %v, balance %v: ", msg.Amount, balance.Amount)
	}

	// Find and subtract this amount from a deposit record if it is big enough
	depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
	pendingDepositRecords := k.RecordsKeeper.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return record.Status == recordstypes.DepositRecord_TRANSFER_QUEUE && record.HostZoneId == hostZone.ChainId
	})
	totalPendingDeposits := k.RecordsKeeper.SumDepositRecords(pendingDepositRecords)
	if nativeAmount.GT(totalPendingDeposits) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAmount, "cannot instant redeem stake an amount %v g.t. pending deposit balance on host zone: %v", nativeAmount, msg.Amount)
	}
	// Subtract all of nativeAmount from one or more pending deposit records
	nativeAmountRemaining := nativeAmount
	for _, depositRecord := range pendingDepositRecords {
		if nativeAmountRemaining.GTE(depositRecord.Amount) {
			nativeAmountRemaining = nativeAmountRemaining.Sub(depositRecord.Amount)
			depositRecord.Amount = sdkmath.ZeroInt()

		} else {
			depositRecord.Amount = depositRecord.Amount.Sub(nativeAmountRemaining)
			nativeAmountRemaining = sdkmath.ZeroInt()
		}
		k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
	}
	bech32ZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.Address, hostZone.ChainId)
	}

	// Send that amount back to user
	err = k.bankKeeper.SendCoins(ctx, bech32ZoneAddress, sender, sdk.NewCoins(redemptionCoin))
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(redemptionCoin) back to user")
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't send %v derivative %s tokens from module account. err: %s", redemptionCoin.Amount, redemptionCoin.Denom, err.Error())
	}
	bech32PoolAddress, err := sdk.AccAddressFromBech32(types.CommunityPoolAccount)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode community pool address %s", types.CommunityPoolAccount)
	}
	// Send the fee amount to the community pool
	err = k.bankKeeper.SendCoins(ctx, bech32ZoneAddress, bech32PoolAddress, sdk.NewCoins(feeCoin))
	if err != nil {
		k.Logger(ctx).Error("Failed to send sdk.NewCoins(feeCoin) to the community pool")
		return nil, sdkerrors.Wrapf(types.ErrInsufficientFunds, "couldn't send %v derivative %s tokens from module account. err: %s", feeCoin.Amount, feeCoin.Denom, err.Error())
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

	k.Logger(ctx).Info(fmt.Sprintf("executed instant redeem stake: %s", msg.String()))
	return &types.MsgInstantRedeemStakeResponse{}, nil
}
