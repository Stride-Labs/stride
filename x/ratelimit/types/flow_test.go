package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

func TestAddInflow(t *testing.T) {
	totalValue := sdkmath.NewInt(100)
	quota := types.Quota{
		MaxPercentRecv: sdkmath.NewInt(10),
		MaxPercentSend: sdkmath.NewInt(10),
		DurationHours:  uint64(1),
	}

	tests := []struct {
		name         string
		flow         types.Flow
		expectedFlow types.Flow
		amount       sdkmath.Int
		succeeds     bool
	}{
		{
			name: "AddInflow__Successful__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.ZeroInt(),
				Outflow:      sdkmath.ZeroInt(),
				ChannelValue: totalValue,
			},
			amount: sdkmath.NewInt(5),
			expectedFlow: types.Flow{
				Inflow:       sdkmath.NewInt(5),
				Outflow:      sdkmath.ZeroInt(),
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddInflow__Successful__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(100),
				Outflow:      sdkmath.NewInt(100),
				ChannelValue: totalValue,
			},
			amount: sdkmath.NewInt(5),
			expectedFlow: types.Flow{
				Inflow:       sdkmath.NewInt(105),
				Outflow:      sdkmath.NewInt(100),
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddInflow__Failure__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.ZeroInt(),
				Outflow:      sdkmath.ZeroInt(),
				ChannelValue: totalValue,
			},
			amount:   sdkmath.NewInt(15),
			succeeds: false,
		},
		{
			name: "AddInflow__Failure__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(100),
				Outflow:      sdkmath.NewInt(100),
				ChannelValue: totalValue,
			},
			amount:   sdkmath.NewInt(15),
			succeeds: false,
		},
		{
			name: "AddInflow__Successful__Large amount but net outflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(1),
				Outflow:      sdkmath.NewInt(10),
				ChannelValue: totalValue,
			},
			amount: sdkmath.NewInt(15),
			expectedFlow: types.Flow{
				Inflow:       sdkmath.NewInt(16),
				Outflow:      sdkmath.NewInt(10),
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddInflow__Failure__Small amount but net inflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(10),
				Outflow:      sdkmath.NewInt(1),
				ChannelValue: totalValue,
			},
			amount:   sdkmath.NewInt(5),
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
	totalValue := sdkmath.NewInt(100)
	quota := types.Quota{
		MaxPercentRecv: sdkmath.NewInt(10),
		MaxPercentSend: sdkmath.NewInt(10),
		DurationHours:  uint64(1),
	}

	tests := []struct {
		name         string
		flow         types.Flow
		expectedFlow types.Flow
		amount       sdkmath.Int
		succeeds     bool
	}{
		{
			name: "AddOutflow__Successful__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.ZeroInt(),
				Outflow:      sdkmath.ZeroInt(),
				ChannelValue: totalValue,
			},
			amount: sdkmath.NewInt(5),
			expectedFlow: types.Flow{
				Inflow:       sdkmath.ZeroInt(),
				Outflow:      sdkmath.NewInt(5),
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddOutflow__Successful__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(100),
				Outflow:      sdkmath.NewInt(100),
				ChannelValue: totalValue,
			},
			amount: sdkmath.NewInt(5),
			expectedFlow: types.Flow{
				Inflow:       sdkmath.NewInt(100),
				Outflow:      sdkmath.NewInt(105),
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddOutflow__Failure__Zero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.ZeroInt(),
				Outflow:      sdkmath.ZeroInt(),
				ChannelValue: totalValue,
			},
			amount:   sdkmath.NewInt(15),
			succeeds: false,
		},
		{
			name: "AddOutflow__Failure__Nonzero inflow and outflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(100),
				Outflow:      sdkmath.NewInt(100),
				ChannelValue: totalValue,
			},
			amount:   sdkmath.NewInt(15),
			succeeds: false,
		},
		{
			name: "AddOutflow__Succeesful__Large amount but net inflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(10),
				Outflow:      sdkmath.NewInt(1),
				ChannelValue: totalValue,
			},
			amount: sdkmath.NewInt(15),
			expectedFlow: types.Flow{
				Inflow:       sdkmath.NewInt(10),
				Outflow:      sdkmath.NewInt(16),
				ChannelValue: totalValue,
			},
			succeeds: true,
		},
		{
			name: "AddOutflow__Failure__Small amount but net outflow",
			flow: types.Flow{
				Inflow:       sdkmath.NewInt(1),
				Outflow:      sdkmath.NewInt(10),
				ChannelValue: totalValue,
			},
			amount:   sdkmath.NewInt(5),
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
