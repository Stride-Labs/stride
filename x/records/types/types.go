package types

import "fmt"

func UserRedemptionRecordKeyFormatter(chainId string, epochNumber uint64, receiver string) string {
	return fmt.Sprintf("%s.%d.%s", chainId, epochNumber, receiver) // {chain_id}.{epoch}.{receiver}
}
