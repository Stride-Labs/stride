package types

// DONTCOVER

import (
	"fmt"
)

// x/records module sentinel errors
var (
	ErrInvalidVersion               = fmt.Errorf("invalid version")
	ErrRedemptionAlreadyExists      = fmt.Errorf("redemption record already exists")
	ErrEpochUnbondingRecordNotFound = fmt.Errorf("epoch unbonding record not found")
	ErrUnknownDepositRecord         = fmt.Errorf("unknown deposit record")
	ErrUnmarshalFailure             = fmt.Errorf("cannot unmarshal")
	ErrAddingHostZone               = fmt.Errorf("could not add hzu to epoch unbonding record")
)
