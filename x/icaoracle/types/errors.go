package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrOracleNotFound = errorsmod.Register(ModuleName, 1101, "oracle not found")
)
