package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Init variables
	// deposit `amount` of `denom` token to the stakeibc module
	// NOTE: Should we add an additional check here? This is a pretty important line of code
	// NOTE: If sender doesn't have enough inCoin, this errors (error is hard to interpret)
	// check that hostZone is registered
	// strided tx stakeibc liquid-stake 100 uatom
	hostZone, err := k.GetHostZoneFromHostDenom(ctx, msg.HostDenom)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found for denom (%s)", msg.HostDenom))
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "no host zone found for denom (%s)", msg.HostDenom)
	}
	// get the sender address
	sender, _ := sdk.AccAddressFromBech32(msg.Creator)
	// get the coins to send, they need to be in the format {amount}{denom}
	// is safe. The converse is not true.
	ibcDenom := hostZone.GetIbcDenom()
	coinString := msg.Amount.String() + ibcDenom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to parse coin (%s)", coinString))
		return nil, sdkerrors.Wrapf(err, "failed to parse coin (%s)", coinString)
	}

	// Creator owns at least "amount" of inCoin
	balance := k.bankKeeper.GetBalance(ctx, sender, ibcDenom)
	if balance.IsLT(inCoin) {
		k.Logger(ctx).Error(fmt.Sprintf("balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount))
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount)
	}
	// check that the token is an IBC token
	isIbcToken := types.IsIBCToken(ibcDenom)
	if !isIbcToken {
		k.Logger(ctx).Error("invalid token denom - denom is not an IBC token (%s)", ibcDenom)
		return nil, sdkerrors.Wrapf(types.ErrInvalidToken, "denom is not an IBC token (%s)", ibcDenom)
	}

	// safety check: redemption rate must be above safety threshold
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, *hostZone)
	if !rateIsSafe || (err != nil) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed. hostZone: %s, err: %s", hostZone.String(), err.Error())
		return nil, sdkerrors.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}

	bech32ZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return nil, fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.Address, hostZone.ChainId)
	}
	err = k.bankKeeper.SendCoins(ctx, sender, bech32ZoneAddress, sdk.NewCoins(inCoin))
	if err != nil {
		k.Logger(ctx).Error("failed to send tokens from Account to Module")
		return nil, sdkerrors.Wrap(err, "failed to send tokens from Account to Module")
	}
	// mint user `amount` of the corresponding stAsset
	// NOTE: We should ensure that denoms are unique - we don't want anyone spoofing denoms
	err = k.MintStAssetAndTransfer(ctx, sender, msg.Amount, msg.HostDenom)
	if err != nil {
		k.Logger(ctx).Error("failed to send tokens from Account to Module")
		return nil, sdkerrors.Wrapf(err, "failed to mint %s stAssets to user", msg.HostDenom)
	}

	// create a deposit record of these tokens (pending transfer)
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	// Does this use too much gas?
	depositRecord, found := k.RecordsKeeper.GetDepositRecordByEpochAndChain(ctx, strideEpochTracker.EpochNumber, hostZone.ChainId)
	if !found {
		k.Logger(ctx).Error("failed to find deposit record")
		return nil, sdkerrors.Wrapf(sdkerrors.ErrNotFound, fmt.Sprintf("no deposit record for epoch (%d)", strideEpochTracker.EpochNumber))
	}
	depositRecord.Amount = depositRecord.Amount.Add(msg.Amount)
	k.RecordsKeeper.SetDepositRecord(ctx, *depositRecord)

	k.hooks.AfterLiquidStake(ctx, sender)
	return &types.MsgLiquidStakeResponse{}, nil
}

func (k msgServer) MintStAssetAndTransfer(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Int, denom string) error {
	stAssetDenom := types.StAssetDenomFromHostZoneDenom(denom)

	// TODO(TEST-7): Add an exchange rate here! What object should we store the exchange rate on?
	// How can we ensure that the exchange rate is not manipulated?
	hz, _ := k.GetHostZoneFromHostDenom(ctx, denom)
	amountToMint := (sdk.NewDecFromInt(amount).Quo(hz.RedemptionRate)).TruncateInt()
	coinString := amountToMint.String() + stAssetDenom
	stCoins, err := sdk.ParseCoinsNormalized(coinString)
	if err != nil {
		k.Logger(ctx).Error("Failed to parse coins")
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to parse coins %s", coinString)
	}

	// Mints coins to the module account, will error if the module account does not exist or is unauthorized.

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, stCoins)
	if err != nil {
		k.Logger(ctx).Error("Failed to mint coins")
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to mint coins")
	}

	// transfer those coins to the user
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, stCoins)
	if err != nil {
		k.Logger(ctx).Error("Failed to send coins from module to account")
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to send %s from module to account", stCoins.GetDenomByIndex(0))
	}

	k.Logger(ctx).Info(fmt.Sprintf("[MINT ST ASSET] success on %s.", hz.GetChainId()))
	return nil
}
