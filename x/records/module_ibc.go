package records

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"

	"github.com/Stride-Labs/stride/utils"
	"github.com/Stride-Labs/stride/x/records/keeper"
	"github.com/Stride-Labs/stride/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"

	// "google.golang.org/protobuf/proto" <-- this breaks tx parsing

	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
)

// IBC MODULE IMPLEMENTATION
// IBCModule implements the ICS26 interface for transfer given the transfer keeper.
type IBCModule struct {
	keeper          keeper.Keeper
	app             porttypes.IBCModule
	moduleAddresses map[string]string
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper, app porttypes.IBCModule, moduleAddresses map[string]string) IBCModule {
	return IBCModule{
		keeper:          k,
		app:             app,
		moduleAddresses: moduleAddresses,
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
	// doCustomLogic(packet, ack)
	// ICS-20 ack
	var ack channeltypes.Acknowledgement
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		im.keeper.Logger(ctx).Error("Error unmarshalling ack  %v", err)
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}
	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		im.keeper.Logger(ctx).Error("Error unmarshalling packet  %v", err)
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	// Custom ack logic only applies to ibc transfers initiated from the `stakeibc` module account
	// NOTE: if the `stakeibc` module account IBC transfers tokens for some other reason in the future,
	// this will need to be updated
	if data.Sender == im.keeper.AccountKeeper.GetModuleAddress(stakeibctypes.ModuleName).String() {
		switch resp := ack.Response.(type) {
		case *channeltypes.Acknowledgement_Result:
			log := fmt.Sprintf("\t [IBC-TRANSFER] Acknowledgement_Result {%s}", string(resp.Result))
			im.keeper.Logger(ctx).Info(log)
			// UPDATE RECORD
			// match record based on amount
			amount, err := strconv.ParseInt(data.Amount, 10, 64)
			if err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Error parsing int %d", amount)
			}
			record, found := im.keeper.GetTransferDepositRecordByAmount(ctx, amount)
			if !found {
				return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "No deposit record found for amount: %d", amount)
			}
			// update the record
			record.Status = types.DepositRecord_STAKE
			im.keeper.SetDepositRecord(ctx, *record)
			log = fmt.Sprintf("\t [IBC-TRANSFER] Deposit record updated: {%v}", record)
			im.keeper.Logger(ctx).Info(log)
		case *channeltypes.Acknowledgement_Error:
			log := fmt.Sprintf("\t [IBC-TRANSFER] Acknowledgement_Error {%s}", resp.Error)
			im.keeper.Logger(ctx).Error(log)
		}
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
	return im.ICS_20_module_patch_OnTimeoutPacket(ctx, packet, relayer)
}

// Wrapper for the refundPacketToken patch
// (Identical to the transfer module's `OnTimeoutPacket` except for the `refundPacketToken` call which calls the patched function)
func (im IBCModule) ICS_20_module_patch_OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	var data ibctransfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}
	// refund tokens
	if err := im.ICS_20_module_patch_refundPacketToken(ctx, packet, data); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTimeout,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(ibctransfertypes.AttributeKeyRefundReceiver, data.Sender),
			sdk.NewAttribute(ibctransfertypes.AttributeKeyRefundDenom, data.Denom),
			sdk.NewAttribute(ibctransfertypes.AttributeKeyRefundAmount, data.Amount),
		),
	)

	return nil
}

// The refundPacketToken function in the IBC transfer module will panic if it attempts to refund to a module address
// The module's address is blacklisted in the bankKeeper's `blockedAddr`, which is referenced in `SendCoinsFromModuleToAccount``
// This must be patched, otherwise, if a transfer from a module account fails, it will not be refunded
func (im IBCModule) ICS_20_module_patch_refundPacketToken(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data ibctransfertypes.FungibleTokenPacketData,
) error {
	// NOTE: packet data type already checked in handler.go

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
	if err := im.keeper.BankKeeper.MintCoins(ctx, ibctransfertypes.ModuleName, sdk.NewCoins(token)); err != nil {
		return err
	}

	// If the recipient is a module address, use SendCoinsFromModuleToModule
	for _, moduleName := range utils.StringToStringMapKeys(im.moduleAddresses) {
		moduleAddress := im.moduleAddresses[moduleName]
		if sender.String() == moduleAddress {
			if err := im.keeper.BankKeeper.SendCoinsFromModuleToModule(ctx, ibctransfertypes.ModuleName, moduleName, sdk.NewCoins(token)); err != nil {
				errMsg := fmt.Sprintf("unable to send coins from module (%s) to module (%s) ", types.ModuleName, moduleName)
				errMsg += fmt.Sprintf("despite previously minting coins to module account: %v", err)
				panic(errMsg)
			}
			return nil
		}
	}

	// Otherwise, if the recipient is not a module, use SendCoinsFromModuleToAccount
	if err := im.keeper.BankKeeper.SendCoinsFromModuleToAccount(ctx, ibctransfertypes.ModuleName, sender, sdk.NewCoins(token)); err != nil {
		errMsg := fmt.Sprintf("unable to send coins from module (%s) to account (%s) ", types.ModuleName, sender.String())
		errMsg += fmt.Sprintf("despite previously minting coins to module account: %v", err)
		panic(errMsg)
	}

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
