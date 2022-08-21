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

func (k msgServer) SubmitQueryResponse(goCtx context.Context, msg *types.MsgSubmitQueryResponse) (*types.MsgSubmitQueryResponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	q, found := k.GetQuery(ctx, msg.QueryId)

	if found {

		// ABORT PROCESSING QUERY RESPONSE IF WE EXCEEDED THE TTL
		k.Logger(ctx).Info(fmt.Sprintf("[ICQ Resp] query %s with ttl: %d, resp time: %d.", msg.QueryId, q.Ttl, ctx.BlockHeader().Time.UnixNano()))
		curT, err := cast.ToUint64E(ctx.BlockTime().UnixNano())

		if err != nil {
			return nil, err
		}
		if q.Ttl < curT {
			errMsg := fmt.Sprintf("[ICQ Resp] aborting query callback due to ttl expiry! ttl is %d, time now %d for query of type %s with id %s, on chain %s", q.Ttl, ctx.BlockHeader().Time.UnixNano(), q.QueryType, q.ChainId, msg.QueryId)
			k.DeleteQuery(ctx, msg.QueryId)
			k.Logger(ctx).Info(errMsg)
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
		}

		// PROCESS QUERY RESPONSE
		pathParts := strings.Split(q.QueryType, "/")

		// verify the query results are valid by checking the assocaited proof
		if pathParts[len(pathParts)-1] == "key" {
			if msg.ProofOps == nil {
				return nil, fmt.Errorf("unable to validate proof. No proof submitted")
			}
			connection, _ := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, q.ConnectionId)

			msgHeight, err := cast.ToUint64E(msg.Height)
			if err != nil {
				return nil, err
			}
			height := clienttypes.NewHeight(clienttypes.ParseChainID(q.ChainId), msgHeight+1)
			consensusState, found := k.IBCKeeper.ClientKeeper.GetClientConsensusState(ctx, connection.ClientId, height)

			if !found {
				return nil, fmt.Errorf("unable to fetch consensus state")
			}

			clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
			if !found {
				return nil, fmt.Errorf("unable to fetch client state")
			}
			path := commitmenttypes.NewMerklePath([]string{pathParts[1], url.PathEscape(string(q.Request))}...)

			merkleProof, err := commitmenttypes.ConvertProofs(msg.ProofOps)
			if err != nil {
				k.Logger(ctx).Error("error converting proofs")
			}

			tmclientstate, ok := clientState.(*tmclienttypes.ClientState)
			if !ok {
				k.Logger(ctx).Error(fmt.Sprintf("error unmarshaling client state %v", clientState))
			}

			if len(msg.Result) != 0 {
				// if we got a non-nil response, verify inclusion proof.
				if err := merkleProof.VerifyMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path, msg.Result); err != nil {
					return nil, fmt.Errorf("unable to verify proof: %s", err)
				}
				k.Logger(ctx).Info(fmt.Sprintf("Proof validated! module: %s, queryId %s", types.ModuleName, q.Id))

			} else {
				// if we got a nil response, verify non inclusion proof.
				if err := merkleProof.VerifyNonMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path); err != nil {
					return nil, fmt.Errorf("unable to verify proof: %s", err)
				}
				k.Logger(ctx).Info(fmt.Sprintf("Non-inclusion Proof validated! module: %s, queryId %s", types.ModuleName, q.Id))
			}
		}

		noDelete := false
		// execute registered callbacks.

		keys := []string{}
		for k := range k.callbacks {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		k.Logger(ctx).Info(fmt.Sprintf("Executing callbacks for queryId %s", q.Id))
		for _, key := range keys {
			k.Logger(ctx).Info(fmt.Sprintf("Executing callback for module %s", key))
			module := k.callbacks[key]
			if module.Has(q.CallbackId) {
				err := module.Call(ctx, q.CallbackId, msg.Result, q)
				k.Logger(ctx).Info(fmt.Sprintf("Callback %s executed", q.CallbackId))
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("error executing callback %s: %v", q.CallbackId, err))
					// handle edge case; callback has resent the same query!
					// set noDelete to true and short circuit error handling!
					if err == types.ErrSucceededNoDelete {
						noDelete = true
					} else {
						k.Logger(ctx).Error(fmt.Sprintf("error in callback, error: %s, msg: %s, result: %v, type: %s, params: %v", err.Error(), msg.QueryId, msg.Result, q.QueryType, q.Request))
						return nil, err
					}
				}
			} else {
				k.Logger(ctx).Info(fmt.Sprintf("Callback not found for module %s", key))
			}
		}

		if q.Period.IsNegative() {
			if !noDelete {
				k.DeleteQuery(ctx, msg.QueryId)
			}
		} else {
			k.SetQuery(ctx, q)
		}

	} else {
		k.Logger(ctx).Info("Ignoring duplicate query")
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
