package ratelimit_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/ratelimit"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := 1; i <= 3; i++ {
		suffix := strconv.Itoa(i)
		rateLimit := types.RateLimit{
			Path:  &types.Path{Denom: "denom-" + suffix, ChannelId: "channel-" + suffix},
			Quota: &types.Quota{MaxPercentSend: uint64(i), MaxPercentRecv: uint64(i), DurationHours: uint64(i)},
			Flow:  &types.Flow{Inflow: uint64(i), Outflow: uint64(i), ChannelValue: uint64(i)},
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
