package stakeibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// NewHandler ...
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		_ = ctx

		switch msg := msg.(type) {
		case *types.MsgLiquidStake:
			res, err := msgServer.LiquidStake(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClearBalance:
			res, err := msgServer.ClearBalance(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRegisterHostZone:
			res, err := msgServer.RegisterHostZone(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRedeemStake:
			res, err := msgServer.RedeemStake(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClaimUndelegatedTokens:
			res, err := msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRebalanceValidators:
			res, err := msgServer.RebalanceValidators(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddValidator:
			res, err := msgServer.AddValidator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgChangeValidatorWeight:
			res, err := msgServer.ChangeValidatorWeight(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteValidator:
			res, err := msgServer.DeleteValidator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRestoreInterchainAccount:
			res, err := msgServer.RestoreInterchainAccount(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateValidatorSharesExchRate:
			res, err := msgServer.UpdateValidatorSharesExchRate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
			// this line is used by starport scaffolding # 1
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// NewAddValidatorHandler creates a new governance Handler for a AddValidatorProposal
func NewAddValidatorProposalHandler(k keeper.Keeper) govtypes.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddValidatorProposal:
			_, err := msgServer.AddValidator(sdk.WrapSDKContext(ctx), &types.MsgAddValidator{
				Creator:  "GOV",
				HostZone: c.HostZone,
				Name:     c.ValidatorName,
				Address:  c.ValidatorAddress,
			})
			if err != nil {
				return err
			}
			return nil

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized addValidator proposal content type: %T", c)
		}
	}
}
