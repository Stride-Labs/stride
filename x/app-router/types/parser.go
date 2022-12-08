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
// stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl|stakeibc/LiquidStake
// need args: Creator, HostDenom, Amount
func ParseReceiverData(receiverData string) (*ParsedReceiver, error) {
	parts := strings.Split(receiverData, "|")

	switch len(parts) {
	case 1:
		return &ParsedReceiver{
			ShouldLiquidStake: false,
		}, nil
	case 2:
		addressPart := parts[0]
		functionPart := parts[1]

		// verify the rightmost field is stakeibc/LiquidStake
		if functionPart != "stakeibc/LiquidStake" {
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
	default:
		return nil, ErrInvalidReceiverData
	}
}
