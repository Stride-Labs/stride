package keeper

import (
	"context"
	"strconv"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Init variables
	// deposit `amount` of `denom` token to the stakeibc module
	// NOTE: Should we add an additional check here? This is a pretty important line of code
	// NOTE: If sender doesn't have enough inCoin, this panics (error is hard to interpret)
	// check that hostZone is registered
	// strided tx stakeibc liquid-stake 100 uatom
	hostDenom := msg.Denom
	hostZone, hostZoneFound := k.GetHostZoneFromDenom(ctx, msg.Denom)
	if !hostZoneFound {
		k.Logger(ctx).Info("Host Zone not found for denom (%s)", hostDenom)
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "Host Zone not found for denom (%s)", hostDenom)
	}
	// get the sender address
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		k.Logger(ctx).Info("Invalid address")
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "address invalid")
	}
	// get the coins to send, they need to be in the format {amount}{denom}
	// NOTE: int is an int32 or int64 (depending on machine type) so converting from int32 -> int
	// is safe. The converse is not true.
	ibcDenom := hostZone.GetIBCDenom()
	coinString := strconv.Itoa(int(msg.Amount)) + ibcDenom
	inCoin, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "failed to parse %s inCoin", inCoin)
	}

	// Safety checks
	// ensure Amount is positive, liquid staking 0 or less tokens is invalid
	if !inCoin.IsPositive() {
		k.Logger(ctx).Info("amount must be positive")
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be positive")
	}
	// Creator owns at least "amount" of inCoin
	balance := k.bankKeeper.GetBalance(ctx, sender, ibcDenom)
	if balance.IsLT(inCoin) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "balance is lower than staking amount. staking amount: %s, balance %s: ", balance.Amount, msg.Amount)
	}
	// check that the token is an IBC token
	isIbcToken := types.IsIBCToken(ibcDenom)
	if !isIbcToken {
		k.Logger(ctx).Info("invalid token denom")
		return nil, sdkerrors.Wrapf(types.ErrInvalidToken, "invalid token denom (%s)", ibcDenom)
	}

	sdkerror := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(inCoin))
	if sdkerror != nil {
		k.Logger(ctx).Error("failed to send tokens from Account to Module")
		panic(sdkerror)
	}
	// mint user `amount` of the corresponding stAsset
	// NOTE: We should ensure that denoms are unique - we don't want anyone spoofing denoms
	err = k.MintStAsset(ctx, sender, msg.Amount, hostDenom)
	if err != nil {
		k.Logger(ctx).Info("failed to send tokens from Account to Module")
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "failed to mint stAssets to user")
	}
	// create a deposit record of these tokens
	depositRecord := types.NewDepositRecord(msg.Amount, hostDenom, hostZone.ChainId,
		sender.String(), types.DepositRecord_RECEIPT)
	k.AppendDepositRecord(ctx, *depositRecord)

	return &types.MsgLiquidStakeResponse{}, nil
}

func (k msgServer) MintStAsset(ctx sdk.Context, sender sdk.AccAddress, amount int32, denom string) error {
	// repeat safety checks from LiquidStake in case MintStAsset is called from another site
	// ensure Amount is positive, liquid staking 0 or less tokens is invalid
	if amount < 1 {
		k.Logger(ctx).Info("Amount to mint must be positive")
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Amount to mint must be positive")
	}

	// NOTE: should we pass in a zone to this function and pull the stAssetDenom off of that object?
	// get the asset type to transfer
	// TODO(TEST-23): store denoms on HostZone
	prefix := "st"
	// get the denom of the st asset type to transfer
	stAssetDenom := prefix + denom

	// TODO(TEST-7): Add an exchange rate here! What object should we store the exchange rate on?
	// How can we ensure that the exchange rate is not manipulated?
	coinString := strconv.Itoa(int(amount)) + stAssetDenom
	stCoins, err := sdk.ParseCoinsNormalized(coinString)
	if err != nil {
		k.Logger(ctx).Info("Failed to parse coins")
		panic(err)
	}

	// mint new coins of the asset type
	// MintCoins creates new coins from thin air and adds it to the module account.
	// It will panic if the module account does not exist or is unauthorized.
	sdkerror := k.bankKeeper.MintCoins(ctx, types.ModuleName, stCoins)
	if sdkerror != nil {
		k.Logger(ctx).Info("Failed to mint coins")
		panic(sdkerror)
	}
	// transfer those coins to the user
	sdkerror = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, stCoins)
	if sdkerror != nil {
		k.Logger(ctx).Info("Failed to send coins from module to account")
		panic(sdkerror)
	}
	return nil
}
