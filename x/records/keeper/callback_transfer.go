package keeper

import (
	"fmt"

	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v9/x/records/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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

func TransferCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	k.Logger(ctx).Info("TransferCallback executing", "packet", packet)

	// deserialize the args
	transferCallbackData, err := k.UnmarshalTransferCallbackArgs(ctx, args)
	if err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "cannot unmarshal transfer callback args: %s", err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("TransferCallback %v", transferCallbackData))
	depositRecord, found := k.GetDepositRecord(ctx, transferCallbackData.DepositRecordId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("TransferCallback deposit record not found, packet %v", packet))
		return errorsmod.Wrapf(types.ErrUnknownDepositRecord, "deposit record not found %d", transferCallbackData.DepositRecordId)
	}

	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		// timeout
		// put record back in the TRANSFER_QUEUE
		depositRecord.Status = types.DepositRecord_TRANSFER_QUEUE
		k.SetDepositRecord(ctx, depositRecord)
		k.Logger(ctx).Error(fmt.Sprintf("TransferCallback timeout, ack is nil, packet %v", packet))
		return nil
	}

	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		// error on host chain
		// put record back in the TRANSFER_QUEUE
		depositRecord.Status = types.DepositRecord_TRANSFER_QUEUE
		k.SetDepositRecord(ctx, depositRecord)
		k.Logger(ctx).Error(fmt.Sprintf("Error  %s", ackResponse.Error))
		return nil
	}

	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error unmarshalling packet  %v", err.Error()))
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("TransferCallback unmarshalled FungibleTokenPacketData %v", data))

	// put the deposit record in the DELEGATION_QUEUE
	depositRecord.Status = types.DepositRecord_DELEGATION_QUEUE
	k.SetDepositRecord(ctx, depositRecord)
	k.Logger(ctx).Info(fmt.Sprintf("\t [IBC-TRANSFER] Deposit record updated: {%v}, status: {%s}", depositRecord.Id, depositRecord.Status.String()))
	k.Logger(ctx).Info(fmt.Sprintf("[IBC-TRANSFER] success to %s", depositRecord.HostZoneId))
	return nil
}
