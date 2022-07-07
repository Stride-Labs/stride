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
	ErrEpochNotFound          = sdkerrors.Register(ModuleName, 1505, "epoch not found")
	ErrRecordNotFound         = sdkerrors.Register(ModuleName, 1506, "record not found")
	ErrInvalidAmount          = sdkerrors.Register(ModuleName, 1507, "invalid unstaking amount")
	ErrValidatorAlreadyExists = sdkerrors.Register(ModuleName, 1508, "validator already exists")
	ErrNoValidatorWeights     = sdkerrors.Register(ModuleName, 1509, "no non-zero validator weights")
	ErrValidatorNotFound      = sdkerrors.Register(ModuleName, 1510, "validator not found")
	ErrWeightsNotDifferent    = sdkerrors.Register(ModuleName, 1511, "validator weights haven't changed")
	ErrValidatorDelegationChg = sdkerrors.Register(ModuleName, 1512, "can't change delegation on validator")
	ErrAcctNotScopedForFunc   = sdkerrors.Register(ModuleName, 1513, "this account can't call this function")
	ErrInsufficientFunds      = sdkerrors.Register(ModuleName, 1514, "balance is insufficient")
)
