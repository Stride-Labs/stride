package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func UserRedemptionRecordKeyFormatter(chainId string, epochNumber sdk.Int, sender string) string {
	return fmt.Sprintf("%s.%d.%s", chainId, epochNumber, sender) // {chain_id}.{epoch}.{sender}
}
