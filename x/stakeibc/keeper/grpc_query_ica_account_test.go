package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (suite *KeeperTestSuite) TestICAAccountQuery() {
	item := suite.createTestICAAccount()
	wctx := sdk.WrapSDKContext(suite.Ctx)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetICAAccountRequest
		response *types.QueryGetICAAccountResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetICAAccountRequest{},
			response: &types.QueryGetICAAccountResponse{IcaAccount: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		response, err := suite.App.StakeibcKeeper.ICAAccount(wctx, tc.request)
		if tc.err != nil {
			suite.Require().ErrorIs(err, tc.err)
		} else {
			suite.Require().NoError(err)
			suite.Require().Equal(tc.response, response)
		}
	}
}
