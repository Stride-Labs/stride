package types

import "strings"

// QUESTION FOR REVIEWER: Should I place this here or in keeper/path.go?

// Checks if the path's denom is native by checking if it has an ibc/ prefix
func (p *Path) IsNative() bool {
	return !strings.HasPrefix(p.TraceDenom, "ibc/")
}
