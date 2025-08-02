package keeper_test

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v28/x/records/types"
)

func (s *KeeperTestSuite) TestEpochUnbondingRecordQuerySingle() {
	msgs, _ := s.createNEpochUnbondingRecord(2)
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
		s.Run(tc.desc, func() {
			response, err := s.App.RecordsKeeper.EpochUnbondingRecord(s.Ctx, tc.request)
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

func (s *KeeperTestSuite) TestEpochUnbondingRecordQueryPaginated() {
	msgs, _ := s.createNEpochUnbondingRecord(5)

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
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.RecordsKeeper.EpochUnbondingRecordAll(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.EpochUnbondingRecord), step)
			s.Require().Subset(
				msgs,
				resp.EpochUnbondingRecord,
			)
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.RecordsKeeper.EpochUnbondingRecordAll(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.EpochUnbondingRecord), step)
			s.Require().Subset(
				msgs,
				resp.EpochUnbondingRecord,
			)
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.App.RecordsKeeper.EpochUnbondingRecordAll(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(msgs), int(resp.Pagination.Total))
		s.Require().ElementsMatch(
			msgs,
			resp.EpochUnbondingRecord,
		)
	})
	s.Run("InvalidRequest", func() {
		_, err := s.App.RecordsKeeper.EpochUnbondingRecordAll(s.Ctx, nil)
		s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
