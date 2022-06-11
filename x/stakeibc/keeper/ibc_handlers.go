package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// TODO: uncomment below
// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/Stride-Labs/stride/x/stakeibc/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
// 	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
// 	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
// 	"google.golang.org/protobuf/proto"

// 	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
// )

// Implements core logic for OnAcknowledgementPacket
// TODO(TEST-28): Add ack handling logic for various ICA calls
// TODO(TEST-33): Scope out what to store on different acks (by function call, success/failure)
func (k Keeper) HandleAcknowledgement(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte) error {
	// ack := channeltypes.Acknowledgement_Result{}
	// var eventType string
	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		eventType,
	// 		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
	// 		sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ack)),
	// 	),
	// )
	// err := json.Unmarshal(acknowledgement, &ack)
	// if err != nil {
	// 	ackErr := channeltypes.Acknowledgement_Error{}
	// 	err := json.Unmarshal(acknowledgement, &ackErr)
	// 	if err != nil {
	// 		ctx.EventManager().EmitEvent(
	// 			sdk.NewEvent(
	// 				eventType,
	// 				sdk.NewAttribute(types.AttributeKeyAckError, ackErr.Error),
	// 			),
	// 		)
	// 		k.Logger(ctx).Error("Unable to unmarshal acknowledgement error", "error", err, "data", acknowledgement)
	// 		return err
	// 	}
	// 	k.Logger(ctx).Error("Unable to unmarshal acknowledgement result", "error", err, "remote_err", ackErr, "data", acknowledgement)
	// 	return err
	// }

	

	// txMsgData := &sdk.TxMsgData{}
	// err = proto.Unmarshal(ack.Result, txMsgData)
	// if err != nil {
	// 	k.Logger(ctx).Error("Unable to unmarshal acknowledgement", "error", err, "ack", ack.Result)
	// 	return err
	// }

	// var packetData icatypes.InterchainAccountPacketData
	// err = icatypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &packetData)
	// if err != nil {
	// 	k.Logger(ctx).Error("unable to unmarshal acknowledgement packet data", "error", err, "data", packetData)
	// 	return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	// }

	// msgs, err := icatypes.DeserializeCosmosTx(k.cdc, packetData.Data)
	// if err != nil {
	// 	k.Logger(ctx).Error("unable to decode messages", "err", err)
	// 	return err
	// }

	// for msgIndex, msgData := range txMsgData.Data {
	// 	src := msgs[msgIndex]
	// 	switch msgData.MsgType {
	// 	// staking to validators
	// 	case "/cosmos.staking.v1beta1.MsgDelegate":
	// 		// TODO: Implement!
	// 		continue
	// 	// unstake
	// 	case "/cosmos.staking.v1beta1.MsgUndelegate":
	// 		response := stakingtypes.MsgUndelegateResponse{}
	// 		err := proto.Unmarshal(msgData.Data, &response)
	// 		if err != nil {
	// 			k.Logger(ctx).Error("Unable to unmarshal MsgDelegate response", "error", err)
	// 			return err
	// 		}
	// 		k.Logger(ctx).Debug("Delegated", "response", response)
	// 		// we should update delegation records here.
	// 		if err := k.HandleUndelegate(ctx, src); err != nil {
	// 			return err
	// 		}
	// 		continue
	// 	// withdrawing rewards ()
	// 	case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
	// 		// TODO: Implement!
	// 		continue
	// 	// IBC transfer
	// 	case "/ibc.applications.transfer.v1.MsgTransfer":
	// 		// TODO: Implement!
	// 		continue
	// 	case "cosmos.bank.v1beta1.MsgSend":
	// 		// Construct the transaction
	// 		// TODO(TEST-39): Implement validator selection
	// 		// validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e" // gval2
	// 		// Implement!
	// 		// // set withdraw address to WithdrawAccount
	// 		// setWithdrawAddress := &distributionTypes.MsgSetWithdrawAddress{DelegatorAddress: delegationAccount.GetAddress(), WithdrawAddress: withdrawAccount.GetAddress()}
	// 		// msgs = append(msgs, setWithdrawAddress)
	// 		// // withdraw
	// 		// msgWithdraw := &distributionTypes.MsgWithdrawDelegatorReward{DelegatorAddress: delegationAccount.GetAddress(), ValidatorAddress: validator_address}
	// 		// msgs = append(msgs, msgWithdraw)
	// 	default:
	// 		k.Logger(ctx).Error("Unhandled acknowledgement packet", "type", msgData.MsgType)
	// 	}
	// }

	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		eventType,
	// 		sdk.NewAttribute(types.AttributeKeyAckSuccess, string(ack.Result)),
	// 	),
	// )
	return nil
}

// TODO(TEST-28): Burn stAssets if RedeemStake succeeds
func (k Keeper) HandleUndelegate(ctx sdk.Context, msg sdk.Msg) error {
	// k.Logger(ctx).Info("Received MsgUndelegate acknowledgement")
	// // first, type assertion. we should have stakingtypes.MsgDelegate
	// undelegateMsg, ok := msg.(*stakingtypes.MsgUndelegate)
	// if !ok {
	// 	k.Logger(ctx).Error("unable to cast source message to MsgUndelegate")
	// 	return fmt.Errorf("unable to cast source message to MsgUndelegate")
	// }

	// // Implement!
	// // burn stAssets if successful
	// // return stAssets to user if unsuccessful

	return nil
}
