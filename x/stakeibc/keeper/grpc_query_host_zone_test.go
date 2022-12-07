package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func TestHostZoneQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNHostZone(keeper, ctx, 2)
	for _, msg := range msgs {
		t.Log(msg.ChainId)
	}
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetHostZoneRequest
		response *types.QueryGetHostZoneResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetHostZoneRequest{ChainId: msgs[0].ChainId},
			response: &types.QueryGetHostZoneResponse{HostZone: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetHostZoneRequest{ChainId: msgs[1].ChainId},
			response: &types.QueryGetHostZoneResponse{HostZone: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetHostZoneRequest{ChainId: strconv.Itoa((len(msgs)))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.HostZone(wctx, tc.request)
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

func TestHostZoneQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNHostZone(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllHostZoneRequest {
		return &types.QueryAllHostZoneRequest{
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
			resp, err := keeper.HostZoneAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.HostZone), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.HostZone),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.HostZoneAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.HostZone), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.HostZone),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.HostZoneAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.HostZone),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.HostZoneAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
