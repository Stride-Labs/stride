package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

func UserRedemptionRecordKeyFormatter(chainId string, epochNumber sdkmath.Int, sender string) string {
	return fmt.Sprintf("%s.%s.%s", chainId, epochNumber.String(), sender) // {chain_id}.{epoch}.{sender}
}
