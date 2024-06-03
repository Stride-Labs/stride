package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	MaxUnbondingEntries = 7
)

// Per an SDK constraint, we can issue no more than 7 undelegation messages
// in a given unbonding period
//
// The unbonding period dictates the cadence (in number of days) with which we submit
// undelegation messages, such that the 7 messages are spaced out throughout the period
//
// We calculate this by dividing the period by 7 and then adding 1 as a buffer
// Ex: If our unbonding period is 21 days, we issue an undelegation every 4th day
func (h HostZone) GetUnbondingFrequency() uint64 {
	return (h.UnbondingPeriod / MaxUnbondingEntries) + 1
}

// Gets the rebate struct if it exists on the host zone
func (h HostZone) SafelyGetCommunityPoolRebate() (rebate CommunityPoolRebate, exists bool) {
	if h.CommunityPoolRebate == nil {
		return CommunityPoolRebate{}, false
	}
	if h.CommunityPoolRebate.LiquidStakedStTokenAmount.IsNil() || h.CommunityPoolRebate.RebateRate.IsNil() {
		return CommunityPoolRebate{}, false
	}
	return *h.CommunityPoolRebate, true
}

// Generates a new stride-side address on the host zone to escrow deposits
func NewHostZoneDepositAddress(chainId string) sdk.AccAddress {
	key := append([]byte("zone"), []byte(chainId)...)
	return address.Module(ModuleName, key)
}

// Generates a new stride-side module account for a host zone, given an alias
func NewHostZoneModuleAddress(chainId string, accountAlias string) sdk.AccAddress {
	key := append([]byte(chainId), []byte(accountAlias)...)
	return address.Module(ModuleName, key)
}

// TODO [cleanup]: Remove this function and use the one from utils
// isIBCToken checks if the token came from the IBC module
// Each IBC token starts with an ibc/ denom, the check is rather simple
func IsIBCToken(denom string) bool {
	return strings.HasPrefix(denom, "ibc/")
}

// TODO [cleanup]: Remove this function and use the one from utils
// Returns the stDenom from a native denom by appending a st prefix
func StAssetDenomFromHostZoneDenom(hostZoneDenom string) string {
	return "st" + hostZoneDenom
}

// TODO [cleanup]: Remove this function and use the one from utils
// Returns the native denom from an stDenom by removing the st prefix
func HostZoneDenomFromStAssetDenom(stAssetDenom string) string {
	return stAssetDenom[2:]
}
