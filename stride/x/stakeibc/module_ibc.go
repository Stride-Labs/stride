package stakeibc

import (
	"fmt"
	"strings"

	"github.com/Stride-labs/stride/x/stakeibc/keeper"
	"github.com/Stride-labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
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
	// Scaffolded for ibc v2
	// Require portID is the portID module is bound to
	// boundPort := im.keeper.GetPort(ctx)
	// if boundPort != portID {
	// 	return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	// }

	// if version != types.Version {
	// 	return sdkerrors.Wrapf(types.ErrInvalidVersion, "got %s, expected %s", version, types.Version)
	// }

	// // Claim channel capability passed back by IBC module
	// if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
	// 	return err
	// }

	// return nil
	// ibc v3 https://github.com/cosmos/ibc/blob/f19c5d188d6de301d10a212406155cbb2ba2982f/spec/app/ics-027-interchain-accounts/README.md
	return im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID))
}

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
	// ibc v3
	return "", nil

	// ibc v2 scaffolded code
	// Require portID is the portID module is bound to
	// boundPort := im.keeper.GetPort(ctx)
	// if boundPort != portID {
	// 	return sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	// }

	// if version != types.Version {
	// 	return sdkerrors.Wrapf(types.ErrInvalidVersion, "got: %s, expected %s", version, types.Version)
	// }

	// if counterpartyVersion != types.Version {
	// 	return sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s", counterpartyVersion, types.Version)
	// }

	// // Module may have already claimed capability in OnChanOpenInit in the case of crossing hellos
	// // (ie chainA and chainB both call ChanOpenInit before one of them calls ChanOpenTry)
	// // If module can already authenticate the capability then module already owns it so we don't need to claim
	// // Otherwise, module does not have channel capability and we must claim it from IBC
	// if !im.keeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
	// 	// Only claim channel capability passed back by IBC module if we do not already own it
	// 	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
	// 		return err
	// 	}
	// }

	// return nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// TODO(TEST-21): Implement this!
	if counterpartyVersion != types.Version {
		return sdkerrors.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}

	// Get HostZone from the chain associated with the connection for the input port
	// Port has a connection => connection has an ICA => chain has a HostZone

	// Get connection from port
	connectionId, err := im.keeper.GetConnectionForPort(ctx, portID)
	if err != nil {
		ctx.Logger().Error("Failed to get connectionID for port " + portID)
	}
	// Get ICA for connection
	address, found := im.keeper.ICAControllerKeeper.GetInterchainAccountAddress(ctx, connectionId, portID)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find an address for %s/%s", connectionId, portID))
	}

	// Get chain for connection
	chainID, err := im.keeper.GetChainID(ctx, connectionId)
	if err != nil {
		ctx.Logger().Error(
			"Unable to obtain chain for given connection and port",
			"ConnectionID", connectionId,
			"PortID", portID,
			"Error", err,
		)
		return nil
	}

	// Get zone info for chain
	zoneInfo, found := im.keeper.GetHostZoneInfo(ctx, chainID)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Unable to find host zone info for %v", chainID))
		return nil
	}

	// found a matching zone!
	ctx.Logger().Info(fmt.Sprintf("Found address matching zone %s: %s (port: %s)", zoneInfo.Id, address, portID))
	portParts := strings.Split(portID, ".")

	// SETUP HOSTZONE ACCOUNTS FOR THIS ZONE
	// TODO(TEST-43) understand portID nomenclature for this switch statement
	switch {
	case portParts[1] == "validator":
		// set HostZone validators
		// TODO(TEST-42) create validator setup helpers and replace dummy validator init values below
		validator := &types.Validator{Name: "", Address: address, Status: "active", CommissionRate: 0, DelegationAmt: 0}
		zoneInfo.Validators = append(zoneInfo.Validators, validator)
	case portParts[1] == "delegation":
		// set HostZone delegationAccounts
		for _, existing := range zoneInfo.DelegationAccounts {
			if existing.Address == address {
				ctx.Logger().Error("Address is already a delegation address: " + address)
			}
		}
		// TODO(TEST-) create delegation account setup helpers and replace dummy delegation account init values below
		delegationAccount := &types.ICAAccount{Address: address, Balance: 0, DelegatedBalance: 0, Delegations: []*types.Delegation{}}
		zoneInfo.DelegationAccounts = append(zoneInfo.DelegationAccounts, delegationAccount)
	case portParts[1] == "feeAccount":
		// set HostZone feeAccount
		// TODO (TEST-) add Fee account safety checks (we can't have this changing without rigorous validation)
		zoneInfo.FeeAccount = address
	default:
		ctx.Logger().Error("Unrecognized channel for portID: " + portID)
	}
	im.keeper.SetHostZone(ctx, zoneInfo)
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for channels
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	var ack channeltypes.Acknowledgement

	// this line is used by starport scaffolding # oracle/packet/module/recv

	var modulePacketData types.StakeibcPacketData
	if err := modulePacketData.Unmarshal(modulePacket.GetData()); err != nil {
		return channeltypes.NewErrorAcknowledgement(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error()).Error())
	}

	// Dispatch packet
	switch packet := modulePacketData.Packet.(type) {
	// this line is used by starport scaffolding # ibc/packet/module/recv
	default:
		errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
		return channeltypes.NewErrorAcknowledgement(errMsg)
	}

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet acknowledgement: %v", err)
	}

	// this line is used by starport scaffolding # oracle/packet/module/ack

	var modulePacketData types.StakeibcPacketData
	if err := modulePacketData.Unmarshal(modulePacket.GetData()); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	var eventType string

	// Dispatch packet
	switch packet := modulePacketData.Packet.(type) {
	// this line is used by starport scaffolding # ibc/packet/module/ack
	default:
		errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
		return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyAck, fmt.Sprintf("%v", ack)),
		),
	)

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				eventType,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				eventType,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	var modulePacketData types.StakeibcPacketData
	if err := modulePacketData.Unmarshal(modulePacket.GetData()); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	// Dispatch packet
	switch packet := modulePacketData.Packet.(type) {
	// this line is used by starport scaffolding # ibc/packet/module/timeout
	default:
		errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
		return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
	}

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
