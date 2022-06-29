package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func TestValidatorQuery(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	item := []*types.Validator{
		{Name: "A", Address: "A", Status: types.Validator_Active},
	}
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetValidatorsRequest
		response *types.QueryGetValidatorsResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetValidatorsRequest{},
			response: &types.QueryGetValidatorsResponse{Validators: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Validators(wctx, tc.request)
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
