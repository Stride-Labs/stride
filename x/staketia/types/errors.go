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
	ErrInvalidUnbondingRecord            = errorsmod.Register(ModuleName, 1910, "unbonding record in incorrect state")
	ErrInvalidTxHash                     = errorsmod.Register(ModuleName, 1911, "tx hash is invalid")
	ErrInsufficientFunds                 = errorsmod.Register(ModuleName, 1912, "not enough funds")
	ErrDelegationRecordAlreadyExists     = errorsmod.Register(ModuleName, 1913, "delegation record already exists")
	ErrInvariantBroken                   = errorsmod.Register(ModuleName, 1914, "invariant broken")
	ErrInvalidAdmin                      = errorsmod.Register(ModuleName, 1915, "signer is not an admin")
	ErrHostZoneNotHalted                 = errorsmod.Register(ModuleName, 1916, "host zone is not halted")
	ErrInvalidBounds                     = errorsmod.Register(ModuleName, 1917, "invalid inner bounds")
	ErrUnbondAmountToLarge               = errorsmod.Register(ModuleName, 1918, "unbonding more than exists on host zone")
	ErrUnbondingRecordAlreadyExists      = errorsmod.Register(ModuleName, 1919, "unbonding record already exists")
	ErrInvalidRecordType                 = errorsmod.Register(ModuleName, 1920, "invalid record type")
	ErrInvalidHostZone                   = errorsmod.Register(ModuleName, 1921, "invalid host zone during genesis")
	ErrInvalidGenesisRecords             = errorsmod.Register(ModuleName, 1922, "invalid records during genesis")
)
