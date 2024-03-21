package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v20/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v20/x/stakedym/types"

	"github.com/Stride-Labs/stride/v20/x/icacallbacks"
)

func (k Keeper) ArchiveFailedTransferRecord(ctx sdk.Context, recordId uint64) error {
	// Mark the record as a failed transfer
	delegationRecord, found := k.GetDelegationRecord(ctx, recordId)
	if !found {
		return types.ErrDelegationRecordNotFound.Wrapf("delegation record not found for %d", recordId)
	}
	delegationRecord.Status = types.TRANSFER_FAILED
	k.ArchiveDelegationRecord(ctx, delegationRecord)

	return nil
}

// OnTimeoutPacket: Delete the DelegationRecord
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	recordId, recordIdFound := k.GetTransferInProgressRecordId(ctx, packet.SourceChannel, packet.Sequence)
	if !recordIdFound {
		return nil
	}

	err := k.ArchiveFailedTransferRecord(ctx, recordId)
	if err != nil {
		return err
	}

	// Clean up the callback store
	k.RemoveTransferInProgressRecordId(ctx, packet.SourceChannel, packet.Sequence)

	return nil
}

// OnAcknowledgementPacket success: Update the DelegationRecord's status to DELEGATION_QUEUE
// OnAcknowledgementPacket failure: Delete the DelegationRecord
func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	recordId, recordIdFound := k.GetTransferInProgressRecordId(ctx, packet.SourceChannel, packet.Sequence)
	if !recordIdFound {
		return nil
	}

	// Parse whether the ack was successful or not
	isICATx := false
	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, k.Logger(ctx), acknowledgement, isICATx)
	if err != nil {
		return err
	}

	// Grab the delegation record
	record, found := k.GetDelegationRecord(ctx, recordId)
	if !found {
		return errorsmod.Wrapf(err, "record not found for record id %d", recordId)
	}

	// If the ack was successful, update the record id to DELEGATION_QUEUE
	if ackResponse.Status == icacallbacktypes.AckResponseStatus_SUCCESS {
		record.Status = types.DELEGATION_QUEUE
		k.SetDelegationRecord(ctx, record)
	} else {
		// Otherwise there must be an error, so archive the record
		err := k.ArchiveFailedTransferRecord(ctx, recordId)
		if err != nil {
			return err
		}
	}

	// Clean up the callback store
	k.RemoveTransferInProgressRecordId(ctx, packet.SourceChannel, packet.Sequence)

	return nil
}
