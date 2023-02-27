package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ParsedReceiver struct {
	ShouldLiquidStake bool
	ShouldRedeemStake bool
	StrideAccAddress  sdk.AccAddress
	ResultReceiver    string
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
			ShouldRedeemStake: false,
		}, nil
	case 2:
		addressPart := parts[0]
		functionPart := parts[1]

		// verify the rightmost field is stakeibc/LiquidStake
		switch functionPart {
		case "stakeibc/LiquidStake":
			strideAccAddress, err := sdk.AccAddressFromBech32(addressPart)
			if err != nil {
				return nil, err
			}

			return &ParsedReceiver{
				ShouldLiquidStake: true,
				ShouldRedeemStake: false,
				StrideAccAddress:  strideAccAddress,
			}, nil
		}
		return &ParsedReceiver{
			ShouldLiquidStake: false,
			ShouldRedeemStake: false,
		}, nil
	case 3:
		addressPart := parts[0]
		functionPart := parts[1]
		receiverPart := parts[2]

		switch functionPart {
		case "stakeibc/LiquidStakeAndIBCTransfer":
			strideAccAddress, err := sdk.AccAddressFromBech32(addressPart)
			if err != nil {
				return nil, err
			}

			return &ParsedReceiver{
				ShouldLiquidStake: true,
				ShouldRedeemStake: false,
				StrideAccAddress:  strideAccAddress,
				ResultReceiver:    receiverPart,
			}, nil
		case "stakeibc/RedeemStake":
			strideAccAddress, err := sdk.AccAddressFromBech32(addressPart)
			if err != nil {
				return nil, err
			}

			return &ParsedReceiver{
				ShouldLiquidStake: false,
				ShouldRedeemStake: true,
				StrideAccAddress:  strideAccAddress,
				ResultReceiver:    receiverPart,
			}, nil
		}
		return &ParsedReceiver{
			ShouldLiquidStake: false,
			ShouldRedeemStake: false,
		}, nil
	default:
		return nil, ErrInvalidReceiverData
	}
}
