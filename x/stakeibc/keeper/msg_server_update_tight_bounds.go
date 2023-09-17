package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (k msgServer) UpdateTightBounds(goCtx context.Context, msg *types.MsgUpdateTightBounds) (*types.MsgUpdateTightBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Confirm host zone exists
	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	// Get the wide bounds
	outerMinSafetyThreshold, outerMaxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	innerMinSafetyThreshold := msg.MinTightRedemptionRate
	innerMaxSafetyThreshold := msg.MaxTightRedemptionRate

	// Confirm the inner bounds are within the outer bounds
	if innerMinSafetyThreshold.LT(outerMinSafetyThreshold) {
		k.Logger(ctx).Error(fmt.Sprintf("Inner min safety threshold (%s) is less than outer min safety threshold (%s)", innerMinSafetyThreshold, outerMinSafetyThreshold))
		return nil, types.ErrInvalidBounds
	}

	if innerMaxSafetyThreshold.GT(outerMaxSafetyThreshold) {
		k.Logger(ctx).Error(fmt.Sprintf("Inner max safety threshold (%s) is greater than outer max safety threshold (%s)", innerMaxSafetyThreshold, outerMaxSafetyThreshold))
		return nil, types.ErrInvalidBounds
	}

	// Confirm the max is greater than the min
	if innerMaxSafetyThreshold.LTE(innerMinSafetyThreshold) {
		k.Logger(ctx).Error(fmt.Sprintf("Inner max safety threshold (%s) is less than inner min safety threshold (%s)", innerMaxSafetyThreshold, innerMinSafetyThreshold))
		return nil, types.ErrInvalidBounds
	}

	// Set the inner bounds on the host zone
	zone.MinTightRedemptionRate = innerMinSafetyThreshold
	zone.MaxTightRedemptionRate = innerMaxSafetyThreshold
	k.SetHostZone(ctx, zone)

	return &types.MsgUpdateTightBoundsResponse{}, nil
}
