package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrHostZoneNotFound                  = errorsmod.Register(ModuleName, 1901, "host zone not found")
	ErrDelegationRecordNotFound          = errorsmod.Register(ModuleName, 1902, "delegation record not found")
	ErrUnbondingRecordNotFound           = errorsmod.Register(ModuleName, 1903, "unbonding record not found")
	ErrRedemptionRecordNotFound          = errorsmod.Register(ModuleName, 1904, "redemption record not found")
	ErrBrokenUnbondingRecordInvariant    = errorsmod.Register(ModuleName, 1905, "broken unbonding record invariant")
	ErrRedemptionRateOutsideSafetyBounds = errorsmod.Register(ModuleName, 1906, "host zone redemption rate outside safety bounds")
	ErrInvalidRedemptionRateBounds       = errorsmod.Register(ModuleName, 1907, "invalid host zone redemption rate inner bounds")
	ErrHostZoneHalted                    = errorsmod.Register(ModuleName, 1908, "host zone is halted")
	ErrInvalidAmountBelowMinimum         = errorsmod.Register(ModuleName, 1909, "amount provided is too small")
	ErrInvalidBounds                     = errorsmod.Register(ModuleName, 1910, "invalid inner bounds")
	ErrHostZoneNotHalted                 = errorsmod.Register(ModuleName, 1911, "host zone is not halted")
)
