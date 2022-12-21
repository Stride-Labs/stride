package ratelimit

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	ratelimitkeeper "github.com/Stride-Labs/stride/v5/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

// Parse the denom from the Send Packet that will be used by the rate limit module
// The denom that the rate limiter will use for a SEND packet depends on whether
//
//	it was a NATIVE token (e.g. ustrd, stuatom, etc.) or NON-NATIVE token (e.g. ibc/...)...
//
// We can identify if the token is native or not by parsing the trace denom from the packet
// If the token is NATIVE, it will not have a prefix (e.g. ustrd),
//
//	and if it is NON-NATIVE, it will have a prefix (e.g. transfer/channel-2/uosmo)
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
//
//			Source: The packet's is being received by a chain it was just sent from (i.e. the token has gone back and forth)
//	             (e.g. strd is sent -> to osmosis -> and then back to stride)
//	     Sink:   The packet's is being received by a chain that either created it or previous received it from somewhere else
//	             (e.g. atom is sent -> to stride) (e.g.2. atom is sent -> to osmosis -> which is then sent to stride)
//
//		     If the chain is acting as a SINK:
//		     	We add on the Stride port and channel and hash it
//		         Ex1: uosmo sent from Osmosis to Stride
//		             Packet Denom:   uosmo
//		              -> Add Prefix: transfer/channel-X/uosmo
//		              -> Hash:       ibc/...
//
//		         Ex2: ujuno sent from Osmosis to Stride
//		             PacketDenom:    transfer/channel-Y/ujuno  (channel-Y is the Juno <> Osmosis channel)
//		              -> Add Prefix: transfer/channel-X/transfer/channel-Y/ujuno
//		              -> Hash:       ibc/...
//
//		     If the chain is acting as a SOURCE:
//		     	First, remove the prefix. Then if there is still a denom trace, hash it
//		         Ex1: ustrd sent back to Stride from Osmosis
//		             Packet Denom:      transfer/channel-X/ustrd
//		              -> Remove Prefix: ustrd
//		              -> Leave as is:   ustrd
//
//					Ex2: juno was sent to Stride, then to Osmosis, then back to Stride
//		             Packet Denom:      transfer/channel-X/transfer/channel-Z/ujuno
//		              -> Remove Prefix: transfer/channel-Z/ujuno
//		              -> Hash:          ibc/...
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
func SendRateLimitedPacket(ctx sdk.Context, keeper ratelimitkeeper.Keeper, packet exported.PacketI) error {
	// For a send packet, the channel on stride is the "Source" channel
	//  This is because the Source and Desination are defined from the perspective of a packet recipient
	//    i.e., when this packet lands on a the host chain, the "Source" will show the Stride Channel
	channelId := packet.GetSourceChannel()

	// Parse the packet data
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}

	amount, ok := sdk.NewIntFromString(packetData.Amount)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Unable to cast packet amount to sdk.Int")
	}

	denom := ParseDenomFromSendPacket(packetData)

	err := keeper.CheckRateLimitAndUpdateFlow(ctx, types.PACKET_SEND, denom, channelId, amount)
	if err != nil {
		return err
	}
	return nil
}

// Middleware implementation for RecvPacket with rate limiting
func ReceiveRateLimitedPacket(ctx sdk.Context, keeper ratelimitkeeper.Keeper, packet channeltypes.Packet) error {
	// For a receive packet, the channel on stride is the "Destination" channel
	// This is because the Source and Desination is defined from the perspective of a packet recipient
	// Meaning, when this packet lands on a Stride, the "Destination" will show the Stride Channel
	channelId := packet.GetDestChannel()

	// Parse the amount and denom from the packet
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}

	amount, ok := sdk.NewIntFromString(packetData.Amount)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Unable to cast packet amount to sdk.Int")
	}

	denom := ParseDenomFromRecvPacket(packet, packetData)

	// Check whether the rate limit has been exceeded - and if it hasn't, send the packet
	err := keeper.CheckRateLimitAndUpdateFlow(ctx, types.PACKET_RECV, denom, channelId, amount)
	if err != nil {
		return err
	}
	return nil
}
