package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

type msgServer struct {
	Keeper
}

const USTRD = "ustrd"

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// User transaction to burn STRD tokens
func (k msgServer) Burn(goCtx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	userAddress, err := sdk.AccAddressFromBech32(msg.Burner)
	if err != nil {
		return nil, err
	}

	// Check if there are enough unlocked tokens to cover the burn amount
	// If there aren't enough locked tokens, but the balance is sufficent, it is likely a vesting account
	hasSufficientTotalBalance := k.bankKeeper.GetBalance(ctx, userAddress, USTRD).Amount.GTE(msg.Amount)
	hasInsufficientUnlockedBalance := k.bankKeeper.SpendableCoins(ctx, userAddress).AmountOf(USTRD).LT(msg.Amount)
	if hasSufficientTotalBalance && hasInsufficientUnlockedBalance {
		if err := k.DowngradeVestingAccount(ctx, userAddress); err != nil {
			return nil, err
		}
	}

	// Send tokens from the user to this burner module
	burnCoin := sdk.NewCoins(sdk.NewCoin(USTRD, msg.Amount))
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, userAddress, types.ModuleName, burnCoin)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to transfer tokens to the burner module account")
	}

	// Burn from the module account
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoin)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to burn coins")
	}

	// Update accounting
	k.IncrementTotalUserStrdBurned(ctx, msg.Amount)
	k.IncrementStrdBurnedByAddress(ctx, userAddress, msg.Amount)

	// Emit burn events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurn,
			sdk.NewAttribute(types.AttributeAmount, burnCoin.String()),
		),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUserBurn,
			sdk.NewAttribute(types.AttributeAddress, msg.Burner),
			sdk.NewAttribute(types.AttributeAmount, burnCoin.String()),
		),
	)

	return &types.MsgBurnResponse{}, nil
}

// User transaction to link a non-stride address
func (k msgServer) Link(goCtx context.Context, msg *types.MsgLink) (*types.MsgLinkResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	strideAddress, err := sdk.AccAddressFromBech32(msg.StrideAddress)
	if err != nil {
		return nil, err
	}

	k.SetLinkedAddress(ctx, strideAddress, msg.LinkedAddress)

	return &types.MsgLinkResponse{}, nil
}
