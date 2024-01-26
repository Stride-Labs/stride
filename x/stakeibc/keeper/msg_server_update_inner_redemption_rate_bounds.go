package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/stakeibc/types"
)

func (k msgServer) UpdateInnerRedemptionRateBounds(goCtx context.Context, msg *types.MsgUpdateInnerRedemptionRateBounds) (*types.MsgUpdateInnerRedemptionRateBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: we're intentionally not checking the zone is halted
	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	// Get the wide bounds
	outerMinSafetyThreshold, outerMaxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	innerMinSafetyThreshold := msg.MinInnerRedemptionRate
	innerMaxSafetyThreshold := msg.MaxInnerRedemptionRate

	// Confirm the inner bounds are within the outer bounds
	if innerMinSafetyThreshold.LT(outerMinSafetyThreshold) {
		errMsg := fmt.Sprintf("inner min safety threshold (%s) is less than outer min safety threshold (%s)", innerMinSafetyThreshold, outerMinSafetyThreshold)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds, errMsg)
	}

	if innerMaxSafetyThreshold.GT(outerMaxSafetyThreshold) {
		errMsg := fmt.Sprintf("inner max safety threshold (%s) is greater than outer max safety threshold (%s)", innerMaxSafetyThreshold, outerMaxSafetyThreshold)
		k.Logger(ctx).Error(errMsg)
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds, errMsg)
	}

	// Set the inner bounds on the host zone
	zone.MinInnerRedemptionRate = innerMinSafetyThreshold
	zone.MaxInnerRedemptionRate = innerMaxSafetyThreshold

	k.SetHostZone(ctx, zone)

	return &types.MsgUpdateInnerRedemptionRateBoundsResponse{}, nil
}
