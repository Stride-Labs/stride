package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/interchainquery/types"
)

var _ types.QueryServiceServer = Keeper{}

// Queries all queries that have been requested but have not received a response
func (k Keeper) PendingQueries(c context.Context, req *types.QueryPendingQueriesRequest) (*types.QueryPendingQueriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	pendingQueries := []types.Query{}
	for _, query := range k.AllQueries(ctx) {
		if query.RequestSent {
			pendingQueries = append(pendingQueries, query)
		}
	}

	return &types.QueryPendingQueriesResponse{PendingQueries: pendingQueries}, nil
}
