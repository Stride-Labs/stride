package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/autopilot module sentinel errors
var (
	ErrInvalidReceiverData = errorsmod.Register(ModuleName, 1501, "invalid receiver data")
)
