package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// User transaction to liquid stake native tokens into stTokens
func (server msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgLiquidStakeResponse{}, nil
}

// User transaction to redeem stake stTokens into native tokens
func (server msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgRedeemStakeResponse{}, nil
}

// Operator transaction to confirm a delegation was submitted on the host chain
func (server msgServer) ConfirmDelegation(goCtx context.Context, msg *types.MsgConfirmDelegation) (*types.MsgConfirmDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgConfirmDelegationResponse{}, nil
}

// Operator transaction to confirm an undelegation was submitted on the host chain
func (server msgServer) ConfirmUndelegation(goCtx context.Context, msg *types.MsgConfirmUndelegation) (*types.MsgConfirmUndelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgConfirmUndelegationResponse{}, nil
}

// Operator transaction to confirm unbonded tokens were transferred back to stride
func (server msgServer) ConfirmUnbondedTokenSweep(goCtx context.Context, msg *types.MsgConfirmUnbondedTokenSweep) (*types.MsgConfirmUnbondedTokenSweepResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgConfirmUnbondedTokenSweepResponse{}, nil
}

// Operator transaction to adjust the delegated balance after a validator was slashed
func (server msgServer) AdjustDelegatedBalance(goCtx context.Context, msg *types.MsgAdjustDelegatedBalance) (*types.MsgAdjustDelegatedBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgAdjustDelegatedBalanceResponse{}, nil
}

// Adjusts the inner redemption rate bounds on the host zone
func (server msgServer) UpdateInnerRedemptionRateBounds(goCtx context.Context, msg *types.MsgUpdateInnerRedemptionRateBounds) (*types.MsgUpdateInnerRedemptionRateBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgUpdateInnerRedemptionRateBoundsResponse{}, nil
}

// Unhalts the host zone if redemption rates were exceeded
func (server msgServer) ResumeHostZone(goCtx context.Context, msg *types.MsgResumeHostZone) (*types.MsgResumeHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgResumeHostZoneResponse{}, nil
}
