package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func TestAddInflow(t *testing.T) {
	totalValue := uint64(100)
	quota := types.Quota{
		MaxPercentRecv: uint64(10),
		MaxPercentSend: uint64(10),
		DurationHours:  uint64(1),
	}

	tests := []struct {
		name         string
		flow         types.Flow
		expectedFlow types.Flow
		amount       uint64
		succeeds     bool
	}{
		{
			name: "AddInflow__Successful__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       0,
				Outflow:      0,
				ChannelValue: totalValue,
			},
			amount: 5,
			expectedFlow: types.Flow{
				Inflow:       5,
				Outflow:      0,
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddInflow__Successful__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       100,
				Outflow:      100,
				ChannelValue: totalValue,
			},
			amount: 5,
			expectedFlow: types.Flow{
				Inflow:       105,
				Outflow:      100,
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddInflow__Failure__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       0,
				Outflow:      0,
				ChannelValue: totalValue,
			},
			amount:   15,
			succeeds: false,
		},
		{
			name: "AddInflow__Failure__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       100,
				Outflow:      100,
				ChannelValue: totalValue,
			},
			amount:   15,
			succeeds: false,
		},
		{
			name: "AddInflow__Successful__Large amount but net outflow",
			flow: types.Flow{
				Inflow:       1,
				Outflow:      10,
				ChannelValue: totalValue,
			},
			amount: 15,
			expectedFlow: types.Flow{
				Inflow:       16,
				Outflow:      10,
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddInflow__Failure__Small amount but net inflow",
			flow: types.Flow{
				Inflow:       10,
				Outflow:      1,
				ChannelValue: totalValue,
			},
			amount:   5,
			succeeds: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initialFlow := test.flow
			err := test.flow.AddInflow(test.amount, quota)
			actualFlow := test.flow

			if test.succeeds {
				require.NoError(t, err)
				require.Equal(t, test.expectedFlow, actualFlow)
			} else {
				require.ErrorContains(t, err, "Inflow exceeds quota", "test: %v", test.name)
				require.Equal(t, initialFlow, actualFlow)
			}
		})
	}
}

func TestOutInflow(t *testing.T) {
	totalValue := uint64(100)
	quota := types.Quota{
		MaxPercentRecv: uint64(10),
		MaxPercentSend: uint64(10),
		DurationHours:  uint64(1),
	}

	tests := []struct {
		name         string
		flow         types.Flow
		expectedFlow types.Flow
		amount       uint64
		succeeds     bool
	}{
		{
			name: "AddOutflow__Successful__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       0,
				Outflow:      0,
				ChannelValue: totalValue,
			},
			amount: 5,
			expectedFlow: types.Flow{
				Inflow:       0,
				Outflow:      5,
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddOutflow__Successful__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       100,
				Outflow:      100,
				ChannelValue: totalValue,
			},
			amount: 5,
			expectedFlow: types.Flow{
				Inflow:       100,
				Outflow:      105,
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddOutflow__Failure__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       0,
				Outflow:      0,
				ChannelValue: totalValue,
			},
			amount:   15,
			succeeds: false,
		},
		{
			name: "AddOutflow__Failure__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       100,
				Outflow:      100,
				ChannelValue: totalValue,
			},
			amount:   15,
			succeeds: false,
		},
		{
			name: "AddOutflow__Succeesful__Large amount but net inflow",
			flow: types.Flow{
				Inflow:       10,
				Outflow:      1,
				ChannelValue: totalValue,
			},
			amount: 15,
			expectedFlow: types.Flow{
				Inflow:       10,
				Outflow:      16,
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddOutflow__Failure__Small amount but net outflow",
			flow: types.Flow{
				Inflow:       1,
				Outflow:      10,
				ChannelValue: totalValue,
			},
			amount:   5,
			succeeds: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initialFlow := test.flow
			err := test.flow.AddOutflow(test.amount, quota)
			actualFlow := test.flow

			if test.succeeds {
				require.NoError(t, err)
				require.Equal(t, test.expectedFlow, actualFlow)
			} else {
				require.ErrorContains(t, err, "Outflow exceeds quota", "test: %v", test.name)
				require.Equal(t, initialFlow, actualFlow)
			}
		})
	}
}
