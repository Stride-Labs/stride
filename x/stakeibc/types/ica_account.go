package types

func FormatICAAccountOwner(id string, accountType ICAAccountType) (result string) {
	return id + "." + accountType.String()
}
