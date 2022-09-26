package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/app_router module sentinel errors
var (
	ErrInvalidVersion = sdkerrors.Register(ModuleName, 1501, "invalid version")
)
