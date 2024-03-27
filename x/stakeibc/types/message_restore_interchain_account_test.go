package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v21/testutil/sample"
)

func TestMsgRestoreInterchainAccount_ValidateBasic(t *testing.T) {
	validChainId := "chain-0"
	validConnectionId := "connection-0"
	validHostZoneOwner := "chain-0.DELEGATION"
	validTradeRouteOwner := "chain-0.reward.host.CONVERTER_TRADE"

	tests := []struct {
		name string
		msg  MsgRestoreInterchainAccount
		err  string
	}{
		{
			name: "valid host zone message",
			msg: MsgRestoreInterchainAccount{
				Creator:      sample.AccAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: validHostZoneOwner,
			},
		},
		{
			name: "valid trade route message",
			msg: MsgRestoreInterchainAccount{
				Creator:      sample.AccAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: validTradeRouteOwner,
			},
		},
		{
			name: "missing chain id",
			msg: MsgRestoreInterchainAccount{
				Creator:      sample.AccAddress(),
				ChainId:      "",
				ConnectionId: validConnectionId,
				AccountOwner: validHostZoneOwner,
			},
			err: "chain ID must be specified",
		},
		{
			name: "missing connection id",
			msg: MsgRestoreInterchainAccount{
				Creator:      sample.AccAddress(),
				ChainId:      validChainId,
				ConnectionId: "con-0",
				AccountOwner: validHostZoneOwner,
			},
			err: "connection ID must be specified",
		},
		{
			name: "missing account owner",
			msg: MsgRestoreInterchainAccount{
				Creator:      sample.AccAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: "",
			},
			err: "ICA account owner must be specified",
		},
		{
			name: "chain id does not match owner",
			msg: MsgRestoreInterchainAccount{
				Creator:      sample.AccAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: "chain-1.reward.host.CONVERTER_TRADE",
			},
			err: "ICA account owner does not match chain ID",
		},
		{
			name: "invalid address",
			msg: MsgRestoreInterchainAccount{
				Creator: "invalid_address",
			},
			err: "invalid address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
