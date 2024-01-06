package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v16/x/autopilot/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v16/x/icacallbacks/types"
)

func (k Keeper) TransferCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbacktypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	var transferCallback types.TransferCallback
	if err := proto.Unmarshal(args, &transferCallback); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal autopilot transfer callback")
	}

	// If the ack timed-out, retry the transfer
	if ackResponse.Status == icacallbacktypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(fmt.Sprintf("Autopilot outbound transfer timed out, retrying: %+v", packet))
		return k.RetryTransfer(ctx, packet, transferCallback.FallbackAddress)
	}

	// if the ack failed, send to the fallback address
	if ackResponse.Status == icacallbacktypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(fmt.Sprintf("Autopilot outbound transfer failed, sending to fallback address: %+v", packet))
		return k.SendToFallbackAddress(ctx, packet.Data, transferCallback.FallbackAddress)
	}

	// If the ack was successful, no action necessary
	return nil
}

// If there's a timed out packet, we'll infinitely retry the transfer
func (k Keeper) RetryTransfer(ctx sdk.Context, packet channeltypes.Packet, fallbackAddress string) error {
	// If this was from an autopilot action, unmarshal the transfer metadata to get the original transfer info
	var transferMetadata transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.Data, &transferMetadata); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal ICS-20 packet data")
	}

	// Build the token from the transfer metadata
	token, err := k.BuildCoinFromTransferMetadata(transferMetadata)
	if err != nil {
		return err
	}

	// Submit the transfer again with a new timeout
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + transfertypes.DefaultRelativePacketTimeoutTimestamp
	msgTransfer := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    packet.SourceChannel,
		Token:            token,
		Sender:           transferMetadata.Sender,
		Receiver:         transferMetadata.Receiver,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             transferMetadata.Memo,
	}
	retryResponse, err := k.transferKeeper.Transfer(ctx, &msgTransfer)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to submit transfer retry of %+v", msgTransfer)
	}
	sequence := retryResponse.Sequence

	// Store the original receiver as the fallback address in case the transfer fails
	transferCallback := types.TransferCallback{
		FallbackAddress: fallbackAddress,
	}
	transferCallbackBz, err := proto.Marshal(&transferCallback)
	if err != nil {
		return err
	}

	callbackData := icacallbacktypes.CallbackData{
		CallbackKey:  icacallbacktypes.PacketID(transfertypes.PortID, packet.SourceChannel, sequence),
		PortId:       transfertypes.PortID,
		ChannelId:    packet.SourceChannel,
		Sequence:     sequence,
		CallbackId:   IBCCallbackID_Transfer,
		CallbackArgs: transferCallbackBz,
	}
	k.ibccallbacksKeeper.SetCallbackData(ctx, callbackData)

	return nil
}

// Build an sdk.Coin type from the transfer metadata which includes strings for the amount and denom
func (k Keeper) BuildCoinFromTransferMetadata(transferMetadata transfertypes.FungibleTokenPacketData) (coin sdk.Coin, err error) {
	amount, ok := sdkmath.NewIntFromString(transferMetadata.Amount)
	if !ok {
		return coin, fmt.Errorf("unable to parse amount from transfer packet: %v", transferMetadata)
	}
	coin = sdk.NewCoin(transferMetadata.Denom, amount)
	return coin, nil
}

// In the event of an ack error after a outbound transfer, we'll have to bank send to a fallback address
func (k Keeper) SendToFallbackAddress(ctx sdk.Context, packetData []byte, fallbackAddress string) error {
	// First unmarshal the transfer metadata to get the sender/reciever, and token amount/denom
	var transferMetadata transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packetData, &transferMetadata); err != nil {
		return err
	}

	// Pull out the original sender of the transfer which will also be the bank sender
	sender := transferMetadata.Sender
	senderAccount, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address")
	}
	fallbackAccount, err := sdk.AccAddressFromBech32(fallbackAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid fallback address")
	}

	// Build the token from the transfer metadata
	token, err := k.BuildCoinFromTransferMetadata(transferMetadata)
	if err != nil {
		return err
	}

	// Finally send to the fallback account
	if err := k.bankKeeper.SendCoins(ctx, senderAccount, fallbackAccount, sdk.NewCoins(token)); err != nil {
		return err
	}

	return nil
}
