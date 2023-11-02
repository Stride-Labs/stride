package stakeibc

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Handles stakeibc transactions
func NewMessageHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		_ = ctx

		switch msg := msg.(type) {
		case *types.MsgLiquidStake:
			res, err := msgServer.LiquidStake(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgLSMLiquidStake:
			res, err := msgServer.LSMLiquidStake(sdk.WrapSDKContext(ctx), msg)
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
		case *types.MsgAddValidators:
			res, err := msgServer.AddValidators(sdk.WrapSDKContext(ctx), msg)
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
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// Handles stakeibc gov proposals
func NewStakeibcProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddValidatorsProposal:
			return k.AddValidatorsProposal(ctx, c)
		case *types.ToggleLSMProposal:
			return k.ToggleLSMProposal(ctx, c)

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized stakeibc proposal content type: %T", c)
		}
	}
}
