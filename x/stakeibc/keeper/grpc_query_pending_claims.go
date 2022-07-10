package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) PendingClaimsAll(c context.Context, req *types.QueryAllPendingClaimsRequest) (*types.QueryAllPendingClaimsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var pendingClaimss []types.PendingClaims
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	pendingClaimsStore := prefix.NewStore(store, types.KeyPrefix(types.PendingClaimsKeyPrefix))

	pageRes, err := query.Paginate(pendingClaimsStore, req.Pagination, func(key []byte, value []byte) error {
		var pendingClaims types.PendingClaims
		if err := k.cdc.Unmarshal(value, &pendingClaims); err != nil {
			return err
		}

		pendingClaimss = append(pendingClaimss, pendingClaims)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPendingClaimsResponse{PendingClaims: pendingClaimss, Pagination: pageRes}, nil
}

func (k Keeper) PendingClaims(c context.Context, req *types.QueryGetPendingClaimsRequest) (*types.QueryGetPendingClaimsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetPendingClaims(
	    ctx,
	    req.Sequence,
        )
	if !found {
	    return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPendingClaimsResponse{PendingClaims: val}, nil
}