package types_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func TestMsgRebalanceValidators_ValidateBasic(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	tests := []struct {
		name string
		msg  types.MsgRebalanceValidators
		err  error
	}{
		{
			name: "successful message min vals",
			msg: types.MsgRebalanceValidators{
				Creator:      validAdminAddress,
				NumRebalance: 1,
			},
		},
		{
			name: "successful message mid vals",
			msg: types.MsgRebalanceValidators{
				Creator:      validAdminAddress,
				NumRebalance: 1,
			},
		},
		{
			name: "successful message max vals",
			msg: types.MsgRebalanceValidators{
				Creator:      validAdminAddress,
				NumRebalance: 1,
			},
		},
		{
			name: "too few validators",
			msg: types.MsgRebalanceValidators{
				Creator:      validAdminAddress,
				NumRebalance: 0,
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "too many validators",
			msg: types.MsgRebalanceValidators{
				Creator:      validAdminAddress,
				NumRebalance: 2,
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid address",
			msg: types.MsgRebalanceValidators{
				Creator: invalidAddress,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid admin address",
			msg: types.MsgRebalanceValidators{
				Creator: validNotAdminAddress,
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
