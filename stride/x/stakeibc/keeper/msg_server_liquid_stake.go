package keeper

import (
	"context"
	"strconv"

	"github.com/Stride-labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Init variables
	// get the sender address
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "address invalid")
	}
	// get the coins to send, they need to be in the format {amount}{denom}
	// NOTE: int is an int32 or int64 (depending on machine type) so converting from int32 -> int
	// is safe. The converse is not true.
	coinString := strconv.Itoa(int(msg.Amount)) + msg.Denom
	coins, err := sdk.ParseCoinsNormalized(coinString)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "failed to parse %s coins", coins)
	}
	
	// Safety checks
	// ensure Amount is non-negative, liquid staking 0 or less tokens is invalid
	if msg.Amount < 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be non-negative")
	}
	// check that the token is an IBC token
	isIbcToken := types.IsIBCToken(msg.Denom)
	if !isIbcToken {
		return nil, sdkerrors.Wrapf(types.ErrInvalidToken, "invalid token denom (%s)", msg.Denom)
	}

	

	// deposit `amount` of `denom` token to the stakeibc module
	// NOTE: Should we add an additional check here? This is a pretty important line of code
	// NOTE: If sender doesn't have enough coins, this panics (error is hard to interpret)
	sdkerror := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins)
	if sdkerror != nil {
		panic(sdkerror)
	}

	// mint user `amount` of the corresponding stAsset
	// NOTE: We should ensure that denoms are unique - we don't want anyone spoofing denoms
	err = k.MintStAsset(ctx, sender, msg.Amount, msg.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "failed to mint stAssets to user")
	}

	return &types.MsgLiquidStakeResponse{}, nil
}

func (k msgServer) MintStAsset(ctx sdk.Context, sender sdk.AccAddress, amount int32, denom string) error {
	// repeat safety checks from LiquidStake in case MintStAsset is called from another site
	// ensure Amount is non-negative, liquid staking 0 or less tokens is invalid
	if amount < 1 {
		return nil
	}
	// check that the token is an IBC token
	isIbcToken := types.IsIBCToken(denom)
	if !isIbcToken {
		return nil
	}

	// NOTE: should we pass in a zone to this function and pull the stAssetDenom off of that object?
	// get the asset type to transfer
	prefix := "st"
	// get the denom of the st asset type to transfer
	stAssetDenom := prefix + denom
	
	// TODO(TEST-7): Add an exchange rate here! What object should we store the exchange rate on?
	// How can we ensure that the exchange rate is not manipulated?
	coinString := strconv.Itoa(int(amount)) + stAssetDenom
	stCoins, err := sdk.ParseCoinsNormalized(coinString)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Failed to parse coins")
	}

	// mint new coins of the asset type
	// MintCoins creates new coins from thin air and adds it to the module account.
	// It will panic if the module account does not exist or is unauthorized.
	sdkerror := k.bankKeeper.MintCoins(ctx, types.ModuleName, stCoins)
	if sdkerror != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Failed to mint coins")
	}
	// transfer those coins to the user
	sdkerror = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, stCoins)
	if sdkerror != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Failed to send coins from module to account")
	}
	return nil
}
