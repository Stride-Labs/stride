package types

import (
	"testing"

	"github.com/Stride-Labs/stride/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgSetNumValidators_ValidateBasic(t *testing.T) {
	adminAddress := ""
	for addr, _ := range utils.ADMINS {
		adminAddress = addr
		break
	}
	sdk.GetConfig().SetBech32PrefixForAccount("stride", "pub")

	tests := []struct {
		name string
		msg  MsgSetNumValidators
		err  error
	}{
		{
			name: "valid message",
			msg: MsgSetNumValidators{
				Creator:       adminAddress,
				NumValidators: 1,
			},
		},
		{
			name: "not enough validators",
			msg: MsgSetNumValidators{
				Creator:       adminAddress,
				NumValidators: 0,
			},
			err: ErrInvalidNumValidator,
		},
		{
			name: "invalid address",
			msg: MsgSetNumValidators{
				Creator: "invalid_address",
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
