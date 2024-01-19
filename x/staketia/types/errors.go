package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrHostZoneNotFound         = errorsmod.Register(ModuleName, 1901, "host zone not found")
	ErrDelegationRecordNotFound = errorsmod.Register(ModuleName, 1902, "delegation record not found")
)
