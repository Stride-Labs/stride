package keeper

import (
	"fmt"

	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/x/records/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func (k Keeper) MarshalTransferCallbackArgs(ctx sdk.Context, delegateCallback types.TransferCallback) ([]byte, error) {
	out, err := proto.Marshal(&delegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalTransferCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

func (k Keeper) UnmarshalTransferCallbackArgs(ctx sdk.Context, delegateCallback []byte) (*types.TransferCallback, error) {
	unmarshalledTransferCallback := types.TransferCallback{}
	if err := proto.Unmarshal(delegateCallback, &unmarshalledTransferCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalTransferCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &unmarshalledTransferCallback, nil
}

func TransferCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
	k.Logger(ctx).Info("TransferCallback executing", "packet", packet)
	if ack.GetError() != "" {
		k.Logger(ctx).Error(fmt.Sprintf("TransferCallback does not handle errors %s", ack.GetError()))
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "TransferCallback does not handle errors: %s", ack.GetError())
	}
	if ack == nil {
		// timeout
		k.Logger(ctx).Error(fmt.Sprintf("TransferCallback timeout, ack is nil, packet %v", packet))
		return nil
	}

	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error unmarshalling packet  %v", err.Error()))
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("TransferCallback unmarshalled FungibleTokenPacketData %v", data))

	// deserialize the args
	transferCallbackData, err := k.UnmarshalTransferCallbackArgs(ctx, args)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrUnmarshalFailure, "cannot unmarshal transfer callback args: %s", err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("TransferCallback %v", transferCallbackData))
	depositRecord, found := k.GetDepositRecord(ctx, transferCallbackData.DepositRecordId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("TransferCallback deposit record not found, packet %v", packet))
		return sdkerrors.Wrapf(types.ErrUnknownDepositRecord, "deposit record not found %d", transferCallbackData.DepositRecordId)
	}
	depositRecord.Status = types.DepositRecord_STAKE
	k.SetDepositRecord(ctx, depositRecord)
	k.Logger(ctx).Info(fmt.Sprintf("\t [IBC-TRANSFER] Deposit record updated: {%v}", depositRecord.Id))
	k.Logger(ctx).Info(fmt.Sprintf("[IBC-TRANSFER] success to %s", depositRecord.HostZoneId))
	return nil
}
