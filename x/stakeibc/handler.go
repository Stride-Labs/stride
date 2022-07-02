package stakeibc

import (
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var valid_addrs = map[string]bool{
	"stride15cl9pauj7cqt4lhyrj4snq50gu9u67ese3tvpe": true,
}

func verify_sender(senderAddr string) bool {
	/*
		Verifies if the given address is allowed to run priviledged commands
	*/
	_, ok := valid_addrs[senderAddr]
	return ok
}

// NewHandler ...
func NewHandler(k keeper.Keeper) sdk.Handler {
	WHITELIST := make(map[string]bool)
	WHITELIST["stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"] = true
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
			
		// NOT WHITELISTED!
		case *types.MsgLiquidStake:
			res, err := msgServer.LiquidStake(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRedeemStake:
			res, err := msgServer.RedeemStake(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		// WHITELISTED!
		case *types.MsgRegisterAccount:
			if !WHITELIST[msg.Owner] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.RegisterAccount(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSubmitTx:
			if !WHITELIST[msg.Owner] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.SubmitTx(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRegisterHostZone:
			if !WHITELIST[msg.Creator] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.RegisterHostZone(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClaimUndelegatedTokens:
			if !WHITELIST[msg.Creator] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRebalanceValidators:
			if !WHITELIST[msg.Creator] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			if !WHITELIST[msg.Creator] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.RebalanceValidators(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddValidator:
			if !WHITELIST[msg.Creator] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.AddValidator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgChangeValidatorWeight:
			if !WHITELIST[msg.Creator] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.ChangeValidatorWeight(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteValidator:
			if !WHITELIST[msg.Creator] {
				return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address not whitelisted")
			}
			res, err := msgServer.DeleteValidator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
			// this line is used by starport scaffolding # 1
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
