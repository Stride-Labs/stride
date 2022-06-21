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
	"github.com/ingenuity-build/quicksilver/x/interchainquery/types"
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
	//if found && q.LastHeight.Int64() != ctx.BlockHeader().Height {
	if found {
		pathParts := strings.Split(q.QueryType, "/")
		if pathParts[len(pathParts)-1] == "key" {
			if msg.ProofOps == nil {
				return nil, fmt.Errorf("unable to validate proof. No proof submitted")
			}
			connection, _ := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, q.ConnectionId)

			height := clienttypes.NewHeight(clienttypes.ParseChainID(q.ChainId), uint64(msg.Height)+1)
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
				k.Logger(ctx).Error("error unmarshaling client state", "cs", clientState)
			}

			if len(msg.Result) != 0 {
				// if we got a non-nil response, verify inclusion proof.
				if err := merkleProof.VerifyMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path, msg.Result); err != nil {
					return nil, fmt.Errorf("unable to verify proof: %s", err)
				}
				k.Logger(ctx).Debug("Proof validated!", "module", types.ModuleName, "queryId", q.Id)

			} else {
				// if we got a nil response, verify non inclusion proof.
				if err := merkleProof.VerifyNonMembership(tmclientstate.ProofSpecs, consensusState.GetRoot(), path); err != nil {
					return nil, fmt.Errorf("unable to verify proof: %s", err)
				}
				k.Logger(ctx).Debug("Non-inclusion Proof validated!", "module", types.ModuleName, "queryId", q.Id)
			}
		}

		noDelete := false
		// execute registered callbacks.

		keys := []string{}
		for k := range k.callbacks {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, key := range keys {
			module := k.callbacks[key]
			if module.Has(q.CallbackId) {
				err := module.Call(ctx, q.CallbackId, msg.Result, q)
				if err != nil {
					// handle edge case; callback has resent the same query!
					// set noDelete to true and short circuit error handling!
					if err == types.ErrSucceededNoDelete {
						noDelete = true
					} else {
						k.Logger(ctx).Error("error in callback", "error", err, "msg", msg.QueryId, "result", msg.Result, "type", q.QueryType, "params", q.Request)
						return nil, err
					}
				}
			}
		}

		if q.Ttl > 0 {
			// don't store if ttl is 0
			if err := k.SetDatapointForId(ctx, msg.QueryId, msg.Result, sdk.NewInt(msg.Height)); err != nil {
				return nil, err
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
