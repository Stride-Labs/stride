package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/x/icacallbacks"
	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"
)

func (k Keeper) MarshalClaimCallbackArgs(ctx sdk.Context, claimCallback types.ClaimCallback) ([]byte, error) {
	out, err := proto.Marshal(&claimCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalClaimCallbackArgs(ctx sdk.Context, claimCallback []byte) (*types.ClaimCallback, error) {
	unmarshalledDelegateCallback := types.ClaimCallback{}
	if err := proto.Unmarshal(claimCallback, &unmarshalledDelegateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalClaimCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledDelegateCallback, nil
}

func ClaimCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	if ack == nil {
		// handle timeout
		k.Logger(ctx).Error(fmt.Sprintf("ClaimCallback timeout, ack is nil, packet %v", packet))
		return nil
	}
	
	txMsgData, err := icacallbacks.GetTxMsgData(ctx, *ack, k.Logger(ctx))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to fetch txMsgData, packet %v", packet))
		return sdkerrors.Wrap(icacallbackstypes.ErrTxMsgData, err.Error())
	}
	k.Logger(ctx).Info("ClaimCallback executing", "packet", packet, "txMsgData", txMsgData, "args", args)

	
	// deserialize the args
	claimCallback, err := k.UnmarshalClaimCallbackArgs(ctx, args)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("ClaimCallback %v", claimCallback))
	userClaimRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, claimCallback.GetUserRedemptionRecordId())
	if !found {
		return sdkerrors.Wrapf(types.ErrRecordNotFound, "user redemption record not found %s", claimCallback.GetUserRedemptionRecordId())
	}

	if len(txMsgData.Data) == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("ClaimCallback failed, packet %v", packet))
		// transaction on the host chain failed
		// set UserClaimRecord as claimable
		userClaimRecord.IsClaimable = true
		k.RecordsKeeper.SetUserRedemptionRecord(ctx, userClaimRecord)
		return nil
	}

	// claim successfully processed
	k.RecordsKeeper.RemoveUserRedemptionRecord(ctx, claimCallback.GetUserRedemptionRecordId())
	k.Logger(ctx).Info(fmt.Sprintf("[CLAIM] success on %s", userClaimRecord.GetHostZoneId()))
	return nil
}
