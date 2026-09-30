// x/stakeibc/types/message_undelegate_from_validators_test.go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgUndelegateFromValidators_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig() // stride bech32 prefix; without it the admin address fails to parse when the test runs alone
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	valA := types.ValidatorUndelegation{Address: "cosmosvaloper1aaa", Offset: sdkmath.ZeroInt()}
	valB := types.ValidatorUndelegation{Address: "cosmosvaloper1bbb", Offset: sdkmath.NewInt(5)}

	tests := []struct {
		name string
		msg  types.MsgUndelegateFromValidators
		err  error
	}{
		{
			name: "valid empty list (drain everything)",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4"},
		},
		{
			name: "valid explicit list with offsets",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{valA, valB}},
		},
		{
			name: "valid nil offset (treated as zero)",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: "cosmosvaloper1ccc"}}},
		},
		{
			name: "invalid creator",
			msg:  types.MsgUndelegateFromValidators{Creator: invalidAddress, ChainId: "cosmoshub-4"},
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "not admin",
			msg:  types.MsgUndelegateFromValidators{Creator: validNotAdminAddress, ChainId: "cosmoshub-4"},
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "missing chain id",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty validator address",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: ""}}},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "address is not a valoper",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: "cosmos1notavalidator"}}},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "duplicate validator",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{valA, valA}},
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "negative offset",
			msg:  types.MsgUndelegateFromValidators{Creator: validAdminAddress, ChainId: "cosmoshub-4", Validators: []types.ValidatorUndelegation{{Address: "cosmosvaloper1aaa", Offset: sdkmath.NewInt(-1)}}},
			err:  sdkerrors.ErrInvalidRequest,
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
			require.Equal(t, tt.msg.Creator, tt.msg.GetSigners()[0].String())
		})
	}
}
