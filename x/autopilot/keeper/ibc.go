package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v16/x/icacallbacks"
	icacallbacktypes "github.com/Stride-Labs/stride/v16/x/icacallbacks/types"
)

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

// If there was a failed ack from an outbound transfer of one of the autopilot actions,
// we'll need to check if there was a fallback address. If one was stored, bank send
// to that fallback address
// If the ack was successful, we should delete the address (if it exists)
func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	// Retrieve the fallback address for the given packet
	// We use the packet source channel here since this will correspond with the channel on Stride
	channelId := packet.SourceChannel
	sequence := packet.Sequence
	fallbackAddress, fallbackAddressFound := k.GetTransferFallbackAddress(ctx, channelId, sequence)

	// If there was no fallback address, there's nothing else to do
	if !fallbackAddressFound {
		return nil
	}

	// Remove the fallback address since the packet is no longer pending
	k.RemoveTransferFallbackAddress(ctx, channelId, sequence)

	// Check whether the ack was successful or was an ack error
	isICATx := false
	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, k.Logger(ctx), packet.Data, isICATx)
	if err != nil {
		return err
	}

	// If successful, no additional action is necessary
	if ackResponse.Status == icacallbacktypes.AckResponseStatus_SUCCESS {
		return nil
	}

	// If the ack was an error, we'll need to bank send to the fallback address
	return k.SendToFallbackAddress(ctx, packet.Data, fallbackAddress)
}

// If there's a timed out packet, we'll infinitely retry the transfer
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	// Retrieve the fallback address from the original packet
	// We use the packet source channel here since this will correspond with the channel on Stride
	channelId := packet.SourceChannel
	originalSequence := packet.Sequence
	fallbackAddress, fallbackAddressFound := k.GetTransferFallbackAddress(ctx, channelId, originalSequence)

	// If there was no fallback address, this packet was not from an autopilot action and there's no need to retry
	if !fallbackAddressFound {
		return nil
	}

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

	// Update the fallback address to use the new sequence number
	updatedSequence := retryResponse.Sequence
	k.RemoveTransferFallbackAddress(ctx, channelId, originalSequence)
	k.SetTransferFallbackAddress(ctx, channelId, updatedSequence, fallbackAddress)

	return nil
}
