package utils

import (
	"fmt"
	"strconv"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var WHITELIST = map[string]bool{
	"stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq": true,
	"stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7": true,
}

func FilterDepositRecords(arr []recordstypes.DepositRecord, condition func(recordstypes.DepositRecord) bool) (ret []recordstypes.DepositRecord) {
	for _, elem := range arr {
		if condition(elem) {
			ret = append(ret, elem)
		}
	}
	return ret
}

func Int64ToCoinString(amount int64, denom string) string {
	return strconv.FormatInt(amount, 10) + denom
}

func ValidateWhitelistedAddress(address string) error {
	if !WHITELIST[address] {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid creator address (%s)", address))
	}
	return nil
}
