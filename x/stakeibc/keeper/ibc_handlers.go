package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"

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
		// unstake
		case "/cosmos.staking.v1beta1.MsgUndelegate":
			response := stakingtypes.MsgUndelegateResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("Unable to unmarshal MsgUndelegate response", "error", err)
				return err
			}
			k.Logger(ctx).Info("Undelegated", "response", response)
			// we should update delegation records here.
			if err := k.HandleUndelegate(ctx, src, response.CompletionTime); err != nil {
				return err
			}
			continue
		// withdrawing rewards ()
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
			k.Logger(ctx).Info(fmt.Sprintf("epoch number: %d", epochUnbondingRecord.UnbondingEpochNumber))
			if epochUnbondingRecord.UnbondingEpochNumber == dayEpochTracker.EpochNumber {
				k.Logger(ctx).Error("epochUnbondingRecord.UnbondingEpochNumber == dayEpochTracker.EpochNumber")
				continue
			}
			hostZoneUnbonding := epochUnbondingRecord.HostZoneUnbondings[zone.ChainId]
			// NOTE: at the beginning of the epoch we mark all PENDING_TRANSFER HostZoneUnbondingRecords as UNBONDED
			// so that they're retried if the transfer fails
			if hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_PENDING_TRANSFER {
				k.Logger(ctx).Error("hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_PENDING_TRANSFER")
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

func (k Keeper) HandleUndelegate(ctx sdk.Context, msg sdk.Msg, completionTime time.Time) error {
	k.Logger(ctx).Info("Received MsgUndelegate acknowledgement")
	// first, type assertion. we should have stakingtypes.MsgDelegate
	undelegateMsg, ok := msg.(*stakingtypes.MsgUndelegate)
	_ = undelegateMsg
	if !ok {
		k.Logger(ctx).Error("unable to cast source message to MsgUndelegate")
		return fmt.Errorf("unable to cast source message to MsgUndelegate")
	}

	// Check if the unbonding message was for a delegate account (msg.DelegatorAddress)
	zone, err := k.GetHostZoneFromHostDenom(ctx, undelegateMsg.Amount.Denom)
	if err != nil {
		return err
	}
	if undelegateMsg.DelegatorAddress != zone.GetDelegationAccount().GetAddress() {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "Undelegate message was not for a delegation account")
	}

	
	undelegateAmt := undelegateMsg.Amount.Amount.Int64()
	if undelegateAmt <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "Undelegate amount must be positive")
	}
	success := k.AddDelegationToValidator(ctx, *zone, undelegateMsg.ValidatorAddress, -undelegateAmt)
	if !success {
		return sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to add delegation to validator")
	}
	zone.StakedBal -= undelegateAmt
	k.SetHostZone(ctx, *zone)

	epochTracker, found := k.GetEpochTracker(ctx, "day")
	if !found {
		return sdkerrors.Wrapf(types.ErrInvalidLengthEpochTracker, "no number for epoch (%s)", "day")
	}
	currentEpochNumber := epochTracker.GetEpochNumber()
	for _, epochUnbonding := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		if epochUnbonding.UnbondingEpochNumber == currentEpochNumber {
			continue
		}
		hostZoneRecord, found := epochUnbonding.HostZoneUnbondings[zone.ChainId]
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("Host zone unbonding record not found for hostZoneId %s in epoch %d", zone.ChainId, epochUnbonding.UnbondingEpochNumber))
			continue
		}
		if hostZoneRecord.Status != recordstypes.HostZoneUnbonding_BONDED {
			continue
		}
		hostZoneRecord.Status = recordstypes.HostZoneUnbonding_UNBONDED
		hostZoneRecord.UnbondingTime = cast.ToUint64(completionTime.UnixNano())
		k.Logger(ctx).Info(fmt.Sprintf("Set unbonding time to %v for host zone %s's unbonding for %d%s", completionTime, zone.ChainId, undelegateMsg.Amount.Amount.Int64(), undelegateMsg.Amount.Denom))
		// save back the altered SetEpochUnbondingRecord
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbonding)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Total supply %s", k.bankKeeper.GetSupply(ctx, "stuatom")))
	return nil
}
