// x/stakeibc/types/message_transfer_from_ica_test.go
package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func TestMsgTransferFromIca_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig() // stride bech32 prefix; without it the admin address fails to parse when the test runs alone
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	amount := sdk.NewCoin("uatom", sdkmath.NewInt(1000))

	tests := []struct {
		name string
		msg  types.MsgTransferFromIca
		err  error
	}{
		{name: "valid delegation", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}},
		{name: "valid withdrawal", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_WITHDRAWAL, Amount: amount}},
		{name: "valid fee", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_FEE, Amount: amount}},
		{name: "valid redemption", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_REDEMPTION, Amount: amount}},
		{name: "valid foreign denom", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "dydx-mainnet-1", IcaType: types.ICAAccountType_WITHDRAWAL, Amount: sdk.NewCoin("ibc/8E27BA2D5493AF5636760E354E46004562C46AB7EC0CC4C1CA14E9E20E2545B5", sdkmath.NewInt(3))}},
		{name: "invalid creator", msg: types.MsgTransferFromIca{Creator: invalidAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: sdkerrors.ErrInvalidAddress},
		{name: "not admin", msg: types.MsgTransferFromIca{Creator: validNotAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: sdkerrors.ErrInvalidAddress},
		{name: "missing chain id", msg: types.MsgTransferFromIca{Creator: validAdminAddress, IcaType: types.ICAAccountType_DELEGATION, Amount: amount}, err: sdkerrors.ErrInvalidRequest},
		{name: "community pool ICA not allowed", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, Amount: amount}, err: sdkerrors.ErrInvalidRequest},
		{name: "converter ICA not allowed", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_CONVERTER_TRADE, Amount: amount}, err: sdkerrors.ErrInvalidRequest},
		{name: "zero amount", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: sdk.NewCoin("uatom", sdkmath.ZeroInt())}, err: sdkerrors.ErrInvalidRequest},
		{name: "invalid denom", msg: types.MsgTransferFromIca{Creator: validAdminAddress, ChainId: "cosmoshub-4", IcaType: types.ICAAccountType_DELEGATION, Amount: sdk.Coin{Denom: "", Amount: sdkmath.NewInt(1)}}, err: sdkerrors.ErrInvalidRequest},
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
