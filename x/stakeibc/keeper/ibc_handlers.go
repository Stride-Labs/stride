package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
)

// Implements core logic for OnAcknowledgementPacket
// TODO(TEST-28): Add ack handling logic for various ICA calls
// TODO(TEST-33): Scope out what to store on different acks (by function call, success/failure)
func (k Keeper) HandleAcknowledgement(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte) error {
	ack := channeltypes.Acknowledgement_Result{}
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
		// staking to validators
		case "/cosmos.staking.v1beta1.MsgDelegate":
			// TODO: Implement! - Handle staking by DELETING deposit records
			continue
		// unstake
		case "/cosmos.staking.v1beta1.MsgUndelegate":
			response := stakingtypes.MsgUndelegateResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("Unable to unmarshal MsgDelegate response", "error", err)
				return err
			}
			k.Logger(ctx).Debug("Delegated", "response", response)
			// we should update delegation records here.
			if err := k.HandleUndelegate(ctx, src); err != nil {
				return err
			}
			continue
		// withdrawing rewards ()
		case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
			// TODO: Implement!
			continue
		// IBC transfer - update the DepositRecord status!
		case "/ibc.applications.transfer.v1.MsgTransfer":
			k.Logger(ctx).Debug("DOGE")
			response := ibctransfertypes.MsgTransferResponse{}
			err := proto.Unmarshal(msgData.Data, &response)
			if err != nil {
				k.Logger(ctx).Error("unable to unmarshal MsgTransfer response", "error", err)
				return err
			}
			k.Logger(ctx).Debug("MsgTranfer acknowledgement received")
			if err := k.HandleIBCTransfer(ctx, src); err != nil {
				return err
			}
			continue
		case "cosmos.bank.v1beta1.MsgSend":
			// Construct the transaction
			// TODO(TEST-39): Implement validator selection
			// validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e" // gval2
			// Implement!
			// WITHDRAW REWARDS
			// TODO(TEST-5): Update rewards records to STATUS STAKE
			// // set withdraw address to WithdrawAccount
			// setWithdrawAddress := &distributionTypes.MsgSetWithdrawAddress{DelegatorAddress: delegationAccount.GetAddress(), WithdrawAddress: withdrawAccount.GetAddress()}
			// msgs = append(msgs, setWithdrawAddress)
			// // withdraw
			// msgWithdraw := &distributionTypes.MsgWithdrawDelegatorReward{DelegatorAddress: delegationAccount.GetAddress(), ValidatorAddress: validator_address}
			// msgs = append(msgs, msgWithdraw)
			// STAKE REWARDS
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

func (k Keeper) HandleIBCTransfer(ctx sdk.Context, msg sdk.Msg) error {
	k.Logger(ctx).Info("Received IbcTransfer acknowledgement")
	// first, type assertion. we should have ibctransfertypes.MsgTransfer
	ibcTransferMsg, ok := msg.(*ibctransfertypes.MsgTransfer)
	if !ok {
		k.Logger(ctx).Error("unable to cast source message to MsgTransfer")
		return fmt.Errorf("unable to cast source message to MsgTransfer")
	}

	// TODO: uppercase account keeper in stakeibc keeper and make this call
	// check if destination is interchainstaking module account
	// if sMsg.Receiver != k.AccountKeeper.GetModuleAddress(types.ModuleName).String() {
	// 	k.Logger(ctx).Error("msgTransfer to unknown account!")
	// 	return nil
	// }


	
	// fetch the deposit record based on the amount
	// NOTE: there must be a better way to do this, in it's current form it feels somewhat unsafe
	// we could add some "dust" to each transfer / deposit record to make this less susceptible to attacks
	// but it's a hack
	record, found := k.GetTransferDepositRecordByAmount(ctx, ibcTransferMsg.Token.Amount.Int64())
	if !found {
		k.Logger(ctx).Error("No record found for %s", ibcTransferMsg.Token.Amount)
		return fmt.Errorf("No record found for %s", ibcTransferMsg.Token.Amount)
	}
	// update the record
	record.Status = types.DepositRecord_STAKE
	k.SetDepositRecord(ctx, *record)
	// set the deposit record state to STAKE

	return nil
}

// TODO(TEST-28): Burn stAssets if RedeemStake succeeds
func (k Keeper) HandleUndelegate(ctx sdk.Context, msg sdk.Msg) error {
	k.Logger(ctx).Info("Received MsgUndelegate acknowledgement")
	// first, type assertion. we should have stakingtypes.MsgDelegate
	undelegateMsg, ok := msg.(*stakingtypes.MsgUndelegate)
	_ = undelegateMsg
	if !ok {
		k.Logger(ctx).Error("unable to cast source message to MsgUndelegate")
		return fmt.Errorf("unable to cast source message to MsgUndelegate")
	}

	// Implement!
	// burn stAssets if successful
	// return stAssets to user if unsuccessful

	return nil
}
