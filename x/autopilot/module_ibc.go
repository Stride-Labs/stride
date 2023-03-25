package autopilot

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"

	"github.com/Stride-Labs/stride/v7/x/autopilot/keeper"
	"github.com/Stride-Labs/stride/v7/x/autopilot/types"

	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"
)

// IBC MODULE IMPLEMENTATION
// IBCModule implements the ICS26 interface for transfer given the transfer keeper.
// TODO: Use IBCMiddleware struct
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
) (string, error) {
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
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelId string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelId, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
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
	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// ibc-go v5 has a Memo field that can store forwarding info
	// For older version of ibc-go, the data must be stored in the receiver field
	var metadata string
	if data.Memo != "" {
		metadata = data.Memo
	} else {
		metadata = data.Receiver
	}

	// parse out any forwarding info
	parsedPacketMetadata, err := types.ParsePacketMetadata(metadata)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// If the parsed metadata is nil, that means there is no forwarding logic
	// Pass the packet down to the next middleware
	if parsedPacketMetadata == nil {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Handle for each module
	if parsedPacketMetadata.Stakeibc.ShouldForward {
		// Modify packet data to process packet transfer for this chain, omitting liquid staking info
		newData := data
		newData.Receiver = parsedPacketMetadata.Stakeibc.StrideAddress.String()
		bz, err := transfertypes.ModuleCdc.MarshalJSON(&newData)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		newPacket := packet
		newPacket.Data = bz

		// process the transfer receipt
		// NOTE: this code is pulled from packet-forwarding-middleware
		ack := im.app.OnRecvPacket(ctx, newPacket, relayer)
		if ack.Success() {
			return im.keeper.TryLiquidStaking(ctx, packet, newData, &parsedPacketMetadata.Stakeibc, ack)
		}
		return ack
	} else if parsedPacketMetadata.Claim.ShouldForward {
		// TODO
		return im.app.OnRecvPacket(ctx, packet, relayer)
	} else {
		// This branch should not be possible - it would mean that the packet was successfully
		// parsed into one of the autopilot modules, but the module is not displayed in this if/else statement
		return channeltypes.NewErrorAcknowledgement(types.ErrUnsupportedAutopilotRoute)
	}
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("[IBC-TRANSFER] OnAcknowledgementPacket  %v", packet))
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Error(fmt.Sprintf("[IBC-TRANSFER] OnTimeoutPacket  %v", packet))
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
