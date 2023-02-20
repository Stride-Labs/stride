package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
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
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, controllerConnectionId)
	if !found {
		return &types.MsgAddOracleResponse{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "connection %s not found", controllerConnectionId)
	}

	// Get chain id from the connection
	clientState, found := k.ICACallbacksKeeper.IBCKeeper.ClientKeeper.GetClientState(ctx, connectionEnd.ClientId)
	if !found {
		return &types.MsgAddOracleResponse{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "client %s not found", connectionEnd.ClientId)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return &types.MsgAddOracleResponse{}, types.ErrClientStateNotTendermint
	}
	chainId := client.ChainId

	// Create the oracle struct, marked as inactive
	oracle := types.Oracle{
		ChainId:      chainId,
		ConnectionId: controllerConnectionId,
		Active:       false,
	}
	k.SetOracle(ctx, oracle)

	// Get the corresponding connection on the host
	hostConnectionId := connectionEnd.Counterparty.ConnectionId
	if hostConnectionId == "" {
		return &types.MsgAddOracleResponse{}, types.ErrHostConnectionNotFound
	}

	// Register the oracle interchain account
	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: controllerConnectionId,
		HostConnectionId:       hostConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	owner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_Oracle)
	if err := k.ICAControllerKeeper.RegisterInterchainAccount(ctx, controllerConnectionId, owner, appVersion); err != nil {
		return &types.MsgAddOracleResponse{}, errorsmod.Wrapf(err, "unable to register oracle interchain account")
	}

	return &types.MsgAddOracleResponse{}, nil
}

// Instantiates the oracle cosmwasm contract
func (k msgServer) InstantiateOracle(goCtx context.Context, msg *types.MsgInstantiateOracle) (*types.MsgInstantiateOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO
	_ = ctx
	return &types.MsgInstantiateOracleResponse{}, nil
}

// Creates a new ICA channel and restores the oracle ICA account after a channel closer
func (k msgServer) RestoreOracleICA(goCtx context.Context, msg *types.MsgRestoreOracleICA) (*types.MsgRestoreOracleICAResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO
	_ = ctx
	return &types.MsgRestoreOracleICAResponse{}, nil
}
