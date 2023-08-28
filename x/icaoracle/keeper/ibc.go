package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func (k Keeper) OnChanOpenAck(ctx sdk.Context, portID, channelID string) error {
	// Get the connectionId from the port and channel
	connectionId, _, err := k.ChannelKeeper.GetChannelConnection(ctx, portID, channelID)
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

func (k Keeper) SubmitICATx(ctx sdk.Context, tx types.ICATx) error {
	// Validate the ICATx struct has all the required fields
	if err := tx.ValidateICATx(); err != nil {
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

	// Submit ICA and grab sequence number for the callback key
	icaMsgServer := icacontrollerkeeper.NewMsgServerImpl(&k.ICAControllerKeeper)
	msgSendTx := icacontrollertypes.NewMsgSendTx(tx.Owner, tx.ConnectionId, tx.GetRelativeTimeoutNano(), packetData)
	res, err := icaMsgServer.SendTx(ctx, msgSendTx)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to send ICA tx")
	}
	sequence := res.Sequence

	// Store the callback data
	callbackArgsBz, err := proto.Marshal(tx.CallbackArgs)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal callback")
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
