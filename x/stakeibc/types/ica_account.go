package types

func FormatICAAccountOwner(chainId string, accountType ICAAccountType) (result string) {
	return chainId + "." + accountType.String()
}
