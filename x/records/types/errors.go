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

// type ErrorInterface interface {
// 	Error() error
// }

// type ErrInvalidVersion struct{}

// func (e ErrInvalidVersion) Error() error {
// 	return fmt.Errorf("invalid version")
// }

// type ErrRedemptionAlreadyExists struct{}

// func (e ErrRedemptionAlreadyExists) Error() error {
// 	return fmt.Errorf("redemption record already exists")
// }

// type ErrEpochUnbondingRecordNotFound struct{}

// func (e ErrEpochUnbondingRecordNotFound) Error() error {
// 	return fmt.Errorf("epoch unbonding record not found")
// }

// type ErrUnknownDepositRecord struct{}

// func (e ErrUnknownDepositRecord) Error() error {
// 	return fmt.Errorf("unknown deposit record")
// }

// type ErrUnmarshalFailure struct{}

// func (e ErrUnmarshalFailure) Error() error {
// 	return fmt.Errorf("cannot unmarshal")
// }

// type ErrAddingHostZone struct{}

// func (e ErrAddingHostZone) Error() error {
// 	return fmt.Errorf("could not add hzu to epoch unbonding record")
// }
