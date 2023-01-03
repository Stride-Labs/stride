package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func UserRedemptionRecordKeyFormatter(chainId string, epochNumber sdk.Int, sender string) string {
	return fmt.Sprintf("%s.%s.%s", chainId, epochNumber.String(), sender) // {chain_id}.{epoch}.{sender}
}

func UserRedemptionRecordKeyFormatterForErr(chainId string, epochNumber sdk.Int, sender string) string {
	return fmt.Sprintf("%s.%s.%s", chainId, epochNumber.String(), sender) // {chain_id}.{epoch}.{sender}
}
