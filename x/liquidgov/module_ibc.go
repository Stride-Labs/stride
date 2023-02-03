package liquidgov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"

	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"

	"github.com/Stride-Labs/stride/v5/x/icacallbacks"
	icacallbacktypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/keeper"
)

// IBCModule implements the ICS26 interface for interchain accounts controller chains
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

func (im IBCModule) Hooks() keeper.Hooks {
	return im.keeper.Hooks()
}

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
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck: portID %s, channelID %s", portID, channelID))
	// // Note: The channel capability must be claimed by the authentication module in OnChanOpenInit otherwise the
	// // authentication module will not be able to send packets on the channel created for the associated interchain account.
	// if err := im.keeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
	// 	return version, err
	// }
	// im.keeper.Logger(ctx).Info(fmt.Sprintf("%s claimed the channel capability %v", types.ModuleName, channelCap))
	// return version, nil

	version, err := im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		version,
	)
	return version, err
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck: portID %s, channelID %s, counterpartyChannelID %s, counterpartyVersion %s", portID, channelID, counterpartyChannelID, counterpartyVersion))
	// controllerConnectionId, err := im.keeper.GetConnectionId(ctx, portID)
	// if err != nil {
	// 	ctx.Logger().Error(fmt.Sprintf("Unable to get connection for port: %s", portID))
	// }
	// _, found := im.keeper.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	// if !found {
	// 	ctx.Logger().Error(fmt.Sprintf("Expected to find an address for %s/%s", controllerConnectionId, portID))
	// 	return nil
	// }
	// // get host chain id from connection
	// // fetch counterparty connection
	// _, err = im.keeper.GetChainID(ctx, controllerConnectionId)
	// if err != nil {
	// 	ctx.Logger().Error(fmt.Sprintf("Unable to obtain counterparty chain for connection: %s, port: %s, err: %s", controllerConnectionId, portID, err.Error()))
	// 	return nil
	// }
	// //  get zone info
	// zoneInfo, found := im.keeper.stakeibcKeeper.GetHostZone(ctx, hostChainId)
	// if !found {
	// 	ctx.Logger().Error(fmt.Sprintf("Expected to find zone info for %v", hostChainId))
	// 	return nil
	// }
	// ctx.Logger().Info(fmt.Sprintf("Found matching address for chain: %s, address %s, port %s", zoneInfo.ChainId, address, portID))

	// delegationAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_DELEGATION))
	// if err != nil {
	// 	return err
	// }

	// // Set ICA account addresses
	// switch {
	// // delegation address
	// case portID == delegationAddress:
	// 	zoneInfo.DelegationAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_DELEGATION}
	// default:
	// 	ctx.Logger().Error(fmt.Sprintf("Missing portId: %s", portID))
	// }

	// im.keeper.SetHostZone(ctx, zoneInfo)
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)

	// return nil
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnAcknowledgementPacket (Liquidgov) - packet: %+v, relayer: %v", modulePacket, relayer))

	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, im.keeper.Logger(ctx), acknowledgement, true)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unpack message data from acknowledgement, Sequence %d, from %s %s, to %s %s: %s",
			modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort, err.Error())
		im.keeper.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(icacallbacktypes.ErrInvalidAcknowledgement, errMsg)
	}

	err = im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, ackResponse)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to call registered callback from liquidgov OnAcknowledgePacket | Sequence %d, from %s %s, to %s %s",
			modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort)
		im.keeper.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(icacallbacktypes.ErrCallbackFailed, errMsg)
	}
	return im.app.OnAcknowledgementPacket(ctx, modulePacket, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnTimeoutPacket: packet %v, relayer %v", modulePacket, relayer))
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_TIMEOUT}
	err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, &ackResponse)
	if err != nil {
		return err
	}
	return im.app.OnTimeoutPacket(ctx, modulePacket, relayer)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {

	// WARNING: For some reason, in IBCv3 the ICA controller module does not call the underlying OnChanCloseConfirm (this function)
	// So, we need to put logic that _should_ execute upon channel closure in the OnTimeoutPacket function
	// This works because ORDERED channels are always closed when a timeout occurs, but if we migrate to using ORDERED channels that don't
	// close on timeout, we will need to move this logic to the OnChanCloseConfirm function
	// relevant IBCv3 code: https://github.com/cosmos/ibc-go/blob/5c0bf8b8a0f79643e36be98fb9883ea163d2d93a/modules/apps/27-interchain-accounts/controller/ibc_module.go#L123
	return nil
}

// ###################################################################################
// 	Helper functions
// ###################################################################################

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
