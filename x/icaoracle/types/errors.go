package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrOracleNotFound            = errorsmod.Register(ModuleName, 1, "oracle not found")
	ErrClientStateNotTendermint  = errorsmod.Register(ModuleName, 2, "client state is not tendermint")
	ErrHostConnectionNotFound    = errorsmod.Register(ModuleName, 3, "host connection not found")
	ErrOracleNotInstantiated     = errorsmod.Register(ModuleName, 4, "oracle already instantiated")
	ErrOracleAlreadyInstantiated = errorsmod.Register(ModuleName, 5, "oracle already instantiated")
	ErrOracleICANotRegistered    = errorsmod.Register(ModuleName, 6, "oracle ICA channel has not been registered")
	ErrInvalidICARequest         = errorsmod.Register(ModuleName, 7, "invalid ICA request")
	ErrInvalidICAResponse        = errorsmod.Register(ModuleName, 8, "invalid ICA response")
	ErrMarshalFailure            = errorsmod.Register(ModuleName, 9, "unable to marshal data structure")
	ErrUnmarshalFailure          = errorsmod.Register(ModuleName, 10, "unable to unmarshal data structure")
)
