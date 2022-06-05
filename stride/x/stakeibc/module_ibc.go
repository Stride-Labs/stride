package stakeibc

import (
	"fmt"

	"github.com/Stride-Labs/stride/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"
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

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	// Note: The channel capability must be claimed by the authentication module in OnChanOpenInit otherwise the
	// authentication module will not be able to send packets on the channel created for the associated interchain account.
	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
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
	// TODO(TEST-21): Implement this! The `counterpartyVersion != types.Version` is causing errors
	// if counterpartyVersion != types.Version {
	// 	return sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	// }
	controllerConnectionId, err := im.keeper.GetConnectionId(ctx, portID)
	if err != nil {
		ctx.Logger().Error("Unable to get connection for port " + portID)
	}
	address, found := im.keeper.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find an address for %s/%s", controllerConnectionId, portID))
		return nil
	}
	// get host chain id from connection
	// fetch counterparty connection
	hostChainId, err := im.keeper.GetCounterpartyChainId(ctx, controllerConnectionId)
	if err != nil {
		ctx.Logger().Error(
			"Unable to obtain counterparty chain for given connection and port",
			"ConnectionID", controllerConnectionId,
			"PortID", portID,
			"Error", err,
		)
		return nil
	}
	//  get zone info
	zoneInfo, found := im.keeper.GetHostZone(ctx, hostChainId)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find zone info for %v", hostChainId))
		return nil
	}
	ctx.Logger().Info("Found matching address", "chain", zoneInfo.ChainId, "address", address, "port", portID)

	// addresses
	withdrawalAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_WITHDRAWAL))
	feeAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_FEE))
	delegationAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_DELEGATION))

	// Set ICA account addresses
	switch {
	// withdrawal address
	case portID == withdrawalAddress:
		zoneInfo.WithdrawalAccount = &types.ICAAccount{Address: address, Balance: 0, DelegatedBalance: 0, Target: types.ICAAccountType_WITHDRAWAL}
	// fee address
	case portID == feeAddress:
		zoneInfo.FeeAccount = &types.ICAAccount{Address: address, Balance: 0, DelegatedBalance: 0, Target: types.ICAAccountType_FEE}
	// delegation address
	case portID == delegationAddress:
		zoneInfo.DelegationAccount = &types.ICAAccount{Address: address, Balance: 0, DelegatedBalance: 0, Target: types.ICAAccountType_DELEGATION}
	default:
		ctx.Logger().Error("Missing portId: ", portID)
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
	return nil
	// TODO(TEST-21): Implement OnAcknowledgementPacket logic
	
	// var ack channeltypes.Acknowledgement
	// if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
	// 	return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet acknowledgement: %v", err)
	// }

	// // this line is used by starport scaffolding # oracle/packet/module/ack

	// var modulePacketData types.StakeibcPacketData
	// if err := modulePacketData.Unmarshal(modulePacket.GetData()); err != nil {
	// 	return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	// }

	// var eventType string

	// // Dispatch packet
	// switch packet := modulePacketData.Packet.(type) {
	// // this line is used by starport scaffolding # ibc/packet/module/ack
	// default:
	// 	errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
	// 	return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	// }

	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		eventType,
	// 		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
	// 		sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ack)),
	// 	),
	// )

	// switch resp := ack.Response.(type) {
	// case *channeltypes.Acknowledgement_Result:
	// 	ctx.EventManager().EmitEvent(
	// 		sdk.NewEvent(
	// 			eventType,
	// 			sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
	// 		),
	// 	)
	// case *channeltypes.Acknowledgement_Error:
	// 	ctx.EventManager().EmitEvent(
	// 		sdk.NewEvent(
	// 			eventType,
	// 			sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
	// 		),
	// 	)
	// }

	// return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return nil
	// TODO(TEST-21): Implement OnTimeoutPacket logic
	// var modulePacketData types.StakeibcPacketData
	// if err := modulePacketData.Unmarshal(modulePacket.GetData()); err != nil {
	// 	return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	// }

	// // Dispatch packet
	// switch packet := modulePacketData.Packet.(type) {
	// // this line is used by starport scaffolding # ibc/packet/module/timeout
	// default:
	// 	errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
	// 	return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	// }

	// return nil
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

// #############################################################################
// 	Required functions to satisfy interface but not implemented for ICA
// #############################################################################

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

