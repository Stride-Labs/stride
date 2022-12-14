package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func TestCheckExceedsQuota(t *testing.T) {
	totalValue := uint64(100)
	amountUnderThreshold := uint64(5)
	amountOverThreshold := uint64(15)
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
			name:       "inflow exceeded threshold",
			direction:  types.PACKET_RECV,
			amount:     amountOverThreshold,
			totalValue: totalValue,
			exp:        true,
		},
		{
			name:       "inflow did not exceed threshold",
			direction:  types.PACKET_RECV,
			amount:     amountUnderThreshold,
			totalValue: totalValue,
			exp:        false,
		},
		{
			name:       "outflow exceeded threshold",
			direction:  types.PACKET_SEND,
			amount:     amountOverThreshold,
			totalValue: totalValue,
			exp:        true,
		},
		{
			name:       "outflow did not exceed threshold",
			direction:  types.PACKET_SEND,
			amount:     amountUnderThreshold,
			totalValue: totalValue,
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
