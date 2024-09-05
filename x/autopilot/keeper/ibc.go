package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v24/x/icacallbacks"
	icacallbacktypes "github.com/Stride-Labs/stride/v24/x/icacallbacks/types"
)

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
	amount, ok := sdkmath.NewIntFromString(transferMetadata.Amount)
	if !ok {
		return fmt.Errorf("unable to parse amount from transfer packet: %v", transferMetadata)
	}
	token := sdk.NewCoin(transferMetadata.Denom, amount)

	// Finally send to the fallback account
	if err := k.bankKeeper.SendCoins(ctx, senderAccount, fallbackAccount, sdk.NewCoins(token)); err != nil {
		return err
	}

	return nil
}

// If there was a timeout or failed ack from an outbound transfer of one of the autopilot actions,
// we'll need to check if there was a fallback address. If one was stored, bank send to that address
// If the ack was successful, we should delete the address (if it exists)
func (k Keeper) HandleFallbackAddress(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, packetTimedOut bool) error {
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

	// If the packet timed out, send to the fallback address
	if packetTimedOut {
		return k.SendToFallbackAddress(ctx, packet.Data, fallbackAddress)
	}

	// If the packet did not timeout, check whether the ack was successful or was an ack error
	isICATx := false
	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, k.Logger(ctx), acknowledgement, isICATx)
	if err != nil {
		return err
	}

	// If successful, no additional action is necessary
	if ackResponse.Status == icacallbacktypes.AckResponseStatus_SUCCESS {
		return nil
	}

	// If there was an ack error, we'll need to bank send to the fallback address
	return k.SendToFallbackAddress(ctx, packet.Data, fallbackAddress)
}

// OnTimeoutPacket should always send to the fallback address
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	return k.HandleFallbackAddress(ctx, packet, []byte{}, true)
}

// OnAcknowledgementPacket should send to the fallback address if the ack is an ack error
func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	return k.HandleFallbackAddress(ctx, packet, acknowledgement, false)
}
