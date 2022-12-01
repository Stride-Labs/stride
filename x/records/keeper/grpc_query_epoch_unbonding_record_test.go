package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/records/types"
)

func TestEpochUnbondingRecordQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs, _ := createNEpochUnbondingRecord(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetEpochUnbondingRecordRequest
		response *types.QueryGetEpochUnbondingRecordResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetEpochUnbondingRecordRequest{EpochNumber: msgs[0].EpochNumber},
			response: &types.QueryGetEpochUnbondingRecordResponse{EpochUnbondingRecord: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetEpochUnbondingRecordRequest{EpochNumber: msgs[1].EpochNumber},
			response: &types.QueryGetEpochUnbondingRecordResponse{EpochUnbondingRecord: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetEpochUnbondingRecordRequest{EpochNumber: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.EpochUnbondingRecord(wctx, tc.request)
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

func TestEpochUnbondingRecordQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs, _ := createNEpochUnbondingRecord(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllEpochUnbondingRecordRequest {
		return &types.QueryAllEpochUnbondingRecordRequest{
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
			resp, err := keeper.EpochUnbondingRecordAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.EpochUnbondingRecord), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.EpochUnbondingRecord),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.EpochUnbondingRecordAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.EpochUnbondingRecord), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.EpochUnbondingRecord),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.EpochUnbondingRecordAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.EpochUnbondingRecord),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.EpochUnbondingRecordAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
