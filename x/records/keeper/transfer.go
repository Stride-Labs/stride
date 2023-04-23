package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v9/x/records/types"
)

// Transfers native tokens, accumulated from normal liquid stakes, to the host zone
// This is invoked epochly
func (k Keeper) IBCTransferNativeTokens(ctx sdk.Context, msg *ibctypes.MsgTransfer, depositRecord types.DepositRecord) error {
	goCtx := sdk.WrapSDKContext(ctx)

	// Submit IBC transfer
	msgTransferResponse, err := k.TransferKeeper.Transfer(goCtx, msg)
	if err != nil {
		return err
	}

	// Build callback data
	transferCallback := types.TransferCallback{
		DepositRecordId: depositRecord.Id,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Marshalling TransferCallback args: %+v", transferCallback))
	marshalledCallbackArgs, err := k.MarshalTransferCallbackArgs(ctx, transferCallback)
	if err != nil {
		return err
	}

	// Store the callback data
	sequence := msgTransferResponse.Sequence
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(msg.SourcePort, msg.SourceChannel, sequence),
		PortId:       msg.SourcePort,
		ChannelId:    msg.SourceChannel,
		Sequence:     sequence,
		CallbackId:   IBCCallbacksID_NativeTransfer,
		CallbackArgs: marshalledCallbackArgs,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Storing callback data: %+v", callback))
	k.ICACallbacksKeeper.SetCallbackData(ctx, callback)

	// update the record state to TRANSFER_IN_PROGRESS
	depositRecord.Status = types.DepositRecord_TRANSFER_IN_PROGRESS
	k.SetDepositRecord(ctx, depositRecord)

	return nil
}
