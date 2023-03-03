package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/autopilot module sentinel errors
var (
	ErrInvalidReceiverData = sdkerrors.Register(ModuleName, 1501, "invalid receiver data")
)
