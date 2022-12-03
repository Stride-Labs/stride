package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v3/testutil/sample"
)

func TestMsgRebalanceValidators_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRebalanceValidators
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgRebalanceValidators{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address but not whitelisted",
			msg: MsgRebalanceValidators{
				Creator: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address and whitelisted but invalid number of validators to rebalance (minimum not meet)",
			msg: MsgRebalanceValidators{
				// github.com/cosmos/cosmos-sdk@v0.45.11/types/address.go this test require to change Bech32MainPrefix from cosmos to stride in order to pass
				Creator:      "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh",
				NumRebalance: 0,
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "valid address and whitelisted but invalid number of validators to rebalance (too much)",
			msg: MsgRebalanceValidators{
				// github.com/cosmos/cosmos-sdk@v0.45.11/types/address.go this test require to change Bech32MainPrefix from cosmos to stride in order to pass
				Creator:      "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh",
				NumRebalance: 11,
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "valid address and whitelisted",
			msg: MsgRebalanceValidators{
				// github.com/cosmos/cosmos-sdk@v0.45.11/types/address.go this test require to change Bech32MainPrefix from cosmos to stride in order to pass
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
