package keeper

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	tmclienttypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
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
			return fmt.Errorf("unable to validate proof. No proof submitted")
		}
		connection, _ := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, q.ConnectionId)

		msgHeight, err := cast.ToUint64E(msg.Height)
		if err != nil {
			return err
		}
		height := clienttypes.NewHeight(clienttypes.ParseChainID(q.ChainId), msgHeight+1)
		consensusState, found := k.IBCKeeper.ClientKeeper.GetClientConsensusState(ctx, connection.ClientId, height)
		if !found {
			return fmt.Errorf("unable to fetch consensus state")
		}

		clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
		if !found {
			return fmt.Errorf("unable to fetch client state")
		}
		path := commitmenttypes.NewMerklePath([]string{pathParts[1], url.PathEscape(string(q.Request))}...)

		k.Logger(ctx).Info(fmt.Sprintf("[MOOSE] proof is %#v", msg.ProofOps))
		merkleProof, err := commitmenttypes.ConvertProofs(msg.ProofOps)
		if err != nil {
			return fmt.Errorf("error converting proofs")
		}

		tmclientstate, ok := clientState.(*tmclienttypes.ClientState)
		if !ok {
			return fmt.Errorf("error unmarshaling client state %v", clientState)
		}

		if len(msg.Result) != 0 {
			// if we got a non-nil response, verify inclusion proof.
			if err := merkleProof.VerifyMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path, msg.Result); err != nil {
				return fmt.Errorf("unable to verify proof: %s", err)
			}
			k.Logger(ctx).Info(fmt.Sprintf("Proof validated! module: %s, queryId %s", types.ModuleName, q.Id))

		} else {
			// if we got a nil response, verify non inclusion proof.
			if err := merkleProof.VerifyNonMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path); err != nil {
				return fmt.Errorf("unable to verify proof: %s", err)
			}
			k.Logger(ctx).Info(fmt.Sprintf("Non-inclusion Proof validated! module: %s, queryId %s", types.ModuleName, q.Id))
		}
	}
	return nil
}

// call the query's associated callback function
func (k Keeper) FindAndInvokeCallback(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, q types.Query) error {
	// get all the stored queries and sort them for determinism
	moduleNames := []string{}
	for moduleName := range k.callbacks {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)

	for _, module := range moduleNames {
		k.Logger(ctx).Info(fmt.Sprintf("Executing callback for queryId (%s), module (%s)", q.Id, module))
		moduleCallbackHandler := k.callbacks[module]

		if moduleCallbackHandler.Has(q.CallbackId) {
			k.Logger(ctx).Info(fmt.Sprintf("ICQ Callback (%s) found for module (%s)", q.CallbackId, module))
			// call the correct callback function
			err := moduleCallbackHandler.Call(ctx, q.CallbackId, msg.Result, q)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("error in ICQ callback, error: %s, msg: %s, result: %v, type: %s, params: %v", err.Error(), msg.QueryId, msg.Result, q.QueryType, q.Request))
				return err
			}
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("ICQ Callback not found for module (%s)", module))
		}
	}
	return nil
}

// verify the query has not exceeded its ttl
func (k Keeper) HasQueryExceededTtl(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, q types.Query) (bool, error) {
	k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] query %s with ttl: %d, resp time: %d.", msg.QueryId, q.Ttl, ctx.BlockHeader().Time.UnixNano()))
	curT, err := cast.ToUint64E(ctx.BlockTime().UnixNano())
	if err != nil {
		return false, err
	}

	if q.Ttl < curT {
		errMsg := fmt.Sprintf("[ICQ Resp] aborting query callback due to ttl expiry! ttl is %d, time now %d for query of type %s with id %s, on chain %s",
			q.Ttl, ctx.BlockHeader().Time.UnixNano(), q.QueryType, q.ChainId, msg.QueryId)
		k.DeleteQuery(ctx, msg.QueryId)
		k.Logger(ctx).Error(errMsg)
		return true, nil
	}
	return false, nil
}

func (k msgServer) SubmitQueryResponse(goCtx context.Context, msg *types.MsgSubmitQueryResponse) (*types.MsgSubmitQueryResponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 1. check if the response has an associated query stored on stride
	q, found := k.GetQuery(ctx, msg.QueryId)
	if found {
		// 2. verify the query's ttl is unexpired
		exceeded, err := k.HasQueryExceededTtl(ctx, msg, q)
		if err != nil {
			return nil, err
		}
		if exceeded {
			return &types.MsgSubmitQueryResponseResponse{}, nil
		}

		// 3. verify the response's proof, if one exists
		err = k.VerifyKeyProof(ctx, msg, q)
		if err != nil {
			return nil, err
		}

		// 4. call the query's associated callback function
		err = k.FindAndInvokeCallback(ctx, msg, q)
		if err != nil {
			return nil, err
		}

		// 5. delete the query from the store once it's been processed
		k.DeleteQuery(ctx, msg.QueryId)

	} else {
		k.Logger(ctx).Info("Ignoring duplicate or nonexistent query")
		return &types.MsgSubmitQueryResponseResponse{}, nil // technically this is an error, but will cause the entire tx to fail if we have one 'bad' message, so we can just no-op here.
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &types.MsgSubmitQueryResponseResponse{}, nil
}
