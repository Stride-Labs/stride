package types

import (
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
