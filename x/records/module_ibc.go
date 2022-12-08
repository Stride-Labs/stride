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
	"github.com/Stride-Labs/stride/v4/x/records/types"

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
	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %v", err)
	}

	// log the ack type
	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		im.keeper.Logger(ctx).Info(fmt.Sprintf("\t [IBC-TRANSFER] Acknowledgement_Result {%s}", string(resp.Result)))
	case *channeltypes.Acknowledgement_Error:
		im.keeper.Logger(ctx).Error(fmt.Sprintf("\t [IBC-TRANSFER] Acknowledgement_Error {%s}", resp.Error))
		return im.refundPacketToken(ctx, packet, data)
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
		return im.refundPacketToken(ctx, packet, data)
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

func (im IBCModule) refundPacketToken(ctx sdk.Context, packet channeltypes.Packet, data ibctransfertypes.FungibleTokenPacketData) error {
	callbackDataKey := icacallbacktypes.PacketID(packet.GetSourcePort(), packet.GetSourceChannel(), packet.Sequence)
	callbackData, _ := im.keeper.ICACallbacksKeeper.GetCallbackDataFromPacket(ctx, packet, callbackDataKey)
	transferCallback, _ := im.keeper.UnmarshalTransferCallbackArgs(ctx, callbackData.CallbackArgs)

	// parse the denomination from the full denom path
	trace := ibctransfertypes.ParseDenomTrace(data.Denom)

	// parse the transfer amount
	transferAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.Wrapf(ibctransfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into math.Int", data.Amount)
	}
	token := sdk.NewCoin(trace.IBCDenom(), transferAmount)

	// decode the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	if ibctransfertypes.SenderChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		// unescrow tokens back to sender
		escrowAddress := ibctransfertypes.GetEscrowAddress(packet.GetSourcePort(), packet.GetSourceChannel())
		if err := im.keeper.BankKeeper.SendCoins(ctx, escrowAddress, sender, sdk.NewCoins(token)); err != nil {
			// NOTE: this error is only expected to occur given an unexpected bug or a malicious
			// counterparty module. The bug may occur in bank or any part of the code that allows
			// the escrow address to be drained. A malicious counterparty module could drain the
			// escrow address by allowing more tokens to be sent back then were escrowed.
			return sdkerrors.Wrap(err, "unable to unescrow tokens, this may be caused by a malicious counterparty module or a bug: please open an issue on counterparty module")
		}

		return nil
	}

	// mint vouchers back to sender
	if err := im.keeper.BankKeeper.MintCoins(
		ctx, ibctransfertypes.ModuleName, sdk.NewCoins(token),
	); err != nil {
		return err
	}

	if err := im.keeper.BankKeeper.SendCoinsFromModuleToAccount(ctx, ibctransfertypes.ModuleName, sender, sdk.NewCoins(token)); err != nil {
		panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
	}

	// update deposit record
	depositRecord, found := im.keeper.GetDepositRecord(ctx, transferCallback.DepositRecordId)
	if !found {
		return sdkerrors.Wrapf(types.ErrUnknownDepositRecord, "deposit record not found %d", transferCallback.DepositRecordId)
	}
	depositRecord.Status = types.DepositRecord_TRANSFER_QUEUE
	im.keeper.SetDepositRecord(ctx, depositRecord)

	return nil
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
