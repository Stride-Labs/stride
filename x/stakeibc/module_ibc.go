package stakeibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	icacallbacktypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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

func (im IBCModule) Hooks() keeper.Hooks {
	return im.keeper.Hooks()
}


// OnChanOpenInit is called when a channel is opened. It logs the portID and channelID, claims the channel
// capability, and logs that the channel capability has been claimed. It returns an error if the channel
// capability cannot be claimed.
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
    im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck: portID %s, channelID %s", portID, channelID))

    // Claim the channel capability
    err := im.keeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID))
    if err != nil {
        return err
    }

    // Log that the channel capability has been claimed
    im.keeper.Logger(ctx).Info(fmt.Sprintf("%s claimed the channel capability %v", types.ModuleName, channelCap))

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
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck: portID %s, channelID %s, counterpartyChannelID %s, counterpartyVersion %s", portID, channelID, counterpartyChannelID, counterpartyVersion))
	controllerConnectionId, err := im.keeper.GetConnectionId(ctx, portID)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to get connection for port: %s", portID))
	}
	address, found := im.keeper.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find an address for %s/%s", controllerConnectionId, portID))
		return nil
	}
	// get host chain id from connection
	// fetch counterparty connection
	hostChainId, err := im.keeper.GetChainID(ctx, controllerConnectionId)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to obtain counterparty chain for connection: %s, port: %s, err: %s", controllerConnectionId, portID, err.Error()))
		return nil
	}
	//  get zone info
	zoneInfo, found := im.keeper.GetHostZone(ctx, hostChainId)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find zone info for %v", hostChainId))
		return nil
	}
	ctx.Logger().Info(fmt.Sprintf("Found matching address for chain: %s, address %s, port %s", zoneInfo.ChainId, address, portID))

	// addresses
	withdrawalAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_WITHDRAWAL))
	if err != nil {
		return err
	}
	feeAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_FEE))
	if err != nil {
		return err
	}
	delegationAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_DELEGATION))
	if err != nil {
		return err
	}
	redemptionAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_REDEMPTION))
	if err != nil {
		return err
	}

	// Set ICA account addresses
	switch {
	// withdrawal address
	case portID == withdrawalAddress:
		zoneInfo.WithdrawalAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_WITHDRAWAL}
	// fee address
	case portID == feeAddress:
		zoneInfo.FeeAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_FEE}
	// delegation address
	case portID == delegationAddress:
		zoneInfo.DelegationAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_DELEGATION}
	case portID == redemptionAddress:
		zoneInfo.RedemptionAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_REDEMPTION}
	default:
		ctx.Logger().Error(fmt.Sprintf("Missing portId: %s", portID))
	}

	im.keeper.SetHostZone(ctx, zoneInfo)
	return nil
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// Log information about the acknowledgement packet
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnAcknowledgementPacket: packet %v, relayer %v", modulePacket, relayer))
	ackInfo := fmt.Sprintf("sequence #%d, from %s %s, to %s %s",
		modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort)

	// Unmarshal the acknowledgement
	var ack channeltypes.Acknowledgement
	err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unmarshal ack from stakeibc OnAcknowledgePacket | Sequence %d, from %s %s, to %s %s",
			modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort)
		im.keeper.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}
	
	
	// Log the unmarshalled acknowledgement
	im.keeper.Logger(ctx).Info(fmt.Sprintf("Acknowledgement was successfully unmarshalled: ackInfo: %s", ackInfo))	
	
// Emit an event
eventType := "ack"
ctx.EventManager().EmitEvent(
    	sdk.NewEvent(
        	eventType,
        	sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
        	sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ackInfo)),
    ),
)


// Call the registered callback function
err = im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, &ack)
if err != nil {
    	errMsg := fmt.Sprintf("Unable to call registered callback from stakeibc OnAcknowledgePacket | Sequence %d, from %s %s, to %s %s",
        modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort)
    	im.keeper.Logger(ctx).Error(errMsg)
    	return sdkerrors.Wrapf(icacallbacktypes.ErrCallbackFailed, errMsg)
	}

return nil
}

// OnTimeoutPacket is called when a timeout packet is received. It logs the packet and relayer, calls the
// CallRegisteredICACallback function with the packet and nil as arguments, and returns an error if calling the
// CallRegisteredICACallback function fails.
func (im IBCModule) OnTimeoutPacket(
    ctx sdk.Context,
    modulePacket channeltypes.Packet,
    relayer sdk.AccAddress,
) error {
    // Log the packet and relayer
    im.keeper.Logger(ctx).Info(fmt.Sprintf("OnTimeoutPacket: packet %v, relayer %v", modulePacket, relayer))

    // Call the CallRegisteredICACallback function with the packet and nil as arguments
    err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, nil)
    if err != nil {
        return err
    }

    return nil
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

// NegotiateAppVersion takes as input a channeltypes.Order, a connectionID, a portID, a channeltypes.Counterparty,
// and a proposedVersion. It returns the proposedVersion and nil as the version and error values, respectively.
func (im IBCModule) NegotiateAppVersion(
    ctx sdk.Context,
    order channeltypes.Order,
    connectionID string,
    portID string,
    counterparty channeltypes.Counterparty,
    proposedVersion string,
) (version string, err error) {
    // Return the proposed version and nil as the error value
    return proposedVersion, nil
}


// ###################################################################################
// 	Required functions to satisfy interface but not implemented for ICA auth modules
// ###################################################################################

// OnChanOpenTry is called when a channel open try is received. It panics with the message "UNIMPLEMENTED".
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
    // Panic with the message "UNIMPLEMENTED"
    panic("UNIMPLEMENTED")
}


// OnChanOpenConfirm is called when a channel open confirmation is received. It panics with the message "UNIMPLEMENTED".
func (im IBCModule) OnChanOpenConfirm(
    ctx sdk.Context,
    portID,
    channelID string,
) error {
    // Panic with the message "UNIMPLEMENTED"
    panic("UNIMPLEMENTED")
}


// OnChanCloseInit is called when a channel close initiation is received. It panics with the message "UNIMPLEMENTED".
func (im IBCModule) OnChanCloseInit(
    ctx sdk.Context,
    portID,
    channelID string,
) error {
    // Panic with the message "UNIMPLEMENTED"
    panic("UNIMPLEMENTED")
}

// OnRecvPacket is called when a packet is received. It panics with the message "UNIMPLEMENTED".
func (im IBCModule) OnRecvPacket(
    ctx sdk.Context,
    modulePacket channeltypes.Packet,
    relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
    // Panic with the message "UNIMPLEMENTED"
    panic("UNIMPLEMENTED")
}

