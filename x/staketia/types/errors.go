package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrUnbondingRecordNotFound        = errorsmod.Register(ModuleName, 1902, "unbonding record not found")
	ErrBrokenUnbondingRecordInvariant = errorsmod.Register(ModuleName, 1903, "broken unbonding record invariant")
)
