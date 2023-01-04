package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/ratelimit module sentinel errors
var (
	ErrRateLimitAlreadyExists = sdkerrors.Register(ModuleName, 1,
		"ratelimit key duplicated")
	ErrRateLimitNotFound = sdkerrors.Register(ModuleName, 2,
		"rate limit not found")
	ErrZeroChannelValue = sdkerrors.Register(ModuleName, 3,
		"channel value is zero")
	ErrQuotaExceeded = sdkerrors.Register(ModuleName, 4,
		"quota exceeded")
	ErrInvalidClientState = sdkerrors.Register(ModuleName, 5,
		"unable to determine client state from channelId")
	ErrChannelNotFound = sdkerrors.Register(ModuleName, 6,
		"channel does not exist")
)
