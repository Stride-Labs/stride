package records

import (
	"github.com/Stride-Labs/stride/x/records/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"

	// host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

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
	im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		appVersion, // note we only pass app version here
	)
	return nil
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
	ctx.Logger().Error("version %s: ", version)
	ctx.Logger().Error("cpAppVersion %s: ", cpAppVersion)
	_ = version
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
	im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelId, counterpartyVersion)
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// doCustomLogic()
	im.app.OnChanOpenConfirm(ctx, portID, channelID)
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// doCustomLogic()
	im.app.OnChanCloseInit(ctx, portID, channelID)
	return nil
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// doCustomLogic()
	im.app.OnChanCloseConfirm(ctx, portID, channelID)
	return nil
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
	_ = wrapperAck
	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	// doCustomLogic(packet)
	transferAck := im.app.OnRecvPacket(ctx, packet, relayer)
	_ = transferAck

	// doCustomLogic(transferAck) // middleware may modify outgoing ack
	return wrapperAck
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// doCustomLogic(packet, ack)
	// Store a deposit record here!
	// // create a deposit record of these tokens
	// fmt.Println("packet returned from MsgTransfer: ", 	packet.)

	// // packet returned from MsgTransfer: {1 transfer channel-0 transfer channel-0
	// // [123 34 97 109 111 117 110 116 34 58 34 49 48 48 48 34 44 34 100 101 110 111 109 34 58 34 116 114 97 110 115 102 101 114 47 99 104 97 110 110 101 108 45 48 47 117 97 116 111 109 34 44 34 114 101 99 101 105 118 101 114 34 58 34 99 111 115 109 111 115 49 57 108 54 100 51 100 55 107 50 112 101 108 56 101 112 103 99 112 120 99 57 110 112 54 102 115 118 106 112 97 97 97 48 54 110 109 54 53 118 97 103 119 120 97 112 48 101 52 106 101 122 113 48 53 109 109 118 117 34 44 34 115 101 110 100 101 114 34 58 34 115 116 114 105 100 101 49 109 118 100 113 52 110 108 117 112 108 51 57 50 52 51 113 106 122 55 115 100 115 53 101 122 51 114 108 57 109 110 120 50 53 51 108 122 97 34 125] 0-1000000 0} 4:09PM ERR This is where we're going to add DepositRecords!

	// // var packetData icatypes.InterchainAccountPacketData
	// var packetData ibctransfertypes.MsgTransferReponse
	// // err := icatypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &packetData)
	// err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &packetData)
	// if err != nil {
	// 	fmt.Println("unable to unmarshal acknowledgement packet data", "error", err, "data", packetData)
	// 	return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	// }

	// // msgs, err := icatypes.DeserializeCosmosTx(, packetData.Data)
	// // if err != nil {
	// // 	fmt.Println("unable to decode messages", "err", err)
	// // 	return err
	// // }
	// // fmt.Println("Decoded msgs: ", msgs)

	// // depositRecord := recordtypes.NewDepositRecord(msg.Amount, msg.HostDenom, hostZone.ChainId,
	// // 	sender.String(), recordtypes.DepositRecord_RECEIPT)
	// // k.RecordsKeeper.AppendDepositRecord(ctx, *depositRecord)

	// depositRecord := recordtypes.NewDepositRecord(1, "uatom", "GAIA", "cosmosXXXXXXXX", recordtypes.DepositRecord_RECEIPT)
	// im.keeper.AppendDepositRecord(ctx, *depositRecord)

	// ctx.Logger().Error("This is where we're going to add DepositRecords!")
	// im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// doCustomLogic(packet)
	im.app.OnTimeoutPacket(ctx, packet, relayer)
	return nil
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
	return "ics20-1", true // im.keeper.GetAppVersion(ctx, portID, channelID)
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
