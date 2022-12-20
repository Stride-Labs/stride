package types_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	types "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	apptesting "github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/testutil/sample"
)

func TestMsgRebalanceValidators_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	tests := []struct {
		name string
		msg  types.MsgRebalanceValidators
		err  error
	}{
		{
			name: "invalid address",
			msg: types.MsgRebalanceValidators{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address but not whitelisted",
			msg: types.MsgRebalanceValidators{
				Creator: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address and whitelisted",
			msg: types.MsgRebalanceValidators{
				Creator:      "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh",
				NumRebalance: 10,
			},
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
