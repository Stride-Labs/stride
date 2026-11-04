package keeper

import (
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v11/modules/core/03-connection/types"
	ibctmtypes "github.com/cosmos/ibc-go/v11/modules/light-clients/07-tendermint"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

// OnChanOpenAck validates the channel's connection and ICA, but no longer updates host zone
// or trade route ICA addresses. All ICAs are registered for the wind-down, so their
// addresses are fixed.
func (k Keeper) OnChanOpenAck(ctx sdk.Context, portID, channelID string) error {
	// Lookup connection ID, counterparty chain ID, and ICA address from the channel ID
	controllerConnectionId, _, err := k.IBCKeeper.ChannelKeeper.GetChannelConnection(ctx, portID, channelID)
	if err != nil {
		return err
	}
	address, found := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("No ICA address associated with connection %s and port %s", controllerConnectionId, portID))
		return nil
	}
	chainId, err := k.GetChainIdFromConnectionId(ctx, controllerConnectionId)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("ICA channel opened for chain %s, address %s, port %s, channel %s",
		chainId, address, portID, channelID))

	return nil
}

// Given a connection ID, returns the light client time
func (k Keeper) GetLightClientTime(ctx sdk.Context, connectionID string) (clientTime uint64, err error) {
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return 0, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection-id: %s", connectionID)
	}

	latestHeight := k.IBCKeeper.ClientKeeper.GetClientLatestHeight(ctx, connection.ClientId)
	if latestHeight.IsZero() {
		return 0, errorsmod.Wrapf(clienttypes.ErrConsensusStateNotFound, "client-id: %s", connection.ClientId)
	}

	return k.IBCKeeper.ClientKeeper.GetClientTimestampAtHeight(ctx, connection.ClientId, latestHeight)
}

// Given a connection ID, returns the light client height
func (k Keeper) GetLightClientHeight(ctx sdk.Context, connectionID string) (height uint64, err error) {
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return 0, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection-id: %s", connectionID)
	}

	clientHeight := k.IBCKeeper.ClientKeeper.GetClientLatestHeight(ctx, connection.ClientId)
	if clientHeight.IsZero() {
		return 0, errorsmod.Wrapf(clienttypes.ErrConsensusStateNotFound, "client-id: %s", connection.ClientId)
	}

	return clientHeight.GetRevisionHeight(), nil
}

// Lookup a chain ID from a connection ID by looking up the client state
func (k Keeper) GetChainIdFromConnectionId(ctx sdk.Context, connectionID string) (string, error) {
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return "", errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection %s not found", connectionID)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return "", errorsmod.Wrapf(clienttypes.ErrClientNotFound, "client %s not found", connection.ClientId)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", types.ErrClientStateNotTendermint
	}

	return client.ChainId, nil
}
