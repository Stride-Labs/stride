package types

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v3/testutil/sample"
)

func TestMsgRedeemStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRedeemStake
		err  error
	}{
		{
			name: "success",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Receiver: sample.AccAddress(),
				Amount:   uint64(1),
			},
		},
		{
			name: "invalid creator",
			msg: MsgRedeemStake{
				Creator:  "invalid_address",
				HostZone: "GAIA",
				Receiver: sample.AccAddress(),
				Amount:   uint64(1),
			},
			err: fmt.Errorf("%s", &Error{errorCode: "invalid creator address (decoding bech32 failed: invalid separator index -1)"}),
		},
		{
			name: "no host zone",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Amount:   uint64(1),
			},
			err: fmt.Errorf("%s", &Error{errorCode: "required field is missing%!(EXTRA string=host zone cannot be empty)"}),
		},
		{
			name: "invalid receiver",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Amount:   uint64(1),
			},
			err: fmt.Errorf("%s", &Error{errorCode: "receiver cannot be empty"}),
		},
		{
			name: "amount max int",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Receiver: sample.AccAddress(),
				Amount:   math.MaxUint64,
			},
			err: fmt.Errorf("%s", &Error{errorCode: "invalid amount%!(EXTRA string=amount liquid staked must be less than math.MaxInt64 %d, int=9223372036854775807)"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorAs(t, err, &tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
