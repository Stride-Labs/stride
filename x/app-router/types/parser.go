package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ParsedReceiver struct {
	ShouldLiquidStake bool
	StrideAccAddress  sdk.AccAddress
}

// {stride_address}|{module_id}/{function_id}
// for now, only allow liquid staking transactions
// TODO: consider - addresses that previously *could not* liquid stake, now *can*
// stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl|stakeibc/liquidstake
// need args: Creator, HostDenom, Amount
func ParseReceiverData(receiverData string) (*ParsedReceiver, error) {
	parts := strings.Split(receiverData, "|")
	addressPart := parts[0]

	// Standard address
	if len(parts) == 1 && addressPart != "" {
		return &ParsedReceiver{
			ShouldLiquidStake: false,
		}, nil
	}

	functionPart := parts[1]
	// verify the rightmost field is stakeibc/liquidstake
	if functionPart != "stakeibc/liquidstake" || len(parts) != 2 {
		return &ParsedReceiver{
			ShouldLiquidStake: false,
		}, nil
	}

	strideAccAddress, err := sdk.AccAddressFromBech32(addressPart)
	if err != nil {
		return nil, err
	}

	return &ParsedReceiver{
		ShouldLiquidStake: true,
		StrideAccAddress:  strideAccAddress,
	}, nil
}
