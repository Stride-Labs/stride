package keeper

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/Stride-Labs/stride/v11/x/icacallbacks"
	icacallbacktypes "github.com/Stride-Labs/stride/v11/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v11/x/icaoracle/types"
)

func (k Keeper) OnChanOpenInit(ctx sdk.Context, portID, channelID string, channelCap *capabilitytypes.Capability) error {
	// TODO: Update IBC-go to v6/v7 and then there's no longer a need to claim the channel capability here
	// Until then, we need to make sure we only claim for oracle ports
	if strings.Contains(portID, types.ICAAccountType_Oracle) {
		k.Logger(ctx).Info(fmt.Sprintf("%s claimed the channel capability for %s %s", types.ModuleName, channelID, portID))
		return k.scopedKeeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID))
	}
	return nil
}

func (k Keeper) OnChanOpenAck(ctx sdk.Context, portID, channelID string) error {
	// Get the connectionId from the port and channel
	connectionId, _, err := k.IBCKeeper.ChannelKeeper.GetChannelConnection(ctx, portID, channelID)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get connection from channel (%s) and port (%s)", channelID, portID)
	}

	// If the callback is not for an oracle ICA, it should do nothing and then pass the ack down to stakeibc
	oracle, found := k.GetOracleFromConnectionId(ctx, connectionId)
	if !found {
		return nil
	}
	expectedOraclePort, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(oracle.ChainId, types.ICAAccountType_Oracle))
	if err != nil {
		return err
	}
	if portID != expectedOraclePort {
		return nil
	}

	// If this callback is for an oracle channel, store the ICA address and channel on the oracle struct
	// Get the associated ICA address from the port and connection
	icaAddress, found := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, connectionId, portID)
	if !found {
		return errorsmod.Wrapf(icatypes.ErrInterchainAccountNotFound, "unable to get ica address from connection (%s)", connectionId)
	}
	k.Logger(ctx).Info(fmt.Sprintf("Oracle ICA registered to channel %s and address %s", channelID, icaAddress))

	// Update the ICA address and channel in the oracle
	oracle.IcaAddress = icaAddress
	oracle.ChannelId = channelID
	oracle.PortId = portID

	k.SetOracle(ctx, oracle)

	return nil
}

func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) error {
	// If this is not an oracle packet, pass the ack down to the next middleware
	if !strings.Contains(packet.SourcePort, types.ICAAccountType_Oracle) {
		return nil
	}

	// Unpack the acknowledgement into success/error
	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, k.Logger(ctx), acknowledgement, true)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unpack message data from acknowledgement, Sequence %d, from %s %s, to %s %s: %s",
			packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort, err.Error())
		return errorsmod.Wrapf(icacallbacktypes.ErrInvalidAcknowledgement, errMsg)
	}

	ackInfo := fmt.Sprintf("sequence #%d, from %s %s, to %s %s",
		packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)
	k.Logger(ctx).Info(fmt.Sprintf("Acknowledgement was successfully unmarshalled: ackInfo: %s", ackInfo))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			icacallbacktypes.EventTypeAck,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(icacallbacktypes.AttributeKeyAck, ackInfo),
		),
	)

	if err := k.ICACallbacksKeeper.CallRegisteredICACallback(ctx, packet, ackResponse); err != nil {
		errMsg := fmt.Sprintf("Unable to call registered callback from stakeibc OnAcknowledgePacket | Sequence %d, from %s %s, to %s %s",
			packet.Sequence, packet.SourceChannel, packet.SourcePort, packet.DestinationChannel, packet.DestinationPort)
		return errorsmod.Wrapf(icacallbacktypes.ErrCallbackFailed, errMsg)
	}

	return nil
}

func (k Keeper) SubmitICATx(ctx sdk.Context, tx types.ICATx) error {
	// Validate the ICATx struct has all the required fields
	if err := tx.ValidateICATx(ctx); err != nil {
		return err
	}

	// Serialize tx messages
	txBz, err := icatypes.SerializeCosmosTx(k.cdc, tx.Messages)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to serialize cosmos transaction")
	}
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: txBz,
	}

	// TODO: After upgrading ibc-go to v6/v7, we can just pass nil as the chanCap in SendTx
	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(tx.PortId, tx.ChannelId))
	if !found {
		return errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	// Submit ICA and grab sequence number for the callback key
	sequence, err := k.ICAControllerKeeper.SendTx(ctx, chanCap, tx.ConnectionId, tx.PortId, packetData, tx.Timeout)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to submit ICA transaction")
	}

	// Store the callback data
	callbackArgsBz, err := proto.Marshal(tx.CallbackArgs)
	if err != nil {
		return errorsmod.Wrapf(types.ErrMarshalFailure, "unable to marshal callback: %s", err.Error())
	}
	callbackData := icacallbacktypes.CallbackData{
		CallbackKey:  icacallbacktypes.PacketID(tx.PortId, tx.ChannelId, sequence),
		PortId:       tx.PortId,
		ChannelId:    tx.ChannelId,
		Sequence:     sequence,
		CallbackId:   tx.CallbackId,
		CallbackArgs: callbackArgsBz,
	}
	k.ICACallbacksKeeper.SetCallbackData(ctx, callbackData)

	return nil
}
