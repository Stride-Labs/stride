package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/icqoracle module sentinel errors
var (
	ErrTokenPriceAlreadyExists = sdkerrors.Register(ModuleName, 16001, "token price already exists")
	ErrQuotePriceNotFound      = sdkerrors.Register(ModuleName, 16002, "common quote price not found")
)
