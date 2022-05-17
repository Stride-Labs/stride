package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const CSTEST = "somenewspace"

// x/stakeibc module sentinel errors
var (
	ErrSample               = sdkerrors.Register(CSTEST, 1100, "sample error")
	ErrInvalidPacketTimeout = sdkerrors.Register(CSTEST, 1500, "invalid packet timeout")
	ErrInvalidVersion       = sdkerrors.Register(CSTEST, 1501, "invalid version")
	ErrInvalidToken         = sdkerrors.Register(CSTEST, 1502, "invalid token denom (denom is not IBC token)")
)
