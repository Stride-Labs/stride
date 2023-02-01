package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/liquidgov module sentinel errors
var (
	ErrLockupNotFound        = sdkerrors.Register(ModuleName, 1601, "lockup not found")
	ErrNotEnoughLockupTokens = sdkerrors.Register(ModuleName, 1602, "not enough lockup tokens")
)
