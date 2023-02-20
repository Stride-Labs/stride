package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrOracleNotFound           = errorsmod.Register(ModuleName, 1, "oracle not found")
	ErrClientStateNotTendermint = errorsmod.Register(ModuleName, 2, "client state is not tendermint")
	ErrHostConnectionNotFound   = errorsmod.Register(ModuleName, 3, "host connection not found")
)
