package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

func (k msgServer) LockupTokens(goCtx context.Context, msg *types.MsgLockupTokens) (*types.MsgLockupTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, msg.Denom)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found for denom (%s)", msg.Denom))
		return nil, sdkerrors.Wrapf(stakeibctypes.ErrInvalidHostZone, "no host zone found for denom (%s)", msg.Denom)
	}

	// get the creator address
	creatorAddr, _ := sdk.AccAddressFromBech32(msg.Creator)
	// get the coins to send, they need to be in the format {amount}{denom}
	// is safe. The converse is not true.

	stDenom := stakeibctypes.StAssetDenomFromHostZoneDenom(msg.Denom)
	coinString := msg.Amount.String() + stDenom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to parse coin (%s)", coinString))
		return nil, sdkerrors.Wrapf(err, "failed to parse coin (%s)", coinString)
	}
	// Creator owns at least "amount" of inCoin
	balance := k.bankKeeper.GetBalance(ctx, creatorAddr, stDenom)
	if balance.IsLT(inCoin) {
		k.Logger(ctx).Error(fmt.Sprintf("balance is lower than lockup amount. lockup amount: %v, balance: %v", msg.Amount, balance.Amount))
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than lockup amount. lockup amount: %v, balance: %v", msg.Amount, balance.Amount)
	}
	// send coins to module account
	// make module account per host zone
	k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, hostZone.LockupAddress, sdk.NewCoins(inCoin))

	// get lockup if exists, create if not
	lockup, found := k.GetLockup(ctx, creatorAddr, msg.Denom)
	if !found {
		lockup = types.NewLockup(creatorAddr, sdk.ZeroInt(), msg.Denom)
	}
	// add amount of locked tokens to lockup and set
	lockup.Amount = lockup.Amount.Add(msg.Amount)

	k.SetLockup(ctx, lockup)

	return &types.MsgLockupTokensResponse{}, nil
}

func (k msgServer) UnlockTokens(goCtx context.Context, msg *types.MsgUnlockTokens) (*types.MsgUnlockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, msg.Denom)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found for denom (%s)", msg.Denom))
		return nil, sdkerrors.Wrapf(stakeibctypes.ErrInvalidHostZone, "no host zone found for denom (%s)", msg.Denom)
	}

	// check if a lockup object exists in the store
	lockup, found := k.GetLockup(ctx, sdk.AccAddress(msg.Creator), msg.Denom)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Lockup for creator %s and host zone %s not found", msg.Creator, msg.Denom))
		return nil, types.ErrLockupNotFound
	}

	// ensure that we have enough tokens to remove
	if lockup.Amount.LTE(msg.Amount) {
		k.Logger(ctx).Error(fmt.Sprintf("lockup amount is lower than requested amount. lockup amount: %v, specified: %v", lockup.Amount, msg.Amount))
		return nil, sdkerrors.Wrapf(types.ErrNotEnoughLockupTokens, "lockup amount is lower than requested amount. lockup amount: %v, specified: %v", lockup.Amount, msg.Amount)
	}

	// subtract tokens from lockup
	lockup.Amount = lockup.Amount.Sub(msg.Amount)

	hostUnbondingTime := hostZone.UnbondingTime
	completionTime := ctx.BlockHeader().Time.Add(hostUnbondingTime)
	k.SetUnlockingRecordEntry(ctx, sdk.AccAddress(msg.Creator), msg.Denom, ctx.BlockHeight(), completionTime, msg.Amount)

	return &types.MsgUnlockTokensResponse{
		CompletionTime: completionTime,
	}, nil
}
