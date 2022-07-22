package types

import (
	"testing"

	"github.com/Stride-Labs/stride/testutil/sample"
	"github.com/Stride-Labs/stride/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgRestoreInterchainAccount_ValidateBasic(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("stride", "pub")
	adminAddress := ""
	for address, _ := range utils.ADMINS {
		adminAddress = address
		break
	}

	tests := []struct {
		name string
		msg  MsgRestoreInterchainAccount
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgRestoreInterchainAccount{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "not admin address",
			msg: MsgRestoreInterchainAccount{
				Creator: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid message",
			msg: MsgRestoreInterchainAccount{
				Creator: adminAddress,
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
