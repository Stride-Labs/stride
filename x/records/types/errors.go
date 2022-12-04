package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/records module sentinel errors
var (
	ErrInvalidVersion               = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrRedemptionAlreadyExists      = sdkerrors.Register(ModuleName, 1502, "redemption record already exists")
	ErrEpochUnbondingRecordNotFound = sdkerrors.Register(ModuleName, 1503, "epoch unbonding record not found")
	ErrUnknownDepositRecord         = sdkerrors.Register(ModuleName, 1504, "unknown deposit record")
	ErrUnmarshalFailure             = sdkerrors.Register(ModuleName, 1505, "cannot unmarshal")
	ErrAddingHostZone               = sdkerrors.Register(ModuleName, 1506, "could not add hzu to epoch unbonding record")
)
