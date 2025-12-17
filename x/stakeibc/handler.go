package stakeibc

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v31/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v31/x/stakeibc/types"
)

// Handles stakeibc transactions
// TODO: Remove - no longer used since sdk 47
func NewMessageHandler(k keeper.Keeper) baseapp.MsgServiceHandler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		_ = ctx

		switch msg := msg.(type) {
		case *types.MsgLiquidStake:
			res, err := msgServer.LiquidStake(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgLSMLiquidStake:
			res, err := msgServer.LSMLiquidStake(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClearBalance:
			res, err := msgServer.ClearBalance(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRegisterHostZone:
			res, err := msgServer.RegisterHostZone(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRedeemStake:
			res, err := msgServer.RedeemStake(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClaimUndelegatedTokens:
			res, err := msgServer.ClaimUndelegatedTokens(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRebalanceValidators:
			res, err := msgServer.RebalanceValidators(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddValidators:
			res, err := msgServer.AddValidators(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgChangeValidatorWeights:
			res, err := msgServer.ChangeValidatorWeight(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteValidator:
			res, err := msgServer.DeleteValidator(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgRestoreInterchainAccount:
			res, err := msgServer.RestoreInterchainAccount(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateValidatorSharesExchRate:
			res, err := msgServer.UpdateValidatorSharesExchRate(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgCalibrateDelegation:
			res, err := msgServer.CalibrateDelegation(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateInnerRedemptionRateBounds:
			res, err := msgServer.UpdateInnerRedemptionRateBounds(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgResumeHostZone:
			res, err := msgServer.ResumeHostZone(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgCreateTradeRoute:
			res, err := msgServer.CreateTradeRoute(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteTradeRoute:
			res, err := msgServer.DeleteTradeRoute(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateTradeRoute:
			res, err := msgServer.UpdateTradeRoute(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSetCommunityPoolRebate:
			res, err := msgServer.SetCommunityPoolRebate(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgToggleTradeController:
			res, err := msgServer.ToggleTradeController(ctx, msg)
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
