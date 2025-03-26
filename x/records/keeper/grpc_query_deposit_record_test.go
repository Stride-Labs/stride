package keeper_test

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v26/x/records/types"
)

func (s *KeeperTestSuite) createNDepositRecord(n int) []types.DepositRecord {
	items := make([]types.DepositRecord, n)
	for i := range items {
		items[i].Id = uint64(i)
		items[i].Amount = sdkmath.NewInt(int64(i))
		s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, items[i])
	}
	return items
}

func (s *KeeperTestSuite) TestDepositRecordQuerySingle() {
	msgs := s.createNDepositRecord(2)
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
		s.Run(tc.desc, func() {
			response, err := s.App.RecordsKeeper.DepositRecord(s.Ctx, tc.request)
			if tc.err != nil {
				s.Require().ErrorIs(err, tc.err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(
					tc.response,
					response,
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDepositRecordQueryPaginated() {
	msgs := s.createNDepositRecord(5)

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
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.RecordsKeeper.DepositRecordAll(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.DepositRecord), step)
			s.Require().Subset(
				msgs,
				resp.DepositRecord,
			)
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.RecordsKeeper.DepositRecordAll(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.DepositRecord), step)
			s.Require().Subset(
				msgs,
				resp.DepositRecord,
			)
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.App.RecordsKeeper.DepositRecordAll(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(msgs), int(resp.Pagination.Total))
		s.Require().ElementsMatch(
			msgs,
			resp.DepositRecord,
		)
	})
	s.Run("InvalidRequest", func() {
		_, err := s.App.RecordsKeeper.DepositRecordAll(s.Ctx, nil)
		s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func (s *KeeperTestSuite) TestQueryDepositRecordByHost() {
	// Store deposit records across two hosts
	hostChain1 := "chain-1"
	hostChain2 := "chain-2"

	hostDepositRecords1 := []types.DepositRecord{
		{HostZoneId: hostChain1, Id: 1, Amount: sdkmath.NewInt(1)},
		{HostZoneId: hostChain1, Id: 2, Amount: sdkmath.NewInt(2)},
		{HostZoneId: hostChain1, Id: 3, Amount: sdkmath.NewInt(3)},
	}
	hostDepositRecords2 := []types.DepositRecord{
		{HostZoneId: hostChain2, Id: 4, Amount: sdkmath.NewInt(4)},
		{HostZoneId: hostChain2, Id: 5, Amount: sdkmath.NewInt(5)},
		{HostZoneId: hostChain2, Id: 6, Amount: sdkmath.NewInt(6)},
	}

	for _, depositRecord := range append(hostDepositRecords1, hostDepositRecords2...) {
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)
	}

	// Fetch each list through a host zone id query
	actualHostDepositRecords1, err := s.App.RecordsKeeper.DepositRecordByHost(s.Ctx, &types.QueryDepositRecordByHostRequest{
		HostZoneId: hostChain1,
	})
	s.Require().NoError(err, "no error expected when querying by host %s", hostChain1)
	s.Require().ElementsMatch(hostDepositRecords1, actualHostDepositRecords1.DepositRecord, "deposit records for %s", hostChain1)

	actualHostDepositRecords2, err := s.App.RecordsKeeper.DepositRecordByHost(s.Ctx, &types.QueryDepositRecordByHostRequest{
		HostZoneId: hostChain2,
	})
	s.Require().NoError(err, "no error expected when querying by host %s", hostChain2)
	s.Require().ElementsMatch(hostDepositRecords2, actualHostDepositRecords2.DepositRecord, "deposit records for %s", hostChain2)

	// Finally, fetch a non-existent chain-id and it should return an empty list
	fakeHostDepositRecords, err := s.App.RecordsKeeper.DepositRecordByHost(s.Ctx, &types.QueryDepositRecordByHostRequest{
		HostZoneId: "fake_host",
	})
	s.Require().NoError(err, "no error expected when querying by host %s", hostChain1)
	s.Require().Len(fakeHostDepositRecords.DepositRecord, 0)
}
