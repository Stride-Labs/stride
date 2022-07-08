package types

import "fmt"

func UserRedemptionRecordKeyFormatter(chainId string, epochNumber uint64, sender string) string {
	return fmt.Sprintf("%s.%d.%s", chainId, epochNumber, sender) // {chain_id}.{epoch}.{sender}
}
