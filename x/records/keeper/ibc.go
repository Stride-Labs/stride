package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v14/x/icacallbacks"
	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
)

// OnAcknowledgementPacket unmarshals the acknowledgement object to determine if the ack was successful and
// then passes that to the ICACallback
// If this packet does not have associated callback data, there will be no additional ack logic in CallRegisteredICACallback
func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	packetDescription := fmt.Sprintf("Sequence %d, from %s %s, to %s %s",
		packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)

	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, k.Logger(ctx), acknowledgement, false)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to unpack message data from acknowledgement - %s", packetDescription)
	}

	// Custom ack logic only applies to ibc transfers initiated from the `stakeibc` module account
	// NOTE: if the `stakeibc` module account IBC transfers tokens for some other reason in the future,
	// this will need to be updated
	if err := k.ICACallbacksKeeper.CallRegisteredICACallback(ctx, packet, ackResponse); err != nil {
		return errorsmod.Wrapf(err, "unable to call registered callback for records OnAckPacket - %s", packetDescription)
	}

	return nil
}

// OnTimeoutPacket passes the ack timeout to the ICACallback
// If there was no callback data associated with this packet,
// there will be no additional ack logic in CallRegisteredICACallback
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_TIMEOUT}
	if err := k.ICACallbacksKeeper.CallRegisteredICACallback(ctx, packet, &ackResponse); err != nil {
		return errorsmod.Wrapf(err, "unable to call registered callback for records OnTimeoutPacket")
	}

	return nil
}
