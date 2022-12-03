package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestEpochTrackerQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNEpochTracker(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetEpochTrackerRequest
		response *types.QueryGetEpochTrackerResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetEpochTrackerRequest{
				EpochIdentifier: msgs[0].EpochIdentifier,
			},
			response: &types.QueryGetEpochTrackerResponse{EpochTracker: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetEpochTrackerRequest{
				EpochIdentifier: msgs[1].EpochIdentifier,
			},
			response: &types.QueryGetEpochTrackerResponse{EpochTracker: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetEpochTrackerRequest{
				EpochIdentifier: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.EpochTracker(wctx, tc.request)
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

func TestEpochTrackerQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNEpochTracker(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllEpochTrackerRequest {
		return &types.QueryAllEpochTrackerRequest{
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
			resp, err := keeper.EpochTrackerAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.EpochTracker), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.EpochTracker),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.EpochTrackerAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.EpochTracker), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.EpochTracker),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.EpochTrackerAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.EpochTracker),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.EpochTrackerAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
