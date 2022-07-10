package types

import (
	"testing"

	"github.com/Stride-Labs/stride/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgClaimUndelegatedTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgClaimUndelegatedTokens
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgClaimUndelegatedTokens{
				Creator: "invalid_address",
				MaxClaims: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgClaimUndelegatedTokens{
				Creator: sample.AccAddress(),
				MaxClaims: 1,
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
