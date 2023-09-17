package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const (
	isPrevented = true
)

func (k msgServer) UndelegateHost(goCtx context.Context, msg *types.MsgUndelegateHost) (*types.MsgUndelegateHostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// undelegateHost is callable only if it has not yet been called and succeeded
	if !isPrevented {
		return nil, errorsmod.Wrapf(types.ErrUndelegateHostNotCallable, "")
	}

	// Get host zone unbonding message by summing up the unbonding records
	if err := k.UndelegateHostEvmos(ctx, msg.Amount); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error initiating host zone unbondings for UndelegateHostEvmos", err.Error()))
	}

	// log: issuing an undelegation to Evmos
	k.Logger(ctx).Info(fmt.Sprintf("Issuing an undelegation to Evmos"))

	return &types.MsgUndelegateHostResponse{}, nil
}
