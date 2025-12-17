package keeper_test

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v31/x/icacallbacks/types"
)

func (s *KeeperTestSuite) TestCallbackDataQuerySingle() {
	msgs := s.createNCallbackData(s.Ctx, 2)
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
		s.Run(tc.desc, func() {
			response, err := s.App.IcacallbacksKeeper.CallbackData(s.Ctx, tc.request)
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

func (s *KeeperTestSuite) TestCallbackDataQueryPaginated() {
	msgs := s.createNCallbackData(s.Ctx, 5)

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
	s.Run("ByOffset", func() {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.IcacallbacksKeeper.CallbackDataAll(s.Ctx, request(nil, uint64(i), uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.CallbackData), step)
			s.Require().Subset(
				msgs,
				resp.CallbackData,
			)
		}
	})
	s.Run("ByKey", func() {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := s.App.IcacallbacksKeeper.CallbackDataAll(s.Ctx, request(next, 0, uint64(step), false))
			s.Require().NoError(err)
			s.Require().LessOrEqual(len(resp.CallbackData), step)
			s.Require().Subset(
				msgs,
				resp.CallbackData,
			)
			next = resp.Pagination.NextKey
		}
	})
	s.Run("Total", func() {
		resp, err := s.App.IcacallbacksKeeper.CallbackDataAll(s.Ctx, request(nil, 0, 0, true))
		s.Require().NoError(err)
		s.Require().Equal(len(msgs), int(resp.Pagination.Total))
		s.Require().ElementsMatch(
			msgs,
			resp.CallbackData,
		)
	})
	s.Run("InvalidRequest", func() {
		_, err := s.App.IcacallbacksKeeper.CallbackDataAll(s.Ctx, nil)
		s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
