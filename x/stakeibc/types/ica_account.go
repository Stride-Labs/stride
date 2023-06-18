package types

func FormatICAAccountOwner(chainID string, accountType ICAAccountType) (result string) {
	return chainID + "." + accountType.String()
}
