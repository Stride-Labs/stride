package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrOracleNotFound = errorsmod.Register(ModuleName, 1, "oracle not found")
)
