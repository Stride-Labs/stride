package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v31/app/apptesting"
	"github.com/Stride-Labs/stride/v31/x/stakeibc/types"
)

func TestMsgRestoreInterchainAccount_ValidateBasic(t *testing.T) {
	validChainId := "chain-0"
	validConnectionId := "connection-0"
	validHostZoneOwner := "chain-0.DELEGATION"
	validTradeRouteOwner := "chain-0.reward.host.CONVERTER_TRADE"

	tests := []struct {
		name string
		msg  types.MsgRestoreInterchainAccount
		err  string
	}{
		{
			name: "valid host zone message",
			msg: types.MsgRestoreInterchainAccount{
				Creator:      apptesting.SampleStrideAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: validHostZoneOwner,
			},
		},
		{
			name: "valid trade route message",
			msg: types.MsgRestoreInterchainAccount{
				Creator:      apptesting.SampleStrideAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: validTradeRouteOwner,
			},
		},
		{
			name: "missing chain id",
			msg: types.MsgRestoreInterchainAccount{
				Creator:      apptesting.SampleStrideAddress(),
				ChainId:      "",
				ConnectionId: validConnectionId,
				AccountOwner: validHostZoneOwner,
			},
			err: "chain ID must be specified",
		},
		{
			name: "missing connection id",
			msg: types.MsgRestoreInterchainAccount{
				Creator:      apptesting.SampleStrideAddress(),
				ChainId:      validChainId,
				ConnectionId: "con-0",
				AccountOwner: validHostZoneOwner,
			},
			err: "connection ID must be specified",
		},
		{
			name: "missing account owner",
			msg: types.MsgRestoreInterchainAccount{
				Creator:      apptesting.SampleStrideAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: "",
			},
			err: "ICA account owner must be specified",
		},
		{
			name: "chain id does not match owner",
			msg: types.MsgRestoreInterchainAccount{
				Creator:      apptesting.SampleStrideAddress(),
				ChainId:      validChainId,
				ConnectionId: validConnectionId,
				AccountOwner: "chain-1.reward.host.CONVERTER_TRADE",
			},
			err: "ICA account owner does not match chain ID",
		},
		{
			name: "invalid address",
			msg: types.MsgRestoreInterchainAccount{
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
