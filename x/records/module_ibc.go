package records

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v4/x/records/keeper"

	// "google.golang.org/protobuf/proto" <-- this breaks tx parsing

	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
)

// IBC MODULE IMPLEMENTATION
// IBCModule implements the ICS26 interface for transfer given the transfer keeper.
type IBCModule struct {
	keeper keeper.Keeper
	app    porttypes.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper, app porttypes.IBCModule) IBCModule {
	return IBCModule{
		keeper: k,
		app:    app,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	// Note: The channel capability must be claimed by the authentication module in OnChanOpenInit otherwise the
	// authentication module will not be able to send packets on the channel created for the associated interchain account.
	// NOTE: unsure if we have to claim this here! CHECK ME
	// if err := im.keeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
	// 	return err
	// }
	_, appVersion := channeltypes.SplitChannelVersion(version)
	// doCustomLogic()
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		appVersion, // note we only pass app version here
	)
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	// doCustomLogic()
	// core/04-channel/types contains a helper function to split middleware and underlying app version
	_, cpAppVersion := channeltypes.SplitChannelVersion(counterpartyVersion)

	// call the underlying applications OnChanOpenTry callback
	version, err := im.app.OnChanOpenTry(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		chanCap,
		counterparty,
		cpAppVersion, // note we only pass counterparty app version here
	)
	if err != nil {
		return "", err
	}
	ctx.Logger().Info(fmt.Sprintf("IBC Chan Open Version %s: ", version))
	ctx.Logger().Info(fmt.Sprintf("IBC Chan Open cpAppVersion %s: ", cpAppVersion))
	return version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelId string, // counterpartyChannelId
	counterpartyVersion string,
) error {
	// core/04-channel/types contains a helper function to split middleware and underlying app version
	// _, _ := channeltypes.SplitChannelVersion(counterpartyVersion)
	// doCustomLogic()
	// call the underlying applications OnChanOpenTry callback
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelId, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// doCustomLogic()
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// doCustomLogic()
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// doCustomLogic()
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is successfully decoded and the receive application
// logic returns without error.
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	wrapperAck := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	// handle(wrapperAck)
	_ = wrapperAck
	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	// doCustomLogic(packet)
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("[IBC-TRANSFER] OnAcknowledgementPacket  %v", packet))
	// doCustomLogic(packet, ack)
	// ICS-20 ack
	var ack channeltypes.Acknowledgement
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("Error unmarshalling ack  %v", err.Error()))
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	// log the ack type
	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		im.keeper.Logger(ctx).Info(fmt.Sprintf("\t [IBC-TRANSFER] Acknowledgement_Result {%s}", string(resp.Result)))
	case *channeltypes.Acknowledgement_Error:
		im.keeper.Logger(ctx).Error(fmt.Sprintf("\t [IBC-TRANSFER] Acknowledgement_Error {%s}", resp.Error))
	default:
		im.keeper.Logger(ctx).Error(fmt.Sprintf("\t [IBC-TRANSFER] Unrecognized ack for packet {%v}", packet))
	}

	// Custom ack logic only applies to ibc transfers initiated from the `stakeibc` module account
	// NOTE: if the `stakeibc` module account IBC transfers tokens for some other reason in the future,
	// this will need to be updated
	err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, packet, &ack)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to call registered callback from records OnAcknowledgePacket | Sequence %d, from %s %s, to %s %s | Error %s",
			packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort, err.Error())
		im.keeper.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(icacallbacktypes.ErrCallbackFailed, errMsg)
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// doCustomLogic(packet)
	im.keeper.Logger(ctx).Error(fmt.Sprintf("[IBC-TRANSFER] OnTimeoutPacket  %v", packet))
	err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, packet, nil)
	if err != nil {
		return err
	}
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// This is implemented by ICS4 and all middleware that are wrapping base application.
// The base application will call `sendPacket` or `writeAcknowledgement` of the middleware directly above them
// which will call the next middleware until it reaches the core IBC handler.
// SendPacket implements the ICS4 Wrapper interface
func (im IBCModule) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
) error {
	return nil
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCModule) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	return nil
}

// GetAppVersion returns the interchain accounts metadata.
func (im IBCModule) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return ibctransfertypes.Version, true // im.keeper.GetAppVersion(ctx, portID, channelID)
}

// APP MODULE IMPLEMENTATION
// OnChanOpenInit implements the IBCModule interface
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// OnChanOpenTry implements the IBCModule interface
func (am AppModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version,
	counterpartyVersion string,
) error {
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// OnChanOpenAck implements the IBCModule interface
func (am AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyVersion string,
) error {
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// OnChanOpenConfirm implements the IBCModule interface
func (am AppModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// OnChanCloseInit implements the IBCModule interface
func (am AppModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for channels
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (am AppModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// OnRecvPacket implements the IBCModule interface
func (am AppModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return nil
}

// OnAcknowledgementPacket implements the IBCModule interface
func (am AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// OnTimeoutPacket implements the IBCModule interface
func (am AppModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return nil
}

func (am AppModule) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (version string, err error) {
	return proposedVersion, nil
}
