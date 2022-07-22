package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RestoreInterchainAccount(goCtx context.Context, msg *types.MsgRestoreInterchainAccount) (*types.MsgRestoreInterchainAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	owner := types.FormatICAAccountOwner(msg.ChainId, msg.AccountType)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, hostZone.ConnectionId, owner); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to register %s account : %s", msg.AccountType.String(), err))
		return nil, err
	}

	return &types.MsgRestoreInterchainAccountResponse{}, nil
}
