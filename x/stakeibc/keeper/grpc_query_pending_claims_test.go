package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestPendingClaimsQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNPendingClaims(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetPendingClaimsRequest
		response *types.QueryGetPendingClaimsResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetPendingClaimsRequest{
				Sequence: msgs[0].Sequence,
			},
			response: &types.QueryGetPendingClaimsResponse{PendingClaims: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetPendingClaimsRequest{
				Sequence: msgs[1].Sequence,
			},
			response: &types.QueryGetPendingClaimsResponse{PendingClaims: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetPendingClaimsRequest{
				Sequence: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.PendingClaims(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestPendingClaimsQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNPendingClaims(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPendingClaimsRequest {
		return &types.QueryAllPendingClaimsRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.PendingClaimsAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PendingClaims), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PendingClaims),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.PendingClaimsAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PendingClaims), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PendingClaims),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.PendingClaimsAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.PendingClaims),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.PendingClaimsAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
