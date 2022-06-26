package utils

import (
	"strconv"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
)

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
