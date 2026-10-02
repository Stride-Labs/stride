// x/stakeibc/types/message_transfer_staketia_claim_balance_test.go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

func TestMsgTransferStaketiaClaimBalance_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig() // stride bech32 prefix; without it the admin address fails to parse when the test runs alone
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	tests := []struct {
		name string
		msg  types.MsgTransferStaketiaClaimBalance
		err  error
	}{
		{name: "valid zero amount (whole balance)", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress, Amount: sdkmath.ZeroInt()}},
		{name: "valid nil amount (whole balance)", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress}},
		{name: "valid positive amount", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress, Amount: sdkmath.NewInt(1_000_000)}},
		{name: "invalid creator", msg: types.MsgTransferStaketiaClaimBalance{Creator: invalidAddress}, err: sdkerrors.ErrInvalidAddress},
		{name: "not admin", msg: types.MsgTransferStaketiaClaimBalance{Creator: validNotAdminAddress}, err: sdkerrors.ErrInvalidAddress},
		{name: "negative amount", msg: types.MsgTransferStaketiaClaimBalance{Creator: validAdminAddress, Amount: sdkmath.NewInt(-1)}, err: sdkerrors.ErrInvalidRequest},
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
