package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrHostZoneNotFound                  = errorsmod.Register(ModuleName, 1901, "host zone not found")
	ErrDelegationRecordNotFound          = errorsmod.Register(ModuleName, 1902, "delegation record not found")
	ErrUnbondingRecordNotFound           = errorsmod.Register(ModuleName, 1903, "unbonding record not found")
	ErrBrokenUnbondingRecordInvariant    = errorsmod.Register(ModuleName, 1904, "broken unbonding record invariant")
	ErrInvalidBounds                     = errorsmod.Register(ModuleName, 1905, "invalid inner bounds")
	ErrRedemptionRecordNotFound          = errorsmod.Register(ModuleName, 1906, "redemption record not found")
	ErrInvalidAmountBelowMinimum         = errorsmod.Register(ModuleName, 1907, "amount provided is too small")
	ErrInvalidAdmin                      = errorsmod.Register(ModuleName, 1908, "signer is not an admin")
	ErrRedemptionRateOutsideSafetyBounds = errorsmod.Register(ModuleName, 1909, "redemption rate outside safety bounds")
	ErrInvalidRedemptionRateBounds       = errorsmod.Register(ModuleName, 1910, "invalid redemption rate bounds")
	ErrHostZoneHalted                    = errorsmod.Register(ModuleName, 1911, "host zone is halted")
	ErrHostZoneNotHalted                 = errorsmod.Register(ModuleName, 1912, "host zone is not halted")
	ErrUnbondAmountToLarge               = errorsmod.Register(ModuleName, 1913, "unbonding more than exists on host zone")
	ErrInsufficientFunds                 = errorsmod.Register(ModuleName, 1914, "not enough tokens in wallet")
	ErrDelegationRecordAlreadyExists     = errorsmod.Register(ModuleName, 1915, "delegation record already exists")
	ErrInvariantBroken                   = errorsmod.Register(ModuleName, 1916, "invariant broken")
)
