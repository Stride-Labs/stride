package types

import fmt "fmt"

func FormatICAAccountOwner(chainId string, accountType ICAAccountType) (result string) {
	return chainId + "." + accountType.String()
}

func (a ICAAccount) FormatTradeRouteICAOwner(tradeRouteId string) string {
	return fmt.Sprintf("%s.%s.%s", a.ChainId, tradeRouteId, a.Type.String())
}
