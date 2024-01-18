package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrHostZoneNotFound  = errorsmod.Register(ModuleName, 1902, "host zone not found")
	ErrHostZoneCorrupted = errorsmod.Register(ModuleName, 1903, "host zone data could not unmarshal")	
)
