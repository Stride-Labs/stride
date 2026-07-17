package app

import (
	ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v11"
	porttypes "github.com/cosmos/ibc-go/v11/modules/core/05-port/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// The core ibc-go rate-limiting middleware requires its underlying application to
// implement the porttypes.PacketUnmarshalerModule interface, but ibc-hooks (which
// sits directly below it in the transfer stack) does not implement UnmarshalPacketData.
//
// Since ibc-hooks never modifies packet data, this adapter wraps the ibc-hooks
// middleware and delegates unmarshaling to the middleware below it (PFM), which
// in turn delegates down to the transfer application.
type ibcHooksWithPacketUnmarshaler struct {
	*ibchooks.IBCMiddleware
	packetUnmarshaler porttypes.PacketDataUnmarshaler
}

var _ porttypes.PacketUnmarshalerModule = ibcHooksWithPacketUnmarshaler{}

func (m ibcHooksWithPacketUnmarshaler) UnmarshalPacketData(ctx sdk.Context, portID, channelID string, bz []byte) (any, string, error) {
	return m.packetUnmarshaler.UnmarshalPacketData(ctx, portID, channelID, bz)
}
