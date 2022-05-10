package keeper

import (
	"context"

	"github.com/Stride-labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// safety checks
	// ensure Amount is non-negative, liquid staking 0 or less tokens is invalid
	if msg.Amount < 1 {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount must be non-negative")
	}
	// check that the token is an IBC token
	isIbcToken := types.IsIBCToken(msg.Denom)
	if !isIbcToken {
		return nil, sdkerrors.Wrapf(types.ErrInvalidToken, "invalid token denom (%s)", msg.Denom)
	}

	// get the sender address
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "address invalid")
	}
	
	// get the coins to send, they need to be in the format {amount}{denom}
	coinString := string(msg.Amount) + msg.Denom
	coins, err := sdk.ParseCoinsNormalized(coinString)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "failed to parse %s coins", coins)
	}

	// deposit `amount` of `denom` token to the stakeibc module
	// NOTE: Should we add an additional check here? This is a pretty important line of code
	sdkerror := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, coins)
	if sdkerror != nil {
		panic(sdkerror)
	}

	// mint user `amount` of the corresponding stAsset
	// NOTE: We should ensure that denoms are unique - we don't want anyone spoofing denoms
	success := k.MintStAsset(ctx, sender, msg.Amount, msg.Denom)
	if !success {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPanic, "failed to mint stAssets to user")
	}

	return &types.MsgLiquidStakeResponse{}, nil
}

func (k msgServer) MintStAsset(ctx sdk.Context, sender sdk.AccAddress, amount int32, denom string) bool {
	// repeat safety checks from LiquidStake in case MintStAsset is called from another site
	// ensure Amount is non-negative, liquid staking 0 or less tokens is invalid
	if amount < 1 {
		return false
	}
	// check that the token is an IBC token
	isIbcToken := types.IsIBCToken(denom)
	if !isIbcToken {
		return false
	}

	// NOTE: should we pass in a zone to this function and pull the stAssetDenom off of that object?
	// get the asset type to transfer
	prefix := "st"
	// get the denom of the st asset type to transfer
	stAssetDenom := prefix + denom
	
	coinString := string(amount) + stAssetDenom
	stCoins, err := sdk.ParseCoinsNormalized(coinString)
	if err != nil {
		panic(err)
	}

	// mint new coins of the asset type
	// MintCoins creates new coins from thin air and adds it to the module account.
	// It will panic if the module account does not exist or is unauthorized.
	sdkerror := k.bankKeeper.MintCoins(ctx, types.ModuleName, stCoins)
	if sdkerror != nil {
		panic(sdkerror)
	}
	// transfer those coins to the user
	sdkerror = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, stCoins)
	if sdkerror != nil {
		panic(sdkerror)
	}
	return true
}
