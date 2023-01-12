package keeper

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	tmclienttypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/x/interchainquery/types"
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
func (k Keeper) VerifyKeyProof(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, q types.Query) error {
	pathParts := strings.Split(q.QueryType, "/")

	// the query does NOT have an associated proof, so no need to verify it.
	if pathParts[len(pathParts)-1] != "key" {
		return nil
	} else {
		// the query is a "key" proof query -- verify the results are valid by checking the proof!
		if msg.ProofOps == nil {
			errMsg := fmt.Sprintf("[ICQ Resp] for query %s, unable to validate proof. No proof submitted", q.Id)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrInvalidICQProof, errMsg)
		}
		connection, _ := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, q.ConnectionId)

		msgHeight, err := cast.ToUint64E(msg.Height)
		if err != nil {
			return err
		}
		height := clienttypes.NewHeight(clienttypes.ParseChainID(q.ChainId), msgHeight+1)
		consensusState, found := k.IBCKeeper.ClientKeeper.GetClientConsensusState(ctx, connection.ClientId, height)
		if !found {
			errMsg := fmt.Sprintf("[ICQ Resp] for query %s, consensus state not found for client %s and height %d", q.Id, connection.ClientId, height)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrInvalidICQProof, errMsg)
		}

		clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
		if !found {
			errMsg := fmt.Sprintf("[ICQ Resp] for query %s, unable to fetch client state for client %s and height %d", q.Id, connection.ClientId, height)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrInvalidICQProof, errMsg)
		}
		path := commitmenttypes.NewMerklePath([]string{pathParts[1], url.PathEscape(string(q.Request))}...)

		merkleProof, err := commitmenttypes.ConvertProofs(msg.ProofOps)
		if err != nil {
			errMsg := fmt.Sprintf("[ICQ Resp] for query %s, error converting proofs", q.Id)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrInvalidICQProof, errMsg)
		}

		tmclientstate, ok := clientState.(*tmclienttypes.ClientState)
		if !ok {
			errMsg := fmt.Sprintf("[ICQ Resp] for query %s, error unmarshaling client state %v", q.Id, clientState)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrInvalidICQProof, errMsg)
		}

		if len(msg.Result) != 0 {
			// if we got a non-nil response, verify inclusion proof.
			if err := merkleProof.VerifyMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path, msg.Result); err != nil {
				errMsg := fmt.Sprintf("[ICQ Resp] for query %s, unable to verify membership proof: %s", q.Id, err)
				k.Logger(ctx).Error(errMsg)
				return sdkerrors.Wrapf(types.ErrInvalidICQProof, errMsg)
			}
			k.Logger(ctx).Info(fmt.Sprintf("Proof validated! module: %s, queryId %s", types.ModuleName, q.Id))

		} else {
			// if we got a nil response, verify non inclusion proof.
			if err := merkleProof.VerifyNonMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path); err != nil {
				errMsg := fmt.Sprintf("[ICQ Resp] for query %s, unable to verify non-membership proof: %s", q.Id, err)
				k.Logger(ctx).Error(errMsg)
				return sdkerrors.Wrapf(types.ErrInvalidICQProof, errMsg)
			}
			k.Logger(ctx).Info(fmt.Sprintf("Non-inclusion Proof validated, stopping here! module: %s, queryId %s", types.ModuleName, q.Id))
		}
	}
	return nil
}

// call the query's associated callback function
func (k Keeper) InvokeCallback(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, q types.Query) error {
	// get all the stored queries and sort them for determinism
	moduleNames := []string{}
	for moduleName := range k.callbacks {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)

	for _, moduleName := range moduleNames {
		k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] executing callback for queryId (%s), module (%s)", q.Id, moduleName))
		moduleCallbackHandler := k.callbacks[moduleName]

		if moduleCallbackHandler.HasICQCallback(q.CallbackId) {
			k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] callback (%s) found for module (%s)", q.CallbackId, moduleName))
			// call the correct callback function
			err := moduleCallbackHandler.CallICQCallback(ctx, q.CallbackId, msg.Result, q)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("[ICQ Resp] error in ICQ callback, error: %s, msg: %s, result: %v, type: %s, params: %v", err.Error(), msg.QueryId, msg.Result, q.QueryType, q.Request))
				return err
			}
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] callback not found for module (%s)", moduleName))
		}
	}
	return nil
}

// verify the query has not exceeded its ttl
func (k Keeper) HasQueryExceededTtl(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, query types.Query) (bool, error) {
	k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] query %s with ttl: %d, resp time: %d.", msg.QueryId, query.Ttl, ctx.BlockHeader().Time.UnixNano()))
	currBlockTime, err := cast.ToUint64E(ctx.BlockTime().UnixNano())
	if err != nil {
		return false, err
	}

	if query.Ttl < currBlockTime {
		errMsg := fmt.Sprintf("[ICQ Resp] aborting query callback due to ttl expiry! ttl is %d, time now %d for query of type %s with id %s, on chain %s",
			query.Ttl, ctx.BlockTime().UnixNano(), query.QueryType, query.ChainId, msg.QueryId)
		fmt.Println(errMsg)
		k.Logger(ctx).Error(errMsg)
		return true, nil
	}
	return false, nil
}

func (k msgServer) SubmitQueryResponse(goCtx context.Context, msg *types.MsgSubmitQueryResponse) (*types.MsgSubmitQueryResponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the response has an associated query stored on stride
	q, found := k.GetQuery(ctx, msg.QueryId)
	if !found {
		k.Logger(ctx).Info("[ICQ Resp] ignoring non-existent query response (note: duplicate responses are nonexistent)")
		return &types.MsgSubmitQueryResponseResponse{}, nil // technically this is an error, but will cause the entire tx to fail if we have one 'bad' message, so we can just no-op here.
	}

	defer ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, q.Id),
		),
		sdk.NewEvent(
			"query_response",
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, q.Id),
			sdk.NewAttribute(types.AttributeKeyChainId, q.ChainId),
		),
	})

	// 1. verify the response's proof, if one exists
	err := k.VerifyKeyProof(ctx, msg, q)
	if err != nil {
		return nil, err
	}
	// 2. immediately delete the query so it cannot process again
	k.DeleteQuery(ctx, q.Id)

	// 3. verify the query's ttl is unexpired
	ttlExceeded, err := k.HasQueryExceededTtl(ctx, msg, q)
	if err != nil {
		return nil, err
	}
	if ttlExceeded {
		k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] %s's ttl exceeded: %d < %d.", msg.QueryId, q.Ttl, ctx.BlockHeader().Time.UnixNano()))
		return &types.MsgSubmitQueryResponseResponse{}, nil
	}

	// 4. if the query is contentless, end
	if len(msg.Result) == 0 {
		k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] query %s is contentless, removing from store.", msg.QueryId))
		return &types.MsgSubmitQueryResponseResponse{}, nil
	}

	// 5. call the query's associated callback function
	err = k.InvokeCallback(ctx, msg, q)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitQueryResponseResponse{}, nil
}
