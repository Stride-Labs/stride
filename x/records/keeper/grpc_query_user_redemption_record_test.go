package keeper_test

import (
	"strconv"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v31/x/records/types"
)

func (s *KeeperTestSuite) TestUserRedemptionRecordQuerySingle() {
	msgs := s.createNUserRedemptionRecord(2)
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
		s.Run(tc.desc, func() {
			response, err := s.App.RecordsKeeper.UserRedemptionRecord(s.Ctx, tc.request)
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

func (s *KeeperTestSuite) TestUserRedemptionRecordQueryPaginated() {
	msgs := s.createNUserRedemptionRecord(5)

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
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.RecordsKeeper.UserRedemptionRecordAll(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.UserRedemptionRecord), step)
			s.Require().Subset(
				msgs,
				resp.UserRedemptionRecord,
			)
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.RecordsKeeper.UserRedemptionRecordAll(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.UserRedemptionRecord), step)
			s.Require().Subset(
				msgs,
				resp.UserRedemptionRecord,
			)
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.App.RecordsKeeper.UserRedemptionRecordAll(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(msgs), int(resp.Pagination.Total))
		s.Require().ElementsMatch(
			msgs,
			resp.UserRedemptionRecord,
		)
	})
	s.Run("InvalidRequest", func() {
		_, err := s.App.RecordsKeeper.UserRedemptionRecordAll(s.Ctx, nil)
		s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
