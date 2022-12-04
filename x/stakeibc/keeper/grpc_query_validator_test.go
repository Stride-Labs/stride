package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func TestValidatorQuery(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	validatorsByHostZone := make(map[string][]*types.Validator)
	validators := []*types.Validator{}
	nullify.Fill(&validators)

	chainId := "GAIA"
	hostZone := &types.HostZone{
		ChainId:    chainId,
		Validators: validators,
	}
	nullify.Fill(&hostZone)
	validatorsByHostZone[chainId] = validators
	keeper.SetHostZone(ctx, *hostZone)

	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetValidatorsRequest
		response *types.QueryGetValidatorsResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetValidatorsRequest{ChainId: chainId},
			response: &types.QueryGetValidatorsResponse{Validators: validators},
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
