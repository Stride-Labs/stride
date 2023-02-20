package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrOracleNotFound            = errorsmod.Register(ModuleName, 1, "oracle not found")
	ErrClientStateNotTendermint  = errorsmod.Register(ModuleName, 2, "client state is not tendermint")
	ErrHostConnectionNotFound    = errorsmod.Register(ModuleName, 3, "host connection not found")
	ErrOracleAlreadyInstantiated = errorsmod.Register(ModuleName, 4, "oracle already instantiated")
	ErrOracleICANotRegistered    = errorsmod.Register(ModuleName, 5, "oracle ICA channel has not been registered")
	ErrMarshalFailure            = errorsmod.Register(ModuleName, 6, "unable to marshal data structure")
	ErrUnmarshalFailure          = errorsmod.Register(ModuleName, 7, "unable to unmarshal data structure")
	ErrInvalidICAResponse        = errorsmod.Register(ModuleName, 8, "invalid ICA response")
)
