package icacallbacks

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/Stride-Labs/stride/v14/x/icacallbacks/keeper"
	"github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
)

var _ porttypes.IBCModule = &IBCModule{}

type IBCModule struct {
	keeper keeper.Keeper
}

func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// No custom logic is necessary in OnChanOpenInit
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
	return version, nil
}

// OnChanOpenTry should not be executed in the ICA stack
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

// No custom logic is necessary in OnChanOpenAck
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return nil
}

// OnChanOpenConfirm should not be executed in the ICA stack
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// OnChanCloseInit should not be executed in the ICA stack
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// No custom logic is necessary in OnChanCloseConfirm
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanOpenAck routes the packet to the relevant callback function
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnAcknowledgementPacket (ICACallbacks) - packet: %+v, relayer: %v", modulePacket, relayer))

	ackResponse, err := UnpackAcknowledgementResponse(ctx, im.keeper.Logger(ctx), acknowledgement, true)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unpack message data from acknowledgement, Sequence %d, from %s %s, to %s %s: %s",
			modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort, err.Error())
		im.keeper.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(types.ErrInvalidAcknowledgement, errMsg)
	}

	ackInfo := fmt.Sprintf("sequence #%d, from %s %s, to %s %s",
		modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort)
	im.keeper.Logger(ctx).Info(fmt.Sprintf("Acknowledgement was successfully unmarshalled: ackInfo: %s", ackInfo))

	eventType := "ack"
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyAck, ackInfo),
		),
	)

	if err := im.keeper.CallRegisteredICACallback(ctx, modulePacket, ackResponse); err != nil {
		errMsg := fmt.Sprintf("Unable to call registered ICACallback from OnAcknowledgePacket | Sequence %d, from %s %s, to %s %s",
			modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort)
		im.keeper.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(types.ErrCallbackFailed, errMsg)
	}
	return nil
}

// OnTimeoutPacket routes the timeout to the relevant callback function
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnTimeoutPacket (ICACallbacks): packet %v, relayer %v", packet, relayer))

	ackResponse := types.AcknowledgementResponse{
		Status: types.AckResponseStatus_TIMEOUT,
	}

	if err := im.keeper.CallRegisteredICACallback(ctx, packet, &ackResponse); err != nil {
		errMsg := fmt.Sprintf("Unable to call registered ICACallback from OnTimeoutPacket, Packet: %+v", packet)
		im.keeper.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(types.ErrCallbackFailed, errMsg)
	}
	return nil
}

// OnRecvPacket should not be executed in the ICA stack
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	panic("UNIMPLEMENTED")
}

// No custom logic required in NegotiateAppVersion
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
