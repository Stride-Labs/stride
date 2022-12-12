package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/ratelimit module sentinel errors
var (
	ErrQuotaNameDuplicated = sdkerrors.Register(ModuleName, 1,
		"same quota name already exists")
)
