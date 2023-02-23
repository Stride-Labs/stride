package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrOracleNotFound            = errorsmod.Register(ModuleName, 1, "oracle not found")
	ErrClientStateNotTendermint  = errorsmod.Register(ModuleName, 2, "client state is not tendermint")
	ErrHostConnectionNotFound    = errorsmod.Register(ModuleName, 3, "host connection not found")
	ErrOracleAlreadyExists       = errorsmod.Register(ModuleName, 4, "oracle already exists")
	ErrOracleNotInstantiated     = errorsmod.Register(ModuleName, 5, "oracle not instantiated")
	ErrOracleAlreadyInstantiated = errorsmod.Register(ModuleName, 6, "oracle already instantiated")
	ErrOracleICANotRegistered    = errorsmod.Register(ModuleName, 7, "oracle ICA channel has not been registered")
	ErrOracleInactive            = errorsmod.Register(ModuleName, 8, "oracle is inactive")
	ErrInvalidICARequest         = errorsmod.Register(ModuleName, 9, "invalid ICA request")
	ErrInvalidICAResponse        = errorsmod.Register(ModuleName, 10, "invalid ICA response")
	ErrMarshalFailure            = errorsmod.Register(ModuleName, 11, "unable to marshal data structure")
	ErrUnmarshalFailure          = errorsmod.Register(ModuleName, 12, "unable to unmarshal data structure")
)
