package icaoracle

import (
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v11/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v11/modules/core/exported"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v32/x/icaoracle/keeper"
)

var _ porttypes.Middleware = &IBCMiddleware{}

type IBCMiddleware struct {
	app         porttypes.IBCModule
	keeper      keeper.Keeper
	ics4Wrapper porttypes.ICS4Wrapper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper
func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper) *IBCMiddleware {
	return &IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// SetICS4Wrapper sets the ICS4Wrapper above this module on the IBC stack.
func (im *IBCMiddleware) SetICS4Wrapper(wrapper porttypes.ICS4Wrapper) {
	im.ics4Wrapper = wrapper
}

// SetUnderlyingApplication sets the underlying IBC application beneath this middleware.
func (im *IBCMiddleware) SetUnderlyingApplication(app porttypes.IBCModule) {
	im.app = app
}

// OnChanOpenInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		counterparty,
		version,
	)
}

// OnChanOpenTry simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck (ICAOracle): portID %s, channelID %s, counterpartyChannelID %s, counterpartyVersion %s",
		portID, channelID, counterpartyChannelID, counterpartyVersion))

	if err := im.keeper.OnChanOpenAck(ctx, portID, channelID); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICAOracle ChanOpenAck failed: %s", err.Error()))
		return errorsmod.Wrapf(err, "ICAOracle OnChanOpenAck failed")
	}

	return im.app.OnChanOpenAck(
		ctx,
		portID,
		channelID,
		counterpartyChannelID,
		counterpartyVersion,
	)
}

// OnChanCloseConfirm simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnChanCloseInit simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanOpenConfirm simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnAcknowledgementPacket simply passes down the to next middleware stack
// The Ack handling and routing is managed by icacallbacks
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
}

// OnTimeoutPacket simply passes down the to next middleware stack
// The Ack handling and routing is managed by icacallbacks
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.app.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
}

// OnRecvPacket simply passes down the to next middleware stack
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return im.app.OnRecvPacket(ctx, channelVersion, packet, relayer)
}

// SendPacket implements the ICS4 Wrapper interface but is not utilized in the ICA stack
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	panic("UNIMPLEMENTED")
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
// but is not utilized in the bottom of ICA stack
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	panic("UNIMPLEMENTED")
}

// GetAppVersion implements the ICS4 Wrapper interface
// but is not utilized in the bottom of ICA stack
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	panic("UNIMPLEMENTED")
}
