package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
)

// ClientKeeper defines the expected IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
}

// ConnectionKeeper defines the expected IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	GetChannelConnection(ctx sdk.Context, portID, channelID string) (string, exported.ConnectionI, error)
}

// ICAControllerKeeper defines the expected ICA controller keeper
type ICAControllerKeeper interface {
	GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
	GetOpenActiveChannel(ctx sdk.Context, connectionID, portID string) (string, bool)
	RegisterInterchainAccount(ctx sdk.Context, connectionID, owner, version string) error
}

// ICACallbacksKeeper defines the expected ICA callback keeper
type ICACallbacksKeeper interface {
	SetCallbackData(ctx sdk.Context, callbackData icacallbackstypes.CallbackData)
}

// ICS4Wrapper defines the expected ICS4Wrapper for middleware
type ICS4Wrapper interface {
	WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error
	SendPacket(
		ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (sequence uint64, err error)
	GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool)
}
