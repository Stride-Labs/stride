package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-labs/stride/testutil/keeper"
	"github.com/Stride-labs/stride/testutil/nullify"
	"github.com/Stride-labs/stride/x/stakeibc/types"
)

func TestMinValidatorRequirementsQuery(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	item := createTestMinValidatorRequirements(keeper, ctx)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetMinValidatorRequirementsRequest
		response *types.QueryGetMinValidatorRequirementsResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetMinValidatorRequirementsRequest{},
			response: &types.QueryGetMinValidatorRequirementsResponse{MinValidatorRequirements: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.MinValidatorRequirements(wctx, tc.request)
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
