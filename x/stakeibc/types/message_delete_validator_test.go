package types_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v27/app/apptesting"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

func TestMsgDeleteValidator_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgDeleteValidator
		err  error
	}{
		{
			name: "invalid address",
			msg: types.MsgDeleteValidator{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address but not whitelisted",
			msg: types.MsgDeleteValidator{
				Creator: apptesting.SampleStrideAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
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
