package icacallbacks

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/Stride-Labs/stride/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/x/icacallbacks/keeper"
)

// IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

type connectionIdContextKey string

func (c connectionIdContextKey) String() string {
	return string(c)
}

// func(ctx, order, connectionHops []string, portID string, channelID string, chanCap, counterparty, version string) (string, error)
// func(ctx , order , connectionHops []string, portID string, channelID string, channelCap , counterparty , version string) error)
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
	if err := im.keeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}
	return nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return nil
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.CallRegisteredICACallback(ctx, modulePacket, acknowledgement)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.CallRegisteredICACallback(ctx, modulePacket, []byte{})
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// TODO(TEST-21): Implement OnTimeoutPacket logic
	return nil
}

func (im IBCModule) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (version string, err error) {
	return proposedVersion, nil
}

func (im IBCModule) CallRegisteredICACallback(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte) error {
	// get the relevant module from the channel and port
	portID := modulePacket.GetSourcePort()
	channelID := modulePacket.GetSourceChannel()
	module, _, err := im.keeper.IBCKeeper.ChannelKeeper.LookupModuleByChannel(ctx, portID, channelID)
	if err != nil {
		return err
	}
	// fetch the callback data
	callbackDataKey := types.CallbackDataKeyFormatter(portID, channelID, modulePacket.Sequence)
	callbackData, found := im.keeper.GetCallbackData(ctx, callbackDataKey)
	if !found {
		errMsg := fmt.Sprintf("callback data not found for portID: %s, channelID: %s, sequence: %d", portID, channelID, modulePacket.Sequence)
		return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}
	// fetch the callback function
	callbackHandler, err := im.keeper.GetICACallbackHandler(module)
	if err != nil {
		return err
	}
	// call the callback
	if callbackHandler.HasICACallback(callbackData.CallbackId) {
		// if acknowledgement is empty, then it is a timeout
		err := callbackHandler.CallICACallback(ctx, callbackData.CallbackId, modulePacket, acknowledgement, callbackData.CallbackArgs)
		if err != nil {
			return err
		}
	}

	// remove the callback data
	// NOTE: Should we remove the callback data here, or above (conditional on HasICACallback == true)?
	im.keeper.RemoveCallbackData(ctx, callbackDataKey)
	return nil
}

// ###################################################################################
// 	Required functions to satisfy interface but not implemented for ICA auth modules
// ###################################################################################

// OnChanOpenTry implements the IBCModule interface
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
	panic("UNIMPLEMENTED")
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	panic("UNIMPLEMENTED")
}
