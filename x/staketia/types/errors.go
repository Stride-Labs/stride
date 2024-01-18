package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrDelegationRecordNotFound = errorsmod.Register(ModuleName, 1901, "delegation record not found")
)
