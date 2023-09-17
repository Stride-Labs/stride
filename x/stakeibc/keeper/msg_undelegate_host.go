package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (k msgServer) UndelegateHost(goCtx context.Context, msg *types.MsgUndelegateHost) (*types.MsgUndelegateHostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// log: issuing an undelegation to Evmos
	k.Logger(ctx).Info(fmt.Sprintf("Issuing an undelegation to Evmos"))

	return &types.MsgUndelegateHostResponse{}, nil
}
