package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/ratelimit module sentinel errors
var (
	ErrRateLimitKeyDuplicated = sdkerrors.Register(ModuleName, 1,
		"ratelimit key duplicated")
	ErrRateLimitKeyNotFound = sdkerrors.Register(ModuleName, 2,
		"ratelimit key not found")
)
