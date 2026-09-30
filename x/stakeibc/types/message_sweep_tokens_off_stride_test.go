package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// withSweepOperator sets the sweep operator constant for one test and restores it after
func withSweepOperator(t *testing.T, address string) {
	t.Helper()
	previous := types.SweepOperatorAddress
	types.SweepOperatorAddress = address
	t.Cleanup(func() { types.SweepOperatorAddress = previous })
}

func randomStrideAddresses(n int) []string {
	addresses := make([]string, 0, n)
	for _, account := range apptesting.CreateRandomAccounts(n) {
		addresses = append(addresses, account.String())
	}
	return addresses
}

func TestMsgSweepTokensOffStride_ValidateBasic(t *testing.T) {
	// Nothing in the app sets the bech32 prefix at init: run in isolation, the SDK config still says
	// "cosmos" and every "valid" case fails on the stride addresses
	apptesting.SetupConfig()
	operator := apptesting.SampleStrideAddress()
	withSweepOperator(t, operator)

	holders := randomStrideAddresses(3)
	tooMany := randomStrideAddresses(types.MaxSweepAddressesPerTx + 1)
	atCap := randomStrideAddresses(types.MaxSweepAddressesPerTx)

	tests := []struct {
		name string
		msg  types.MsgSweepTokensOffStride
		err  error
	}{
		{
			name: "valid: one denom, three holders",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, holders),
		},
		{
			name: "valid: three denoms incl. an ibc voucher, at the address cap",
			msg: *types.NewMsgSweepTokensOffStride(operator,
				[]string{"stuatom", "ustrd", "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"}, atCap),
		},
		{
			name: "invalid creator address",
			msg:  *types.NewMsgSweepTokensOffStride("invalid_address", []string{"stuatom"}, holders),
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "creator is not the sweep operator",
			msg:  *types.NewMsgSweepTokensOffStride(holders[0], []string{"stuatom"}, holders),
			err:  sdkerrors.ErrUnauthorized,
		},
		{
			name: "empty denom list",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{}, holders),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid denom string",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"st uatom"}, holders),
			err:  sdkerrors.ErrInvalidCoins,
		},
		{
			name: "duplicate denom",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom", "stuatom"}, holders),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty address list",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, []string{}),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "batch over the cap",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, tooMany),
			err:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "address with the wrong bech32 prefix",
			msg: *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"},
				[]string{"osmo1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz"}),
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "duplicate address",
			msg:  *types.NewMsgSweepTokensOffStride(operator, []string{"stuatom"}, []string{holders[0], holders[0]}),
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
			require.Equal(t, types.TypeMsgSweepTokensOffStride, tt.msg.Type())
			require.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(operator)}, tt.msg.GetSigners())
		})
	}
}

// While the operator constant is empty (as shipped by PR 4) the gate rejects everyone
func TestMsgSweepTokensOffStride_ValidateBasic_OperatorNotConfigured(t *testing.T) {
	apptesting.SetupConfig()
	withSweepOperator(t, "")

	msg := types.NewMsgSweepTokensOffStride(apptesting.SampleStrideAddress(), []string{"stuatom"}, randomStrideAddresses(1))
	err := msg.ValidateBasic()
	require.ErrorIs(t, err, types.ErrSweepOperatorNotConfigured, fmt.Sprintf("got %v", err))
}
