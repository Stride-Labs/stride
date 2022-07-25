package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
)

type ICACallback func(sdk.Context, []sdk.Msg, []byte) error

// HandleAcknowledgementCallback fetches the ICAAccount from the channel and sequence number
// and calls the callback
func (k Keeper) HandleAcknowledgementCallback(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte, connectionId string) error {
	ack := channeltypes.Acknowledgement_Result{}
	callback, found := k.GetCallback(ctx, modulePacket, connectionId)
	if !found {
		k.Logger(ctx).Info("No callback found for acknowledgement")	
	}
	callbackValues := callback.CallbackValues
	packetSequenceKey := fmt.Sprint(modulePacket.Sequence)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"ica_acknowledgement",
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ack)),
		),
	)

	var ackCallback ICACallback
	switch callback.CallbackId {
	case types.ICACallbackType_NO_CALLBACK:
		ackCallback = nil
		k.Logger(ctx).Info("Handling NO_CALLBACK")
	case types.ICACallbackType_MSG_DELEGATE:
		ackCallback = k.HandleDelegateCallback
		k.Logger(ctx).Info("Handling MSG_DELEGATE callback")
	case types.ICACallbackType_MSG_UNDELEGATE:
		ackCallback = k.HandleUndelegateCallback
		k.Logger(ctx).Info("Handling MSG_UNDELEGATE callback")
	case types.ICACallbackType_MSG_SEND:
		ackCallback = k.HandleSendCallback
		k.Logger(ctx).Info("Handling MSG_SEND callback")
	}

	err := json.Unmarshal(acknowledgement, &ack)
	if err != nil {
		ackErr := channeltypes.Acknowledgement_Error{}
		// Clean up any pending claims
		pendingClaims, found := k.GetPendingClaims(ctx, packetSequenceKey)
		if found {
			k.RemovePendingClaims(ctx, packetSequenceKey)
			userRedemptionRecordKey, err := k.GetUserRedemptionRecordKeyFromPendingClaims(ctx, pendingClaims)
			if err != nil {
				k.Logger(ctx).Error("failed to get user redemption record key from pending claim")
				return err
			}
			record, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, userRedemptionRecordKey)
			if !found {
				k.Logger(ctx).Error("failed to get user redemption record from key %s", userRedemptionRecordKey)
				return err
			}
			record.IsClaimable = true
			k.RecordsKeeper.SetUserRedemptionRecord(ctx, record)
		}
		err := json.Unmarshal(acknowledgement, &ackErr)
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"ica_acknowledgement_error",
					sdk.NewAttribute(types.AttributeKeyAckError, ackErr.Error),
				),
			)
			k.Logger(ctx).Error("Unable to unmarshal acknowledgement error", "error", err, "data", acknowledgement)
			return err
		}
		k.Logger(ctx).Error("Unable to unmarshal acknowledgement result", "error", err, "remote_err", ackErr, "data", acknowledgement)
		return err
	}

	txMsgData := &sdk.TxMsgData{}
	err = proto.Unmarshal(ack.Result, txMsgData)
	if err != nil {
		k.Logger(ctx).Error("Unable to unmarshal acknowledgement", "error", err, "ack", ack.Result)
		return err
	}

	var packetData icatypes.InterchainAccountPacketData
	err = icatypes.ModuleCdc.UnmarshalJSON(modulePacket.GetData(), &packetData)
	if err != nil {
		k.Logger(ctx).Error("unable to unmarshal acknowledgement packet data", "error", err, "data", packetData)
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	msgs, err := icatypes.DeserializeCosmosTx(k.cdc, packetData.Data)
	if err != nil {
		k.Logger(ctx).Error("unable to decode messages", "err", err)
		return err
	}

	err = ackCallback(ctx, msgs, callbackValues)
	if err != nil {
		k.Logger(ctx).Error("unable to handle acknowledgement callback", "err", err)
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"ica_acknowledgement_success",
			sdk.NewAttribute(types.AttributeKeyAckSuccess, string(ack.Result)),
		),
	)
	return nil
}


// Callbacks
func (k *Keeper) HandleSendCallback(ctx sdk.Context, msgs []sdk.Msg, callbackValues []byte) error {
	// HandleSendCallback handles an ICA transaction that is composed of bank sends.
	// The following types of transactions can occur
	// 1) Process bank sends that reinvest user rewards
	// 2) Process unbonding transfers from the DelegationAccount to the RedemptionAccount
	// 3) Process unbonding transfers from the RedemptionAccount to user accounts
	return nil
}

func (k *Keeper) HandleDelegateCallback(ctx sdk.Context, msgs []sdk.Msg, callbackValues []byte) error {
	// HandleDelegateCallback handles an ICA transaction that is composed of delegations.
	// Delegations happen on the delegation ICA account.
	// The associated records are stored on the callback.
	// 0) Sum up totalMsgDelegate by iterating over all records
	// 1) Fetch records
	// 2) Update stakedBal
	// 3) Delete DepositRecords
	return nil
}

func (k Keeper) HandleUndelegateCallback(ctx sdk.Context, msgs []sdk.Msg, callbackValues []byte) error {
	// HandleUndelegateCallback handles an ICA transaction that is composed of undelegations.
	// Undelegations happen on the delegation ICA account.
	// 1) Update stakedBal
	// 2) Update records (mark records as UNBONDED)
	return nil
}
