package types

import fmt "fmt"

// Helper function to build the host zone ICA owner in the form "{chainId}.{ICA_TYPE}"
func FormatHostZoneICAOwner(chainId string, accountType ICAAccountType) (result string) {
	return chainId + "." + accountType.String()
}

// Helper function to build the ICA owner for a trade route ICA
// in the form "{chainId}.{rewardDenom}-{hostDenom}.{ICA_TYPE}"
func FormatTradeRouteICAOwner(chainId, rewardDenom, hostDenom string, icaAccountType ICAAccountType) string {
	return fmt.Sprintf("%s.%s-%s.%s", chainId, rewardDenom, hostDenom, icaAccountType.String())
}

// Helper function to build the ICA owner for a trade route ICA
// in the form "{chainId}.{rewardDenom}-{hostDenom}.{ICA_TYPE}"
// from an ICAAccount
func FormatTradeRouteICAOwnerFromAccount(tradeRouteId string, icaAccount ICAAccount) string {
	return fmt.Sprintf("%s.%s.%s", icaAccount.ChainId, tradeRouteId, icaAccount.Type.String())
}
