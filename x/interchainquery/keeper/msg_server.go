package keeper

import (
	"context"
	"net/url"
	"sort"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ics23 "github.com/cosmos/ics23/go"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/interchainquery/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: &keeper}
}

var _ types.MsgServer = msgServer{}

// check if the query requires proving; if it does, verify it!
func (k Keeper) VerifyKeyProof(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, query types.Query) error {
	pathParts := strings.Split(query.QueryType, "/")

	// the query does NOT have an associated proof, so no need to verify it.
	if pathParts[len(pathParts)-1] != "key" {
		return nil
	}

	// If the query is a "key" proof query, verify the results are valid by checking the poof
	if msg.ProofOps == nil {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to validate proof. No proof submitted")
	}

	// Get the client consensus state at the height 1 block above the message height
	msgHeight, err := cast.ToUint64E(msg.Height)
	if err != nil {
		return err
	}
	height := clienttypes.NewHeight(clienttypes.ParseChainID(query.ChainId), msgHeight+1)

	// Get the client state and consensus state from the connection Id
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, query.ConnectionId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "ConnectionId %s does not exist", query.ConnectionId)
	}
	consensusState, found := k.IBCKeeper.ClientKeeper.GetClientConsensusState(ctx, connection.ClientId, height)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Consensus state not found for client %s and height %d", connection.ClientId, height)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to fetch client state for client %s", connection.ClientId)
	}

	// Cast the client and consensus state to tendermint type
	tendermintConsensusState, ok := consensusState.(*tendermint.ConsensusState)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidICQProof,
			"Only tendermint consensus state is supported (%s provided)", consensusState.ClientType())
	}
	tendermintClientState, ok := clientState.(*tendermint.ClientState)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidICQProof,
			"Only tendermint client state is supported (%s provided)", clientState.ClientType())
	}
	var stateRoot exported.Root = tendermintConsensusState.Root
	var clientStateProof []*ics23.ProofSpec = tendermintClientState.ProofSpecs

	// Get the merkle path and merkle proof
	path := commitmenttypes.NewMerklePath([]string{pathParts[1], url.PathEscape(string(query.Request))}...)
	merkleProof, err := commitmenttypes.ConvertProofs(msg.ProofOps)
	if err != nil {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Error converting proofs: %s", err.Error())
	}

	// If we got a non-nil response, verify inclusion proof
	if len(msg.Result) != 0 {
		if err := merkleProof.VerifyMembership(clientStateProof, stateRoot, path, msg.Result); err != nil {
			return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to verify membership proof: %s", err.Error())
		}
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, query.CallbackId, "Inclusion proof validated - QueryId %s", query.Id))

	} else {
		// if we got a nil query response, verify non inclusion proof.
		if err := merkleProof.VerifyNonMembership(clientStateProof, stateRoot, path); err != nil {
			return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to verify non-membership proof: %s", err.Error())
		}
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, query.CallbackId, "Non-inclusion proof validated - QueryId %s", query.Id))
	}

	return nil
}

// call the query's associated callback function
func (k Keeper) InvokeCallback(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, query types.Query) error {
	// get all the callback handlers and sort them for determinism
	// (each module has their own callback handler)
	moduleNames := []string{}
	for moduleName := range k.callbacks {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)

	// Loop through each module until the callbackId is found in one of the module handlers
	for _, moduleName := range moduleNames {
		moduleCallbackHandler := k.callbacks[moduleName]

		// Once the callback is found, invoke the function
		if moduleCallbackHandler.HasICQCallback(query.CallbackId) {
			return moduleCallbackHandler.CallICQCallback(ctx, query.CallbackId, msg.Result, query)
		}
	}

	// If no callback was found, return an error
	return types.ErrICQCallbackNotFound
}

// Handle ICQ query responses by validating the proof, and calling the query's corresponding callback
func (k msgServer) SubmitQueryResponse(goCtx context.Context, msg *types.MsgSubmitQueryResponse) (*types.MsgSubmitQueryResponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the response has an associated query stored on stride
	query, found := k.GetQuery(ctx, msg.QueryId)
	if !found {
		k.Logger(ctx).Info("ICQ RESPONSE  | Ignoring non-existent query response (note: duplicate responses are nonexistent)")
		return &types.MsgSubmitQueryResponseResponse{}, nil // technically this is an error, but will cause the entire tx to fail if we have one 'bad' message, so we can just no-op here.
	}

	defer ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
		),
		sdk.NewEvent(
			"query_response",
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
			sdk.NewAttribute(types.AttributeKeyChainId, query.ChainId),
		),
	})

	// Verify the response's proof, if one exists
	err := k.VerifyKeyProof(ctx, msg, query)
	if err != nil {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(query.ChainId, query.CallbackId,
			"QUERY PROOF VERIFICATION FAILED - QueryId: %s, Error: %s", query.Id, err.Error()))
		return nil, err
	}

	// Immediately delete the query so it cannot process again
	k.DeleteQuery(ctx, query.Id)

	// Verify the query hasn't expired (if the block time is greater than the TTL timestamp, the query is expired)
	currBlockTime, err := cast.ToUint64E(ctx.BlockTime().UnixNano())
	if err != nil {
		return nil, err
	}
	if query.Ttl < currBlockTime {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(query.ChainId, query.CallbackId,
			"QUERY TIMEOUT - QueryId: %s, TTL: %d, BlockTime: %d", query.Id, query.Ttl, ctx.BlockHeader().Time.UnixNano()))
		return &types.MsgSubmitQueryResponseResponse{}, nil
	}

	// If the query is contentless, end
	if len(msg.Result) == 0 {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, query.CallbackId,
			"Query response is contentless - QueryId: %s", query.Id))
		return &types.MsgSubmitQueryResponseResponse{}, nil
	}

	// Call the query's associated callback function
	err = k.InvokeCallback(ctx, msg, query)
	if err != nil {
		k.Logger(ctx).Error(utils.LogICQCallbackWithHostZone(query.ChainId, query.CallbackId,
			"Error invoking ICQ callback, error: %s, QueryId: %s, QueryType: %s, ConnectionId: %s, QueryRequest: %v, QueryReponse: %v",
			err.Error(), msg.QueryId, query.QueryType, query.ConnectionId, query.Request, msg.Result))
		return nil, err
	}

	return &types.MsgSubmitQueryResponseResponse{}, nil
}
