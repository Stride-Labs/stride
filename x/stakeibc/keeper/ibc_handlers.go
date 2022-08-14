package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
)

// Implements core logic for OnAcknowledgementPacket
// TODO(TEST-28): Add ack handling logic for various ICA calls
// TODO(TEST-33): Scope out what to store on different acks (by function call, success/failure)
func (k Keeper) HandleAcknowledgement(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte) error {
	ack := channeltypes.Acknowledgement_Result{}
	packetSequenceKey := fmt.Sprint(modulePacket.Sequence)
	var eventType string
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ack)),
		),
	)
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
					eventType,
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

	for msgIndex, msgData := range txMsgData.Data {
		src := msgs[msgIndex]
		switch msgData.MsgType {
		case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
			// TODO [TEST-124]: Implement! (lo pri)
			continue
		case "/cosmos.bank.v1beta1.MsgSend":
			response := banktypes.MsgSendResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("unable to unmarshal MsgSend response", "error", err)
				return err
			}
			k.Logger(ctx).Info("Sent", "response", response)

			// we should update delegation records here.
			if err := k.HandleSend(ctx, src, packetSequenceKey); err != nil {
				return err
			}
			continue
		case "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress":
			response := distributiontypes.MsgSetWithdrawAddressResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("unable to unmarshal MsgSend response", "error", err)
				return err
			}
			k.Logger(ctx).Info("WithdrawalAddress set", "response", response)
			continue
		default:
			// TODO: Remove this once everything has been migrated to callbacks
			k.Logger(ctx).Error("Unhandled acknowledgement packet", "type", msgData.MsgType)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyAckSuccess, string(ack.Result)),
		),
	)
	return nil
}

func (k *Keeper) HandleSend(ctx sdk.Context, msg sdk.Msg, sequence string) error {
	// first, type assertion. we should have banktypes.MsgSend
	sendMsg, ok := msg.(*banktypes.MsgSend)
	if !ok {
		k.Logger(ctx).Error("unable to cast source message to MsgSend")
		return fmt.Errorf("unable to cast source message to MsgSend")
	}
	k.Logger(ctx).Info(fmt.Sprintf("Received MsgSend acknowledgement for msg %v", sendMsg))
	coin := sendMsg.Amount[0]
	// Pull host zone
	hostZoneDenom := coin.Denom
	amount := coin.Amount.Int64()
	zone, err := k.GetHostZoneFromHostDenom(ctx, hostZoneDenom)
	if err != nil {
		return err
	}

	// Check to and from addresses
	withdrawalAddress := zone.GetWithdrawalAccount().GetAddress()
	delegationAddress := zone.GetDelegationAccount().GetAddress()
	redemptionAddress := zone.GetRedemptionAccount().GetAddress()

	// Process bank sends that reinvest user rewards
	if sendMsg.FromAddress == withdrawalAddress && sendMsg.ToAddress == delegationAddress {
		// fetch epoch
		strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
		if !found {
			k.Logger(ctx).Error("failed to find epoch")
			return sdkerrors.Wrapf(types.ErrInvalidLengthEpochTracker, "no number for epoch (%s)", epochtypes.STRIDE_EPOCH)
		}
		epochNumber := strideEpochTracker.EpochNumber
		// create a new record so that rewards are reinvested
		record := recordstypes.DepositRecord{
			Id:                 0,
			Amount:             amount,
			Denom:              hostZoneDenom,
			HostZoneId:         zone.ChainId,
			Status:             recordstypes.DepositRecord_STAKE,
			Source:             recordstypes.DepositRecord_WITHDRAWAL_ICA,
			DepositEpochNumber: epochNumber,
		}
		k.RecordsKeeper.AppendDepositRecord(ctx, record)
	} else if sendMsg.FromAddress == redemptionAddress {
		k.Logger(ctx).Error("ACK - sendMsg.FromAddress == redemptionAddress")
		// fetch the record from the packet sequence number, then delete the UserRedemptionRecord and the sequence mapping
		pendingClaims, found := k.GetPendingClaims(ctx, sequence)
		if !found {
			k.Logger(ctx).Error("failed to find pending claim")
			return sdkerrors.Wrapf(types.ErrRecordNotFound, "no pending claim found for sequence (%s)", sequence)
		}
		userRedemptionRecordKey, err := k.GetUserRedemptionRecordKeyFromPendingClaims(ctx, pendingClaims)
		if err != nil {
			k.Logger(ctx).Error("failed to get user redemption record key from pending claim")
			return err
		}
		_, found = k.RecordsKeeper.GetUserRedemptionRecord(ctx, userRedemptionRecordKey)
		if !found {
			errMsg := fmt.Sprintf("User redemption record %s not found on host zone", userRedemptionRecordKey)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrInvalidUserRedemptionRecord, "could not get user redemption record: %s", userRedemptionRecordKey)
		}
		k.RecordsKeeper.RemoveUserRedemptionRecord(ctx, userRedemptionRecordKey)
		k.RemovePendingClaims(ctx, sequence)
	} else {
		// TODO: Remove this once everything has been migrated to callbacks
		k.Logger(ctx).Error("ACK - sendMsg.FromAddress != withdrawalAddress && sendMsg.FromAddress != delegationAddress && sendMsg.FromAddress != redemptionAddress")

	}

	return nil
}
