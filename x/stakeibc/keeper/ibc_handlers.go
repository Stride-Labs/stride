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
				k.Logger(ctx).Error(fmt.Sprintf("failed to get user redemption record from key %s", userRedemptionRecordKey))
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
			k.Logger(ctx).Error(fmt.Sprintf("Unable to unmarshal acknowledgement error: %s, data: %s", err.Error(), acknowledgement))
			return err
		}
		k.Logger(ctx).Error(fmt.Sprintf("Unable to unmarshal acknowledgement result, error: %s, remote_err: %v, data: %v", err.Error(), ackErr, acknowledgement))
		return err
	}

	txMsgData := &sdk.TxMsgData{}
	err = proto.Unmarshal(ack.Result, txMsgData)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Unable to unmarshal acknowledgement, error: %s, ack result: %v", err, ack.Result))
		return err
	}

	var packetData icatypes.InterchainAccountPacketData
	err = icatypes.ModuleCdc.UnmarshalJSON(modulePacket.GetData(), &packetData)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to unmarshal acknowledgement packet data, error: %s, data: %s", err, packetData))
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	msgs, err := icatypes.DeserializeCosmosTx(k.cdc, packetData.Data)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to decode messages, err: %s", err))
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
				k.Logger(ctx).Error(fmt.Sprintf("unable to unmarshal MsgSend response, error: %s", err.Error()))
				return err
			}
			k.Logger(ctx).Info(fmt.Sprintf("Sent response: %v", response))

			// we should update delegation records here.
			if err := k.HandleSend(ctx, src, packetSequenceKey); err != nil {
				return err
			}
			continue
		case "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress":
			response := distributiontypes.MsgSetWithdrawAddressResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("unable to unmarshal MsgSend response, error: %s", err.Error()))
				return err
			}
			k.Logger(ctx).Info(fmt.Sprintf("WithdrawalAddress set response: %v", response))
			continue
		default:
			k.Logger(ctx).Error(fmt.Sprintf("Unhandled acknowledgement packet type: %v", msgData.MsgType))
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
	zone, err := k.GetHostZoneFromHostDenom(ctx, hostZoneDenom)
	if err != nil {
		return err
	}

	// Check to and from addresses
	delegationAddress := zone.GetDelegationAccount().GetAddress()
	redemptionAddress := zone.GetRedemptionAccount().GetAddress()

	if sendMsg.FromAddress == delegationAddress && sendMsg.ToAddress == redemptionAddress {
		k.Logger(ctx).Error("ACK - sendMsg.FromAddress == delegationAddress && sendMsg.ToAddress == redemptionAddress")
		dayEpochTracker, found := k.GetEpochTracker(ctx, "day")
		if !found {
			k.Logger(ctx).Error("failed to find epoch day")
			return sdkerrors.Wrapf(types.ErrInvalidLengthEpochTracker, "no number for epoch (%s)", "day")
		}
		epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)

		for _, epochUnbondingRecord := range epochUnbondingRecords {
			k.Logger(ctx).Info(fmt.Sprintf("epoch number: %d", epochUnbondingRecord.EpochNumber))
			if epochUnbondingRecord.EpochNumber == dayEpochTracker.EpochNumber {
				k.Logger(ctx).Error("epochUnbondingRecord.UnbondingEpochNumber == dayEpochTracker.EpochNumber")
				continue
			}
			hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.Id, zone.ChainId)
			if !found {
				k.Logger(ctx).Error("failed to find hostZoneUnbonding")
				continue
			}
			// NOTE: at the beginning of the epoch we mark all PENDING_TRANSFER HostZoneUnbondingRecords as UNBONDED
			// so that they're retried if the transfer fails
			if hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_PENDING_TRANSFER {
				k.Logger(ctx).Error(fmt.Sprintf("hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_PENDING_TRANSFER (%v)", hostZoneUnbonding.Status))
				continue
			}
			hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_TRANSFERRED
			userRedemptionRecords := hostZoneUnbonding.UserRedemptionRecords
			for _, recordId := range userRedemptionRecords {
				userRedemptionRecord, found := k.RecordsKeeper.GetUserRedemptionRecord(ctx, recordId)
				if !found {
					k.Logger(ctx).Error("failed to find user redemption record")
					return sdkerrors.Wrapf(types.ErrRecordNotFound, "no user redemption record found for id (%s)", recordId)
				}
				if userRedemptionRecord.IsClaimable {
					k.Logger(ctx).Info("user redemption record is already claimable")
					continue
				}
				userRedemptionRecord.IsClaimable = true
				k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
				k.SetHostZone(ctx, *zone)
			}
			k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
		}
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
		k.Logger(ctx).Error("ACK - sendMsg.FromAddress != withdrawalAddress && sendMsg.FromAddress != delegationAddress && sendMsg.FromAddress != redemptionAddress")

	}

	return nil
}
