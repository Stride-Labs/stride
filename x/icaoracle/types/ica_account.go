package types

import fmt "fmt"

const (
	ICAAccountType_Oracle = "ORACLE"
)

func FormatICAAccountOwner(chainId string, accountType string) string {
	return fmt.Sprintf("%s.%s", chainId, accountType)
}
