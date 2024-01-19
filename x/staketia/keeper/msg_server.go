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
func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgLiquidStakeResponse{}, nil
}

// User transaction to redeem stake stTokens into native tokens
func (k msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgRedeemStakeResponse{}, nil
}

// Operator transaction to confirm a delegation was submitted on the host chain
func (k msgServer) ConfirmDelegation(goCtx context.Context, msg *types.MsgConfirmDelegation) (*types.MsgConfirmDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgConfirmDelegationResponse{}, nil
}

// Operator transaction to confirm an undelegation was submitted on the host chain
func (k msgServer) ConfirmUndelegation(goCtx context.Context, msg *types.MsgConfirmUndelegation) (*types.MsgConfirmUndelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgConfirmUndelegationResponse{}, nil
}

// Operator transaction to confirm unbonded tokens were transferred back to stride
func (k msgServer) ConfirmUnbondedTokenSweep(goCtx context.Context, msg *types.MsgConfirmUnbondedTokenSweep) (*types.MsgConfirmUnbondedTokenSweepResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgConfirmUnbondedTokenSweepResponse{}, nil
}

// Operator transaction to adjust the delegated balance after a validator was slashed
func (k msgServer) AdjustDelegatedBalance(goCtx context.Context, msg *types.MsgAdjustDelegatedBalance) (*types.MsgAdjustDelegatedBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgAdjustDelegatedBalanceResponse{}, nil
}

// Adjusts the inner redemption rate bounds on the host zone
func (k msgServer) UpdateInnerRedemptionRateBounds(goCtx context.Context, msg *types.MsgUpdateInnerRedemptionRateBounds) (*types.MsgUpdateInnerRedemptionRateBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Fetch the zone
	zone, err := k.GetHostZone(ctx)
	if err != nil {
		return nil, err
	}

	// Get the outer bounds
	maxOuterBound := zone.MaxRedemptionRate
	minOuterBound := zone.MinRedemptionRate

	// Confirm the inner bounds are within the outer bounds
	maxInnerBound := msg.MaxInnerRedemptionRate
	minInnerBound := msg.MinInnerRedemptionRate
	if maxInnerBound.GT(maxOuterBound) {
		return nil, types.ErrInvalidBounds
	}
	if minInnerBound.LT(minOuterBound) {
		return nil, types.ErrInvalidBounds
	}

	// Set the inner bounds on the host zone
	zone.MinInnerRedemptionRate = minInnerBound
	zone.MaxInnerRedemptionRate = maxInnerBound

	// Update the host zone
	k.SetHostZone(ctx, zone)

	return &types.MsgUpdateInnerRedemptionRateBoundsResponse{}, nil
}

// Unhalts the host zone if redemption rates were exceeded
func (k msgServer) ResumeHostZone(goCtx context.Context, msg *types.MsgResumeHostZone) (*types.MsgResumeHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO [sttia]
	_ = ctx
	return &types.MsgResumeHostZoneResponse{}, nil
}
