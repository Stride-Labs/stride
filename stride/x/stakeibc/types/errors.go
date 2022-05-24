package types

// DONTCOVER

import (
	sdkerrors "github.com/Stride-Labs/cosmos-sdk/types/errors"
)

// x/stakeibc module sentinel errors
var (
	ErrSample               = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion       = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrInvalidToken         = sdkerrors.Register(ModuleName, 1502, "invalid token denom (denom is not IBC token)")
)
