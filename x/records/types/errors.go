package types

// DONTCOVER

import errorsmod "cosmossdk.io/errors"

// x/records module sentinel errors
var (
	ErrInvalidVersion               = errorsmod.Register(ModuleName, 1501, "invalid version")
	ErrRedemptionAlreadyExists      = errorsmod.Register(ModuleName, 1502, "redemption record already exists")
	ErrEpochUnbondingRecordNotFound = errorsmod.Register(ModuleName, 1503, "epoch unbonding record not found")
	ErrUnknownDepositRecord         = errorsmod.Register(ModuleName, 1504, "unknown deposit record")
	ErrUnmarshalFailure             = errorsmod.Register(ModuleName, 1505, "cannot unmarshal")
	ErrAddingHostZone               = errorsmod.Register(ModuleName, 1506, "could not add hzu to epoch unbonding record")
)
