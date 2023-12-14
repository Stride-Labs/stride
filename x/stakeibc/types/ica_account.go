package types

import fmt "fmt"

// Helper function to build the host zone ICA owner in the form "{chainId}.{ICA_TYPE}"
func FormatHostZoneICAOwner(chainId string, accountType ICAAccountType) (result string) {
	return chainId + "." + accountType.String()
}

// Helper function to build the ICA owner for a trade route ICA
// in the form "{chainId}.{rewardDenom}-{hostDenom}.{ICA_TYPE}"
func FormatTradeRouteICAOwner(chainId, rewardDenom, hostDenom string, icaAccountType ICAAccountType) string {
	tradeRouteId := GetTradeRouteId(rewardDenom, hostDenom)
	return FormatTradeRouteICAOwnerFromRouteId(chainId, tradeRouteId, icaAccountType)
}

// Helper function to build the ICA owner for a trade route ICA
// in the form "{chainId}.{tradeRouteId}.{ICA_TYPE}"
func FormatTradeRouteICAOwnerFromRouteId(chainId, tradeRouteId string, icaAccountType ICAAccountType) string {
	return fmt.Sprintf("%s.%s.%s", chainId, tradeRouteId, icaAccountType.String())
}
