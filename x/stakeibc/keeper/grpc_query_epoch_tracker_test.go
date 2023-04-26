package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v9/testutil/keeper"
	"github.com/Stride-Labs/stride/v9/testutil/nullify"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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

func TestAllEpochTrackerQuery(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNEpochTracker(keeper, ctx, 5)

	resp, err := keeper.EpochTrackerAll(wctx, &types.QueryAllEpochTrackerRequest{})
	require.NoError(t, err)
	require.Len(t, resp.EpochTracker, 5)
	require.Subset(t,
		nullify.Fill(msgs),
		nullify.Fill(resp.EpochTracker),
	)
}
