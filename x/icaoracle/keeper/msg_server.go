package keeper

import (
	"context"
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proto "github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Adds a new oracle as a destination for metric updates
// Registers a new ICA account along this connection
func (k msgServer) AddOracle(goCtx context.Context, msg *types.MsgAddOracle) (*types.MsgAddOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Grab the connection and confirm it exists
	controllerConnectionId := msg.ConnectionId
	connectionEnd, found := k.ConnectionKeeper.GetConnection(ctx, controllerConnectionId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "connection (%s) not found", controllerConnectionId)
	}

	// Get chain id from the connection
	clientState, found := k.ClientKeeper.GetClientState(ctx, connectionEnd.ClientId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "client (%s) not found", connectionEnd.ClientId)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return nil, types.ErrClientStateNotTendermint
	}
	chainId := client.ChainId

	// Confirm oracle was not already created
	_, found = k.GetOracle(ctx, chainId)
	if found {
		return nil, types.ErrOracleAlreadyExists
	}

	// Create the oracle struct, marked as inactive
	oracle := types.Oracle{
		ChainId:      chainId,
		ConnectionId: controllerConnectionId,
		Active:       false,
	}
	k.SetOracle(ctx, oracle)

	// Get the expected port ID for the ICA channel
	owner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_Oracle)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return nil, err
	}

	// Check if an ICA account has already been created for this oracle
	// (in the event that an oracle was removed and then added back)
	// If so, there's no need to register a new ICA
	channelID, channelFound := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, controllerConnectionId, portID)
	icaAddress, icaFound := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)

	if channelFound && icaFound {
		oracle.IcaAddress = icaAddress
		oracle.ChannelId = channelID
		oracle.PortId = portID

		k.SetOracle(ctx, oracle)

		return &types.MsgAddOracleResponse{}, nil
	}

	// Get the corresponding connection on the host
	hostConnectionId := connectionEnd.Counterparty.ConnectionId
	if hostConnectionId == "" {
		return nil, types.ErrHostConnectionNotFound
	}

	// Register the oracle interchain account
	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: controllerConnectionId,
		HostConnectionId:       hostConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, controllerConnectionId, owner, appVersion); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register oracle interchain account")
	}

	return &types.MsgAddOracleResponse{}, nil
}

// Instantiates the oracle cosmwasm contract
func (k msgServer) InstantiateOracle(goCtx context.Context, msg *types.MsgInstantiateOracle) (*types.MsgInstantiateOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Confirm the oracle has already been added, but has not yet been instantiated
	oracle, found := k.GetOracle(ctx, msg.OracleChainId)
	if !found {
		return nil, types.ErrOracleNotFound
	}
	if oracle.ContractAddress != "" {
		return nil, types.ErrOracleAlreadyInstantiated
	}

	// Confirm the oracle ICA was registered
	if err := oracle.ValidateICASetup(); err != nil {
		return nil, err
	}

	// Build the contract-specific instantiation message
	contractMsg := types.MsgInstantiateOracleContract{
		AdminAddress:      oracle.IcaAddress,
		TransferChannelId: msg.TransferChannelOnOracle,
	}
	contractMsgBz, err := json.Marshal(contractMsg)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to marshal instantiate oracle contract")
	}

	// Build the ICA message to instantiate the contract
	msgs := []proto.Message{&types.MsgInstantiateContract{
		Sender: oracle.IcaAddress,
		Admin:  oracle.IcaAddress,
		CodeID: msg.ContractCodeId,
		Label:  "Stride ICA Oracle",
		Msg:    contractMsgBz,
	}}

	// Submit the ICA
	callbackArgs := types.InstantiateOracleCallback{
		OracleChainId: oracle.ChainId,
	}
	icaTx := types.ICATx{
		ConnectionId:    oracle.ConnectionId,
		ChannelId:       oracle.ChannelId,
		PortId:          oracle.PortId,
		Owner:           types.FormatICAAccountOwner(oracle.ChainId, types.ICAAccountType_Oracle),
		Messages:        msgs,
		RelativeTimeout: InstantiateOracleTimeout,
		CallbackArgs:    &callbackArgs,
		CallbackId:      ICACallbackID_InstantiateOracle,
	}
	if err := k.SubmitICATx(ctx, icaTx); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to submit instantiate oracle contract ICA")
	}

	return &types.MsgInstantiateOracleResponse{}, nil
}

