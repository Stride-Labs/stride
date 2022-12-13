package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func TestCheckExceedsQuota(t *testing.T) {
	quota := types.Quota{
		MaxPercentRecv: uint64(10),
		MaxPercentSend: uint64(10),
		DurationHours:  uint64(1),
	}

	tests := []struct {
		name       string
		direction  types.PacketDirection
		amount     uint64
		totalValue uint64
		exp        bool
	}{
		{
			name:       "inflow exceed",
			direction:  types.PACKET_RECV,
			amount:     200*quota.MaxPercentRecv/100 + 1,
			totalValue: 200,
			exp:        true,
		},
		{
			name:       "inflow not exceed",
			direction:  types.PACKET_RECV,
			amount:     200 * quota.MaxPercentRecv / 100,
			totalValue: 200,
			exp:        false,
		},
		{
			name:       "outflow exceed",
			direction:  types.PACKET_RECV,
			amount:     200*quota.MaxPercentSend/100 + 1,
			totalValue: 200,
			exp:        true,
		},
		{
			name:       "outflow not exceed",
			direction:  types.PACKET_RECV,
			amount:     200 * quota.MaxPercentSend / 100,
			totalValue: 200,
			exp:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := quota.CheckExceedsQuota(test.direction, test.amount, test.totalValue)
			require.Equal(t, res, test.exp)
		})
	}
}
