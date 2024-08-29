package types_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v24/app/apptesting"
	"github.com/Stride-Labs/stride/v24/x/stakeibc/types"
)

func TestMsgDeleteTradeRoute(t *testing.T) {
	apptesting.SetupConfig()

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	validDenom := "denom"

	tests := []struct {
		name string
		msg  types.MsgDeleteTradeRoute
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgDeleteTradeRoute{
				Authority:   authority,
				HostDenom:   validDenom,
				RewardDenom: validDenom,
			},
		},
		{
			name: "invalid authority message",
			msg: types.MsgDeleteTradeRoute{
				Authority:   "",
				HostDenom:   validDenom,
				RewardDenom: validDenom,
			},
			err: "invalid authority address",
		},
		{
			name: "invalid host denom",
			msg: types.MsgDeleteTradeRoute{
				Authority:   authority,
				HostDenom:   "",
				RewardDenom: validDenom,
			},
			err: "missing host denom",
		},
		{
			name: "invalid reward denom",
			msg: types.MsgDeleteTradeRoute{
				Authority:   authority,
				HostDenom:   validDenom,
				RewardDenom: "",
			},
			err: "missing reward denom",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "delete_trade_route")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), authority)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
