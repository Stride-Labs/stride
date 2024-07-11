package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/airdrop module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 1101, "sample error")
)
