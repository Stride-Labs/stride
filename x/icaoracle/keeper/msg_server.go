package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
// Registers a new ICA account along this connection and instantiates the oracle's cosmwasm contract in the callback
func (k msgServer) AddOracle(goCtx context.Context, msg *types.MsgAddOracle) (*types.MsgAddOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO
	_ = ctx

	return &types.MsgAddOracleResponse{}, nil
}

// Creates a new ICA channel and restores the oracle ICA account after a channel closer
func (k msgServer) RestoreOracleICA(goCtx context.Context, msg *types.MsgRestoreOracleICA) (*types.MsgRestoreOracleICAResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO
	_ = ctx
	return &types.MsgRestoreOracleICAResponse{}, nil
}
