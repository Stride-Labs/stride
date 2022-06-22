package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/stakeibc module sentinel errors
var (
	ErrSample                 = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout   = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion         = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrInvalidToken           = sdkerrors.Register(ModuleName, 1502, "invalid token denom (denom is not IBC token)")
	ErrInvalidHostZone        = sdkerrors.Register(ModuleName, 1503, "host zone not registered")
	ErrICAStake               = sdkerrors.Register(ModuleName, 1504, "ICA stake failed")
	ErrValidatorAlreadyExists = sdkerrors.Register(ModuleName, 1505, "validator already exists")
	ErrNoValidatorWeights     = sdkerrors.Register(ModuleName, 1506, "no non-zero validator weights")
	ErrValidatorNotFound      = sdkerrors.Register(ModuleName, 1507, "validator not found")
)
