package utils

import (
	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
)

// GetDenomPrefix returns the ICS20 hop prefix for a port/channel pair, e.g. "transfer/channel-0/".
func GetDenomPrefix(portID, channelID string) string {
	return fmt.Sprintf("%s/", transfertypes.NewHop(portID, channelID))
}

// GetPrefixedDenom returns the ICS20 trace path for a base denom in the form "port/channel/baseDenom".
func GetPrefixedDenom(portID, channelID, baseDenom string) string {
	return fmt.Sprintf("%s%s", GetDenomPrefix(portID, channelID), baseDenom)
}

// GetIBCDenom returns the hashed IBC denom ("ibc/<hash>") for a base denom received through the given port/channel hop.
func GetIBCDenom(portID, channelID, baseDenom string) string {
	return transfertypes.ExtractDenomFromPath(GetPrefixedDenom(portID, channelID, baseDenom)).IBCDenom()
}
