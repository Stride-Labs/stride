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

	keepertest "github.com/Stride-Labs/stride/v27/testutil/keeper"
	"github.com/Stride-Labs/stride/v27/testutil/nullify"
	"github.com/Stride-Labs/stride/v27/x/records/types"
)

func TestUserRedemptionRecordQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNUserRedemptionRecord(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetUserRedemptionRecordRequest
		response *types.QueryGetUserRedemptionRecordResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetUserRedemptionRecordRequest{Id: msgs[0].Id},
			response: &types.QueryGetUserRedemptionRecordResponse{UserRedemptionRecord: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetUserRedemptionRecordRequest{Id: msgs[1].Id},
			response: &types.QueryGetUserRedemptionRecordResponse{UserRedemptionRecord: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetUserRedemptionRecordRequest{Id: strconv.Itoa(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.UserRedemptionRecord(wctx, tc.request)
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

func TestUserRedemptionRecordQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNUserRedemptionRecord(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllUserRedemptionRecordRequest {
		return &types.QueryAllUserRedemptionRecordRequest{
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
			resp, err := keeper.UserRedemptionRecordAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.UserRedemptionRecord), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.UserRedemptionRecord),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.UserRedemptionRecordAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.UserRedemptionRecord), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.UserRedemptionRecord),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.UserRedemptionRecordAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.UserRedemptionRecord),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.UserRedemptionRecordAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
