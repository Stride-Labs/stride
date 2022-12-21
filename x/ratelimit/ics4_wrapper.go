package ratelimit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	ratelimitkeeper "github.com/Stride-Labs/stride/v5/x/ratelimit/keeper"
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
	if err := SendRateLimitedPacket(ctx, i.rateLimitKeeper, packet); err != nil {
		return err
	}
	return i.channel.SendPacket(ctx, chanCap, packet)
}

func (i *ICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return i.channel.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
