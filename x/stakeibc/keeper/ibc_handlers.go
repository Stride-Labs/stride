package keeper

import (
	"encoding/json"
	"fmt"
	time "time"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"

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

	// store total amount that was delegated
	totalMsgDelegate := int64(0)
	recordIdToDelete := int64(0)
	for msgIndex, msgData := range txMsgData.Data {
		src := msgs[msgIndex]
		switch msgData.MsgType {
		// staking to validators
		case "/cosmos.staking.v1beta1.MsgDelegate":
			response := stakingtypes.MsgDelegateResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("unable to unmarshal MsgDelegate response", "error", err)
				return err
			}
			delegateMsg, ok := src.(*stakingtypes.MsgDelegate)
			if !ok {
				k.Logger(ctx).Error("unable to cast source message to MsgDelegate")
				return fmt.Errorf("unable to cast source message to MsgDelegate")
			}
			totalMsgDelegate += delegateMsg.Amount.Amount.Int64()
		}
	}

	for msgIndex, msgData := range txMsgData.Data {
		src := msgs[msgIndex]
		switch msgData.MsgType {
		// staking to validators
		case "/cosmos.staking.v1beta1.MsgDelegate":
			response := stakingtypes.MsgDelegateResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("unable to unmarshal MsgDelegate response", "error", err)
				return err
			}
			k.Logger(ctx).Info("Delegated", "response", response)
			// we should update delegation records here.
			recordIdToDelete, err = k.HandleDelegate(ctx, src, totalMsgDelegate)
			if err != nil {
				return err
			}
			continue
		// unstake
		case "/cosmos.staking.v1beta1.MsgUndelegate":
			response := stakingtypes.MsgUndelegateResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("Unable to unmarshal MsgUndelegate response", "error", err)
				return err
			}
			k.Logger(ctx).Debug("Undelegated", "response", response)
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

	if recordIdToDelete >= 0 {
		k.RecordsKeeper.RemoveDepositRecord(ctx, uint64(recordIdToDelete))
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
			k.Logger(ctx).Info("failed to find epoch")
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
			DepositEpochNumber: uint64(epochNumber),
		}
		k.RecordsKeeper.AppendDepositRecord(ctx, record)
		// process unbonding transfers from the DelegationAccount to the RedemptionAccount
	} else if sendMsg.FromAddress == delegationAddress && sendMsg.ToAddress == redemptionAddress {
		k.Logger(ctx).Error("ACK - sendMsg.FromAddress == delegationAddress && sendMsg.ToAddress == redemptionAddress")
		dayEpochTracker, found := k.GetEpochTracker(ctx, "day")
		if !found {
			k.Logger(ctx).Info("failed to find epoch day")
			return sdkerrors.Wrapf(types.ErrInvalidLengthEpochTracker, "no number for epoch (%s)", "day")
		}
		epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)

		for _, epochUnbondingRecord := range epochUnbondingRecords {
			k.Logger(ctx).Error(fmt.Sprintf("epoch number: %d", epochUnbondingRecord.Id))
			if epochUnbondingRecord.Id == dayEpochTracker.EpochNumber {
				k.Logger(ctx).Error("epochUnbondingRecord.Id == dayEpochTracker.EpochNumber")
				continue
			}
			// filter out HostZoneUnbondingRecords that are not in a "pending" state
			// this protects against an edge case where a HostZoneUnbondingRecord becomes unbonded after the epoch has been completed
			// but before the ack is received
			hostZoneUnbondings := epochUnbondingRecord.GetHostZoneUnbondings()
			if len(hostZoneUnbondings) == 0 {
				hostZoneUnbondings = make(map[string]*recordstypes.HostZoneUnbonding)
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
					return sdkerrors.Wrapf(types.ErrRecordNotFound, "no user redemption record found for id (%d)", recordId)
				}
				if userRedemptionRecord.IsClaimable == true {
					k.Logger(ctx).Info("user redemption record is already claimable")
					continue
				}
				userRedemptionRecord.IsClaimable = true
				k.RecordsKeeper.SetUserRedemptionRecord(ctx, userRedemptionRecord)
				k.SetHostZone(ctx, *zone)
			}
			k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
		}
	}  else if sendMsg.FromAddress == redemptionAddress {
		k.Logger(ctx).Error("ACK - sendMsg.FromAddress == redemptionAddress")
		// fetch the record from the packet sequence number, then delete the UserRedemptionRecord and the sequence mapping
		pendingClaims, found := k.GetPendingClaims(ctx, sequence)
		if !found {
			k.Logger(ctx).Error("failed to find pending claim")
			return sdkerrors.Wrapf(types.ErrRecordNotFound, "no pending claim found for sequence (%d)", sequence)
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

func (k *Keeper) HandleDelegate(ctx sdk.Context, msg sdk.Msg, totalDelegate int64) (int64, error) {
	k.Logger(ctx).Info("Received MsgDelegate acknowledgement")
	// first, type assertion. we should have stakingtypes.MsgDelegate
	delegateMsg, ok := msg.(*stakingtypes.MsgDelegate)
	if !ok {
		k.Logger(ctx).Error("unable to cast source message to MsgDelegate")
		return -1, fmt.Errorf("unable to cast source message to MsgDelegate")
	}
	// CHECK ZONE
	hostZoneDenom := delegateMsg.Amount.Denom
	amount := delegateMsg.Amount.Amount.Int64()
	zone, err := k.GetHostZoneFromHostDenom(ctx, hostZoneDenom)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to get host zone from host denom %s", hostZoneDenom))
		return -1, err
	}
	record, found := k.RecordsKeeper.GetStakeDepositRecordByAmount(ctx, totalDelegate, zone.ChainId)
	if !found {
		errMsg := fmt.Sprintf("No deposit record found for zone: %s, amount: %d", zone.ChainId, totalDelegate)
		k.Logger(ctx).Error(errMsg)
		return -1, sdkerrors.Wrapf(sdkerrors.ErrNotFound, errMsg)
	}

	// TODO(TEST-112) more safety checks here
	// increment the stakedBal on the hostZome
	k.Logger(ctx).Info(fmt.Sprintf("incrementing stakedBal %d", amount))
	if amount < 0 {
		errMsg := fmt.Sprintf("Balance to stake was negative: %d", amount)
		k.Logger(ctx).Error(errMsg)
		return -1, sdkerrors.Wrapf(sdkerrors.ErrLogic, errMsg)
	} else {
		zone.StakedBal += amount
		success := k.AddDelegationToValidator(ctx, *zone, delegateMsg.ValidatorAddress, amount)
		if !success {
			return 0, sdkerrors.Wrapf(types.ErrValidatorDelegationChg, "Failed to add delegation to validator")
		}
		k.SetHostZone(ctx, *zone)
	}

	return int64(record.Id), nil
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
		hostZoneUnbondings := epochUnbonding.GetHostZoneUnbondings()
		if len(hostZoneUnbondings) == 0 {
			hostZoneUnbondings = make(map[string]*recordstypes.HostZoneUnbonding)
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
		hostZoneRecord.UnbondingTime = uint64(completionTime.UnixNano())
		k.Logger(ctx).Info(fmt.Sprintf("Set unbonding time to %v for host zone %s's unbonding for %d%s", completionTime, zone.ChainId, undelegateMsg.Amount.Amount.Int64(), undelegateMsg.Amount.Denom))
		// save back the altered SetEpochUnbondingRecord
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbonding)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Total supply %s", k.bankKeeper.GetSupply(ctx, "stuatom")))
	return nil
}
