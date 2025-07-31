package types_test

import (
	"math"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v27/app/apptesting"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

func TestMsgClaimUndelegatedTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgClaimUndelegatedTokens
		err  error
	}{
		{
			name: "success",
			msg: types.MsgClaimUndelegatedTokens{
				Creator:    apptesting.SampleStrideAddress(),
				Receiver:   apptesting.SampleHostAddress(),
				HostZoneId: "GAIA",
				Epoch:      uint64(1),
			},
		},
		{
			name: "invalid address",
			msg: types.MsgClaimUndelegatedTokens{
				Creator:    "invalid_address",
				Receiver:   apptesting.SampleHostAddress(),
				HostZoneId: "GAIA",
				Epoch:      uint64(1),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "no host zone",
			msg: types.MsgClaimUndelegatedTokens{
				Creator:  apptesting.SampleStrideAddress(),
				Receiver: apptesting.SampleHostAddress(),
				Epoch:    uint64(1),
			},
			err: types.ErrRequiredFieldEmpty,
		},
		{
			name: "epoch max int",
			msg: types.MsgClaimUndelegatedTokens{
				Creator:    apptesting.SampleStrideAddress(),
				Receiver:   apptesting.SampleHostAddress(),
				HostZoneId: "GAIA",
				Epoch:      math.MaxUint64,
			},
			err: types.ErrInvalidAmount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
