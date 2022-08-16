package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k msgServer) ClearBalance(goCtx context.Context, msg *types.MsgClearBalance) (*types.MsgClearBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInvalidHostZone, "chainId: %s", msg.ChainId)
	}
	feeAccount := zone.GetFeeAccount()
	if feeAccount == nil {
		return nil, sdkerrors.Wrapf(types.ErrFeeAccountNotRegistered, "chainId: %s", msg.ChainId)
	}

	// should this be a param? the transfer port _should_ always be the same across zones
	sourcePort := "transfer"
	// Should this be a param?
	// I think as long as we have a timeout on this, it should be hard to attack (even if someone send a tx on a bad channel, it would be reverted relatively quickly)
	sourceChannel := msg.Channel
	coinString := cast.ToString(msg.Amount) + zone.GetHostDenom()
	tokens, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to parse coin (%s)", coinString))
		return nil, sdkerrors.Wrapf(err, "failed to parse coin (%s)", coinString)
	}
	sender := feeAccount.GetAddress()
	// where to store this?
	receiver := "stride19uvw0azm9u0k6vqe4e22cga6kteskdqq3ulj6q"
	feeTransferTimeoutNanos := k.GetParam(ctx, types.KeyFeeTransferTimeoutNanos)
	timeoutTimestamp := cast.ToUint64(ctx.BlockTime().UnixNano()) + feeTransferTimeoutNanos
	msgs := []sdk.Msg{
		&ibctransfertypes.MsgTransfer{
			SourcePort: sourcePort,
			SourceChannel: sourceChannel,
			Token: tokens,
			Sender: sender,
			Receiver: receiver,
			TimeoutTimestamp: timeoutTimestamp,
		},
	}

	connectionId := zone.GetConnectionId()
	
	icaTimeoutNanos := k.GetParam(ctx, types.KeyICATimeoutNanos)
	icaTimeoutNanos = cast.ToUint64(ctx.BlockTime().UnixNano()) + icaTimeoutNanos

	k.SubmitTxs(ctx, connectionId, msgs, *feeAccount, icaTimeoutNanos, "", nil)
	return &types.MsgClearBalanceResponse{}, nil
}

