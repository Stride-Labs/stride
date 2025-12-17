package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v31/app/apptesting"
	"github.com/Stride-Labs/stride/v31/x/stakeibc/types"
)

func TestMsgUpdateTradeRoute(t *testing.T) {
	apptesting.SetupConfig()

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	validDenom := "denom"
	validMinTransferAmount := sdkmath.NewInt(100)

	tests := []struct {
		name string
		msg  types.MsgUpdateTradeRoute
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgUpdateTradeRoute{
				Authority:         authority,
				HostDenom:         validDenom,
				RewardDenom:       validDenom,
				MinTransferAmount: validMinTransferAmount,
			},
		},
		{
			name: "invalid authority",
			msg: types.MsgUpdateTradeRoute{
				Authority:         "",
				HostDenom:         validDenom,
				RewardDenom:       validDenom,
				MinTransferAmount: validMinTransferAmount,
			},
			err: "invalid authority address",
		},
		{
			name: "invalid host denom",
			msg: types.MsgUpdateTradeRoute{
				Authority:         authority,
				HostDenom:         "",
				RewardDenom:       validDenom,
				MinTransferAmount: validMinTransferAmount,
			},
			err: "missing host denom",
		},
		{
			name: "invalid reward denom",
			msg: types.MsgUpdateTradeRoute{
				Authority:         authority,
				HostDenom:         validDenom,
				RewardDenom:       "",
				MinTransferAmount: validMinTransferAmount,
			},
			err: "missing reward denom",
		},
		{
			name: "invalid min transfer amount - nil",
			msg: types.MsgUpdateTradeRoute{
				Authority:   authority,
				HostDenom:   validDenom,
				RewardDenom: validDenom,
			},
			err: "min transfer amount must be greater than or equal to zero",
		},
		{
			name: "invalid min transfer amount - negative",
			msg: types.MsgUpdateTradeRoute{
				Authority:         authority,
				HostDenom:         validDenom,
				RewardDenom:       validDenom,
				MinTransferAmount: sdkmath.OneInt().Neg(),
			},
			err: "min transfer amount must be greater than or equal to zero",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "update_trade_route")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), authority)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
