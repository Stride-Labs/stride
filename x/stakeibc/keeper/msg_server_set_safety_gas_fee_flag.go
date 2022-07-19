package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SetSafetyGasFeeFlag(goCtx context.Context, msg *types.MsgSetSafetyGasFeeFlag) (*types.MsgSetSafetyGasFeeFlagResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgSetSafetyGasFeeFlagResponse{}, nil
}
