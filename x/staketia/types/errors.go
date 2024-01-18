package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrUnbondingRecordNotFound = errorsmod.Register(ModuleName, 1902, "unbonding record not found")
)
