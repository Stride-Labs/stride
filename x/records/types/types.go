package types

import "fmt"

func UserRedemptionRecordKeyFormatter(chainId string, epochNumber uint64, sender string, timestamp int64) string {
	return fmt.Sprintf("%s.%d.%s.%d", chainId, epochNumber, sender, timestamp) // {chain_id}.{epoch}.{sender}.{timestamp}
}
