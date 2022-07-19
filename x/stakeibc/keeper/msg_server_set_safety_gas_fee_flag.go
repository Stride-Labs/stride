package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SetSafetyGasFeeFlag(goCtx context.Context, msg *types.MsgSetSafetyGasFeeFlag) (*types.MsgSetSafetyGasFeeFlagResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	safetyGasFeeFlag, found := k.GetSafetyGasFeeFlag(ctx)
	if !found {
		return nil, fmt.Errorf("unable to fetch safety gas fee flag")
	}
	safetyGasFeeFlag.Enabled = msg.IsEnabled
	k.SetterSafetyGasFeeFlag(ctx, safetyGasFeeFlag)

	return &types.MsgSetSafetyGasFeeFlagResponse{}, nil
}
