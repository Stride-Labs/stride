package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (s KeeperTestSuite) TestICAAccountQuery() {
	item := s.createTestICAAccount()
	wctx := sdk.WrapSDKContext(s.Ctx())
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetICAAccountRequest
		response *types.QueryGetICAAccountResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetICAAccountRequest{},
			response: &types.QueryGetICAAccountResponse{ICAAccount: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		response, err := s.App.StakeibcKeeper.ICAAccount(wctx, tc.request)
		if tc.err != nil {
			s.Require().ErrorIs(err, tc.err)
		} else {
			s.Require().NoError(err)
			s.Require().Equal(tc.response, response)
		}
	}
}
