package utils

import (
	"strconv"

	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

func FilterDepositRecords(arr []stakeibctypes.DepositRecord, condition func(stakeibctypes.DepositRecord) bool) (ret []stakeibctypes.DepositRecord) {
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
