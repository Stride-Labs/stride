package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
	"github.com/Stride-Labs/stride/v5/testutil/nullify"
	keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
)

func TestAuctionPoolQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.AuctionKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNAuctionPool(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetAuctionPoolRequest
		response *types.QueryGetAuctionPoolResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetAuctionPoolRequest{Id: msgs[0].Id},
			response: &types.QueryGetAuctionPoolResponse{AuctionPool: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetAuctionPoolRequest{Id: msgs[1].Id},
			response: &types.QueryGetAuctionPoolResponse{AuctionPool: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetAuctionPoolRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.AuctionPool(wctx, tc.request)
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

func TestAuctionPoolQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.AuctionKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNAuctionPool(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllAuctionPoolRequest {
		return &types.QueryAllAuctionPoolRequest{
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
			resp, err := keeper.AuctionPoolAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.AuctionPool), step)
			require.Subset(t,
				nullify.Fill(msgs),
            	nullify.Fill(resp.AuctionPool),
            )
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.AuctionPoolAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.AuctionPool), step)
			require.Subset(t,
				nullify.Fill(msgs),
            	nullify.Fill(resp.AuctionPool),
            )
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.AuctionPoolAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.AuctionPool),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.AuctionPoolAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
