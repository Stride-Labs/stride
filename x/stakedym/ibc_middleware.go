package stakedym

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/Stride-Labs/stride/v29/x/stakedym/keeper"
)

var _ porttypes.Middleware = &IBCMiddleware{}

type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper
func NewIBCMiddleware(k keeper.Keeper, app porttypes.IBCModule) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// No custom logic needed for OnChanOpenInit - passes through to next middleware
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck (Stakedym): portID %s, channelID %s", portID, channelID))
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		version,
	)
}

// No custom logic needed for OnChanOpenAck - passes through to next middleware
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck (Stakedym): portID %s, channelID %s, counterpartyChannelID %s, counterpartyVersion %s",
		portID, channelID, counterpartyChannelID, counterpartyVersion))
	return im.app.OnChanOpenAck(
		ctx,
		portID,
		channelID,
		counterpartyChannelID,
		counterpartyVersion,
	)
}

// No custom logic needed for OnChanCloseConfirm - passes through to next middleware
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanCloseConfirm (Stakedym): portID %s, channelID %s", portID, channelID))
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnAcknowledgementPacket must check the ack for outbound transfers of native tokens
// and update record keeping based on whether it succeeded
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnAcknowledgementPacket (Stakedym): SourcePort %s, SourceChannel %s, DestinationPort %s, DestinationChannel %s",
		packet.SourcePort, packet.SourceChannel, packet.DestinationPort, packet.DestinationChannel))
	// Handle stakedym specific logic
	if err := im.keeper.OnAcknowledgementPacket(ctx, packet, acknowledgement); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 stakedym OnAckPacket failed: %s", err.Error()))
		return err
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket must check if an outbound transfer of native tokens timed out,
// and, if so, adjust record keeping
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnTimeoutPacket (Stakedym): packet %v, relayer %v", packet, relayer))
	// Handle stakedym specific logic
	if err := im.keeper.OnTimeoutPacket(ctx, packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 stakedym OnTimeoutPacket failed: %s", err.Error()))
		return err
	}

	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// No custom logic needed for OnChanOpenTry - passes through to next middleware
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		counterpartyVersion,
	)
}

// No custom logic needed for OnChanOpenConfirm - passes through to next middleware
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// No custom logic needed for OnChanCloseInit - passes through to next middleware
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// No custom logic needed for OnRecvPacket - passes through to next middleware
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// Send implements the ICS4Wrapper interface
// Stakedym sits above where ICS4 traffic routes in the transfer stack
// so this should never get called
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	panic("Unexpected ICS4Wrapper route to stakedym module")
}

// WriteAcknowledgement implements the ICS4Wrapper interface
// Stakedym sits above where ICS4 traffic routes in the transfer stack
// so this should never get called
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	panic("Unexpected ICS4Wrapper route to stakedym module")
}

// GetAppVersion implements the ICS4Wrapper interface
// Stakedym sits above where ICS4 traffic routes in the transfer stack
// so this should never get called
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	panic("Unexpected ICS4Wrapper route to stakedym module")
}
