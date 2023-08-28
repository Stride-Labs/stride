package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func TestValidateICASetup(t *testing.T) {
	validConnectionId := "connection-0"
	validChannelId := "channel-0"
	validPortId := "port-0"
	validIcaAddress := "ica-address"

	tests := []struct {
		name   string
		oracle types.Oracle
		err    string
	}{
		{
			name: "successful ICA setup",
			oracle: types.Oracle{
				ConnectionId: validConnectionId,
				ChannelId:    validChannelId,
				PortId:       validPortId,
				IcaAddress:   validIcaAddress,
			},
		},
		{
			name: "invalid connection-id",
			oracle: types.Oracle{
				ConnectionId: "",
				ChannelId:    validChannelId,
				PortId:       validPortId,
				IcaAddress:   validIcaAddress,
			},
			err: "connectionId is empty",
		},
		{
			name: "invalid channel-id",
			oracle: types.Oracle{
				ConnectionId: validConnectionId,
				ChannelId:    "",
				PortId:       validPortId,
				IcaAddress:   validIcaAddress,
			},
			err: "channelId is empty",
		},
		{
			name: "invalid port-id",
			oracle: types.Oracle{
				ConnectionId: validConnectionId,
				ChannelId:    validChannelId,
				PortId:       "",
				IcaAddress:   validIcaAddress,
			},
			err: "portId is empty",
		},
		{
			name: "invalid ICA address",
			oracle: types.Oracle{
				ConnectionId: validConnectionId,
				ChannelId:    validChannelId,
				PortId:       validPortId,
				IcaAddress:   "",
			},
			err: "ICAAddress is empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.oracle.ValidateICASetup(), "test: %v", test.name)
			} else {
				require.ErrorContains(t, test.oracle.ValidateICASetup(), test.err, "test: %v", test.name)
			}
		})
	}
}

func TestValidateContractInstantiated(t *testing.T) {
	require.NoError(t, types.Oracle{ContractAddress: "contract"}.ValidateContractInstantiated())
	require.ErrorContains(t, types.Oracle{}.ValidateContractInstantiated(), "contract address is empty")
}
