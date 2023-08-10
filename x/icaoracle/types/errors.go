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
	ErrOracleICAChannelClosed    = errorsmod.Register(ModuleName, 8, "oracle ICA channel is closed")
	ErrOracleInactive            = errorsmod.Register(ModuleName, 9, "oracle is inactive")
	ErrInvalidICARequest         = errorsmod.Register(ModuleName, 10, "invalid ICA request")
	ErrInvalidICAResponse        = errorsmod.Register(ModuleName, 11, "invalid ICA response")
	ErrInvalidCallback           = errorsmod.Register(ModuleName, 12, "invalid callback data")
	ErrICAAccountDoesNotExist    = errorsmod.Register(ModuleName, 13, "ICA account does not exist")
	ErrInvalidGenesisState       = errorsmod.Register(ModuleName, 14, "Invalid genesis state")
	ErrUnableToRestoreICAChannel = errorsmod.Register(ModuleName, 15, "unable to restore oracle ICA channel")
)
