package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrHostZoneNotFound                  = errorsmod.Register(ModuleName, 1901, "host zone not found")
	ErrDelegationRecordNotFound          = errorsmod.Register(ModuleName, 1902, "delegation record not found")
	ErrUnbondingRecordNotFound           = errorsmod.Register(ModuleName, 1903, "unbonding record not found")
	ErrDelegationRecordAlreadyExists     = errorsmod.Register(ModuleName, 1904, "delegation record already exists")
	ErrUnbondingRecordAlreadyExists      = errorsmod.Register(ModuleName, 1905, "unbonding record already exists")
	ErrRedemptionRecordNotFound          = errorsmod.Register(ModuleName, 1906, "redemption record not found")
	ErrBrokenUnbondingRecordInvariant    = errorsmod.Register(ModuleName, 1907, "broken unbonding record invariant")
	ErrRedemptionRateOutsideSafetyBounds = errorsmod.Register(ModuleName, 1908, "host zone redemption rate outside safety bounds")
	ErrInvalidRedemptionRateBounds       = errorsmod.Register(ModuleName, 1909, "invalid host zone redemption rate inner bounds")
	ErrInvalidAmountBelowMinimum         = errorsmod.Register(ModuleName, 1910, "amount provided is too small")
	ErrInvalidAdmin                      = errorsmod.Register(ModuleName, 1911, "signer is not an admin")
	ErrHostZoneHalted                    = errorsmod.Register(ModuleName, 1912, "host zone is halted")
	ErrHostZoneNotHalted                 = errorsmod.Register(ModuleName, 1913, "host zone is not halted")
	ErrInvariantBroken                   = errorsmod.Register(ModuleName, 1914, "invariant broken")
	ErrUnbondAmountToLarge               = errorsmod.Register(ModuleName, 1915, "unbonding more than exists on host zone")
	ErrInsufficientFunds                 = errorsmod.Register(ModuleName, 1916, "not enough tokens in wallet")
)
