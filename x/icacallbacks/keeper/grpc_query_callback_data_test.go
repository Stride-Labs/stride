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
	"github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestCallbackDataQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.IcacallbacksKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNCallbackData(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetCallbackDataRequest
		response *types.QueryGetCallbackDataResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetCallbackDataRequest{
				CallbackKey: msgs[0].CallbackKey,
			},
			response: &types.QueryGetCallbackDataResponse{CallbackData: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetCallbackDataRequest{
				CallbackKey: msgs[1].CallbackKey,
			},
			response: &types.QueryGetCallbackDataResponse{CallbackData: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetCallbackDataRequest{
				CallbackKey: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.CallbackData(wctx, tc.request)
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

func TestCallbackDataQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.IcacallbacksKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNCallbackData(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllCallbackDataRequest {
		return &types.QueryAllCallbackDataRequest{
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
			resp, err := keeper.CallbackDataAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.CallbackData), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.CallbackData),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.CallbackDataAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.CallbackData), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.CallbackData),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.CallbackDataAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.CallbackData),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.CallbackDataAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
