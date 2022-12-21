package ratelimit_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/ratelimit"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := int64(1); i <= 3; i++ {
		suffix := strconv.Itoa(int(i))
		rateLimit := types.RateLimit{
			Path:  &types.Path{Denom: "denom-" + suffix, ChannelId: "channel-" + suffix},
			Quota: &types.Quota{MaxPercentSend: sdk.NewInt(i), MaxPercentRecv: sdk.NewInt(i), DurationHours: uint64(i)},
			Flow:  &types.Flow{Inflow: sdk.NewInt(i), Outflow: sdk.NewInt(i), ChannelValue: sdk.NewInt(i)},
		}

		rateLimits = append(rateLimits, rateLimit)
	}
	return rateLimits
}

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:     types.Params{},
		RateLimits: createRateLimits(),
	}

	s := apptesting.SetupSuitelessTestHelper()
	ratelimit.InitGenesis(s.Ctx, s.App.RatelimitKeeper, genesisState)
	got := ratelimit.ExportGenesis(s.Ctx, s.App.RatelimitKeeper)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.RateLimits, got.RateLimits)
}
