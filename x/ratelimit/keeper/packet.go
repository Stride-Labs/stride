package keeper

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

// Parse the denom from the Send Packet that will be used by the rate limit module
// The denom that the rate limiter will use for a SEND packet depends on whether
//    it was a NATIVE token (e.g. ustrd, stuatom, etc.) or NON-NATIVE token (e.g. ibc/...)...
//
// We can identify if the token is native or not by parsing the trace denom from the packet
// If the token is NATIVE, it will not have a prefix (e.g. ustrd),
//    and if it is NON-NATIVE, it will have a prefix (e.g. transfer/channel-2/uosmo)
//
// For NATIVE denoms, return as is (e.g. ustrd)
// For NON-NATIVE denoms, take the ibc hash (e.g. hash "transfer/channel-2/usoms" into "ibc/...")
func ParseDenomFromSendPacket(packet transfertypes.FungibleTokenPacketData) (denom string) {
	// Determine the denom by looking at the denom trace path
	denomTrace := transfertypes.ParseDenomTrace(packet.Denom)

	// Native assets will have an empty trace path and can be returned as is
	if denomTrace.Path == "" {
		denom = packet.Denom
	} else {
		// Non-native assets should be hashed
		denom = denomTrace.IBCDenom()
	}

	return denom
}

// Parse the denom from the Recv Packet that will be used by the rate limit module
// The denom that the rate limiter will use for a RECEIVE packet depends on whether it was a source or sink
//      Sink:   The token moves forward, to a chain different than its previous hop
//              The new port and channel are APPENDED to the denom trace.
//              (e.g. A -> B, B is a sink) (e.g. A -> B -> C, C is a sink)
// 		Source: The token moves backwards (i.e. revisits the last chain it was sent from)
// 				The port and channel are REMOVED from the denom trace - undoing the last hop.
//              (e.g. A -> B -> A, A is a source) (e.g. A -> B -> C -> B, B is a source)
//
//      If the chain is acting as a SINK:
//      	We add on the Stride port and channel and hash it
//          Ex1: uosmo sent from Osmosis to Stride
//              Packet Denom:   uosmo
//               -> Add Prefix: transfer/channel-X/uosmo
//               -> Hash:       ibc/...
//
//          Ex2: ujuno sent from Osmosis to Stride
//              PacketDenom:    transfer/channel-Y/ujuno  (channel-Y is the Juno <> Osmosis channel)
//               -> Add Prefix: transfer/channel-X/transfer/channel-Y/ujuno
//               -> Hash:       ibc/...
//
//      If the chain is acting as a SOURCE:
//      	First, remove the prefix. Then if there is still a denom trace, hash it
//          Ex1: ustrd sent back to Stride from Osmosis
//              Packet Denom:      transfer/channel-X/ustrd
//               -> Remove Prefix: ustrd
//               -> Leave as is:   ustrd
//
//			Ex2: juno was sent to Stride, then to Osmosis, then back to Stride
//              Packet Denom:      transfer/channel-X/transfer/channel-Z/ujuno
//               -> Remove Prefix: transfer/channel-Z/ujuno
//               -> Hash:          ibc/...
func ParseDenomFromRecvPacket(packet channeltypes.Packet, packetData transfertypes.FungibleTokenPacketData) (denom string) {
	// To determine the denom, first check whether Stride is acting as source
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), packetData.Denom) {
		// Remove the source prefix (e.g. transfer/channel-X/transfer/channel-Z/ujuno -> transfer/channel-Z/ujuno)
		sourcePrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := packetData.Denom[len(sourcePrefix):]

		// Native assets will have an empty trace path and can be returned as is
		denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
		if denomTrace.Path == "" {
			denom = unprefixedDenom
		} else {
			// Non-native assets should be hashed
			denom = denomTrace.IBCDenom()
		}
	} else {
		// Prefix the destination channel - this will contain the trailing slash (e.g. transfer/channel-X/)
		destinationPrefix := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		prefixedDenom := destinationPrefix + packetData.Denom

		// Hash the denom trace
		denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
		denom = denomTrace.IBCDenom()
	}

	return denom
}

// Middleware implementation for SendPacket with rate limiting
func (k Keeper) SendRateLimitedPacket(ctx sdk.Context, packet ibcexported.PacketI) error {
	// The Stride channelID should always be used as the key for the RateLimit object (not the counterparty channelID)
	// For a SEND packet, the Stride channelID is the SOURCE channel
	// This is because the Source and Desination are defined from the perspective of a packet recipient
	// Meaning, when this packet lands on a the host chain, the "Source" will be the Stride Channel,
	//   and the "Destination" will be the Host Channel
	channelId := packet.GetSourceChannel()

	// Parse the packet data
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}

	amount, ok := sdk.NewIntFromString(packetData.Amount)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Unable to cast packet amount to sdkmath.Int")
	}

	denom := ParseDenomFromSendPacket(packetData)

	err := k.CheckRateLimitAndUpdateFlow(ctx, types.PACKET_SEND, denom, channelId, amount)
	if err != nil {
		return err
	}

	return nil
}

// Middleware implementation for RecvPacket with rate limiting
func (k Keeper) ReceiveRateLimitedPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	// The Stride channelID should always be used as the key for the RateLimit object (not the counterparty channelID)
	// For a RECEIVE packet, the Stride channelID is the DESTINATION channel
	// This is because the Source and Desination are defined from the perspective of a packet recipient
	// Meaning, when this packet lands on a Stride, the "Source" will be the host zone's channel,
	//  and the "Destination" will be the Stride Channel
	channelId := packet.GetDestChannel()

	// Parse the amount and denom from the packet
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}

	amount, ok := sdk.NewIntFromString(packetData.Amount)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Unable to cast packet amount to sdkmath.Int")
	}

	denom := ParseDenomFromRecvPacket(packet, packetData)

	// Check whether the rate limit has been exceeded - and if it hasn't, send the packet
	err := k.CheckRateLimitAndUpdateFlow(ctx, types.PACKET_RECV, denom, channelId, amount)
	if err != nil {
		return err
	}

	return nil
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
// If the packet does not get rate limited, it passes the packet to the IBC Channel keeper
func (k Keeper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	if err := k.SendRateLimitedPacket(ctx, packet); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("ICS20 packet send was denied: %s", err.Error()))
		return err
	}
	return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement wraps IBC ChannelKeeper's WriteAcknowledgement function
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// GetAppVersion wraps IBC ChannelKeeper's GetAppVersion function
func (k Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
