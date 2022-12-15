package ratelimit

import (
	"encoding/json"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	ratelimitkeeper "github.com/Stride-Labs/stride/v4/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

var (
	_ porttypes.Middleware  = &IBCModule{}
	_ porttypes.ICS4Wrapper = &ICS4Wrapper{}
)

type ICS4Wrapper struct {
	channel         porttypes.ICS4Wrapper
	rateLimitKeeper ratelimitkeeper.Keeper
}

func NewICS4Middleware(channel porttypes.ICS4Wrapper, ratelimitKeeper ratelimitkeeper.Keeper) ICS4Wrapper {
	return ICS4Wrapper{
		channel:         channel,
		rateLimitKeeper: ratelimitKeeper,
	}
}

func (i *ICS4Wrapper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	// For a send packet, the channel on stride is the "Source" channel
	//  This is because the Source and Desination are defined from the perspective of a packet recipient
	//    i.e., when this packet lands on a the host chain, the "Source" will show the Stride Channel
	channelId := packet.GetSourceChannel()

	// Parse the packet data
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}

	// TODO: Switch to type sdk.Int
	amount, err := strconv.ParseUint(packetData.Amount, 10, 64)
	if err != nil {
		return err
	}

	denom := ratelimitkeeper.ParseDenomFromSendPacket(packetData)

	err = i.rateLimitKeeper.CheckRateLimit(ctx, types.PACKET_SEND, denom, channelId, amount)
	if err != nil {
		return err
	}
	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i *ICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
