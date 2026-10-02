package keeper

import (
	"context"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v35/x/icaoracle/types"
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
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, oracle.ConnectionId, owner, appVersion, channeltypes.ORDERED); err != nil {
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
