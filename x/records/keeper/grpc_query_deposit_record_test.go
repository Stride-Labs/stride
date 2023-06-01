package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v9/testutil/keeper"
	"github.com/Stride-Labs/stride/v9/testutil/nullify"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v9/x/records/keeper"
	"github.com/Stride-Labs/stride/v9/x/records/types"
)

func createNDepositRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.DepositRecord {
	items := make([]types.DepositRecord, n)
	for i := range items {
		items[i].Id = uint64(i)
		items[i].Amount = sdkmath.NewInt(int64(i))
		keeper.AppendDepositRecord(ctx, items[i])
	}
	return items
}

func TestDepositRecordQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNDepositRecord(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetDepositRecordRequest
		response *types.QueryGetDepositRecordResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetDepositRecordRequest{Id: msgs[0].Id},
			response: &types.QueryGetDepositRecordResponse{DepositRecord: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetDepositRecordRequest{Id: msgs[1].Id},
			response: &types.QueryGetDepositRecordResponse{DepositRecord: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetDepositRecordRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.DepositRecord(wctx, tc.request)
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

func TestDepositRecordQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNDepositRecord(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllDepositRecordRequest {
		return &types.QueryAllDepositRecordRequest{
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
			resp, err := keeper.DepositRecordAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.DepositRecord), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.DepositRecord),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.DepositRecordAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.DepositRecord), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.DepositRecord),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.DepositRecordAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.DepositRecord),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.DepositRecordAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func (s *KeeperTestSuite) TestQueryDepositRecordByHost() {
	wctx := sdk.WrapSDKContext(s.Ctx)

	// Store deposit records across two hosts
	hostChain1 := "chain-1"
	hostChain2 := "chain-2"

	hostDepositRecords1 := []types.DepositRecord{
		{HostZoneId: hostChain1, Id: 1, Amount: sdk.NewInt(1)},
		{HostZoneId: hostChain1, Id: 2, Amount: sdk.NewInt(2)},
		{HostZoneId: hostChain1, Id: 3, Amount: sdk.NewInt(3)},
	}
	hostDepositRecords2 := []types.DepositRecord{
		{HostZoneId: hostChain2, Id: 4, Amount: sdk.NewInt(4)},
		{HostZoneId: hostChain2, Id: 5, Amount: sdk.NewInt(5)},
		{HostZoneId: hostChain2, Id: 6, Amount: sdk.NewInt(6)},
	}

	for _, depositRecord := range append(hostDepositRecords1, hostDepositRecords2...) {
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)
	}

	// Fetch each list through a host zone id query
	actualHostDepositRecords1, err := s.App.RecordsKeeper.DepositRecordByHost(wctx, &types.QueryDepositRecordByHostRequest{
		HostZoneId: hostChain1,
	})
	s.Require().NoError(err, "no error expected when querying by host %s", hostChain1)
	s.Require().ElementsMatch(hostDepositRecords1, actualHostDepositRecords1.DepositRecord, "deposit records for %s", hostChain1)

	actualHostDepositRecords2, err := s.App.RecordsKeeper.DepositRecordByHost(wctx, &types.QueryDepositRecordByHostRequest{
		HostZoneId: hostChain2,
	})
	s.Require().NoError(err, "no error expected when querying by host %s", hostChain2)
	s.Require().ElementsMatch(hostDepositRecords2, actualHostDepositRecords2.DepositRecord, "deposit records for %s", hostChain2)

	// Finally, fetch a non-existent chain-id and it should return an empty list
	fakeHostDepositRecords, err := s.App.RecordsKeeper.DepositRecordByHost(wctx, &types.QueryDepositRecordByHostRequest{
		HostZoneId: "fake_host",
	})
	s.Require().NoError(err, "no error expected when querying by host %s", hostChain1)
	s.Require().Len(fakeHostDepositRecords.DepositRecord, 0)
}
