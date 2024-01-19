package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrHostZoneNotFound               = errorsmod.Register(ModuleName, 1901, "host zone not found")
	ErrDelegationRecordNotFound       = errorsmod.Register(ModuleName, 1902, "delegation record not found")
	ErrUnbondingRecordNotFound        = errorsmod.Register(ModuleName, 1903, "unbonding record not found")
	ErrBrokenUnbondingRecordInvariant = errorsmod.Register(ModuleName, 1904, "broken unbonding record invariant")
	ErrRedemptionRecordNotFound       = errorsmod.Register(ModuleName, 1905, "redemption record not found")
)