// Creates a new ICA channel and restores the oracle ICA account after a channel closer
func (k msgServer) RestoreOracleICA(goCtx context.Context, msg *types.MsgRestoreOracleICA) (*types.MsgRestoreOracleICAResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Confirm the oracle exists and has already had an ICA registered
	oracle, found := k.GetOracle(ctx, msg.OracleChainId)
	if !found {
		return nil, types.ErrOracleNotFound
	}
	if err := oracle.ValidateICASetup(); err != nil {
		return nil, errorsmod.Wrapf(err, "the oracle (%s) has never had an registered ICA", oracle.ChainId)
	}

	// Confirm the channel is closed
	if k.IsOracleICAChannelOpen(ctx, oracle) {
		return nil, errorsmod.Wrapf(types.ErrUnableToRestoreICAChannel,
			"channel already open, chain-id: %s, channel-id: %s", oracle.ChainId, oracle.ChannelId)
	}

	// Grab the connectionEnd for the counterparty connection
	connectionEnd, found := k.ConnectionKeeper.GetConnection(ctx, oracle.ConnectionId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "connection (%s) not found", oracle.ConnectionId)
	}
	hostConnectionId := connectionEnd.Counterparty.ConnectionId

	// Only allow restoring an ICA if the account already exists
	owner := types.FormatICAAccountOwner(oracle.ChainId, types.ICAAccountType_Oracle)
	portId, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to build portId from owner (%s)", owner)
	}
	_, exists := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, oracle.ConnectionId, portId)
	if !exists {
		return nil, errorsmod.Wrapf(types.ErrICAAccountDoesNotExist,
			"cannot find ICA account for connection (%s) and port (%s)", oracle.ConnectionId, portId)
	}

	// Call register ICA again to restore the account
	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: oracle.ConnectionId,
		HostConnectionId:       hostConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, oracle.ConnectionId, owner, appVersion); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register oracle interchain account")
	}

	// Revert all pending metrics for this oracle back to status QUEUED
	for _, metric := range k.GetAllMetrics(ctx) {
		if metric.DestinationOracle == msg.OracleChainId && metric.Status == types.MetricStatus_IN_PROGRESS {
			k.UpdateMetricStatus(ctx, metric, types.MetricStatus_QUEUED)
		}
	}

	return &types.MsgRestoreOracleICAResponse{}, nil
}

// Proposal handler for toggling whether an oracle is currently active (meaning it's a destination for metric pushes)
func (ms msgServer) ToggleOracle(goCtx context.Context, msg *types.MsgToggleOracle) (*types.MsgToggleOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	if err := ms.Keeper.ToggleOracle(ctx, msg.OracleChainId, msg.Active); err != nil {
		return nil, err
	}

	return &types.MsgToggleOracleResponse{}, nil
}

// Proposal handler for removing an oracle from the store
func (ms msgServer) RemoveOracle(goCtx context.Context, msg *types.MsgRemoveOracle) (*types.MsgRemoveOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	_, found := ms.Keeper.GetOracle(ctx, msg.OracleChainId)
	if !found {
		return nil, types.ErrOracleNotFound
	}

	ms.Keeper.RemoveOracle(ctx, msg.OracleChainId)

	// Remove all metrics that were targeting this oracle
	for _, metric := range ms.Keeper.GetAllMetrics(ctx) {
		if metric.DestinationOracle == msg.OracleChainId {
			ms.Keeper.RemoveMetric(ctx, metric.GetMetricID())
		}
	}

	return &types.MsgRemoveOracleResponse{}, nil
}
