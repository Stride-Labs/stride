package types

// DONTCOVER

import (
	"fmt"
)

// x/records module sentinel errors
// var (
//
//	ErrInvalidVersion               = sdkerrors.Register(ModuleName, 1501, "invalid version")
//	ErrRedemptionAlreadyExists      = sdkerrors.Register(ModuleName, 1502, "redemption record already exists")
//	ErrEpochUnbondingRecordNotFound = sdkerrors.Register(ModuleName, 1503, "epoch unbonding record not found")
//	ErrUnknownDepositRecord         = sdkerrors.Register(ModuleName, 1504, "unknown deposit record")
//	ErrUnmarshalFailure             = sdkerrors.Register(ModuleName, 1505, "cannot unmarshal")
//	ErrAddingHostZone               = sdkerrors.Register(ModuleName, 1506, "could not add hzu to epoch unbonding record")
//
// )
type ErrorInterface interface {
	Error() error
}

type ErrInvalidVersion struct{}

func (e ErrInvalidVersion) Error() error {
	return fmt.Errorf("Invalid version")
}

type ErrRedemptionAlreadyExists struct{}

func (e ErrRedemptionAlreadyExists) Error() error {
	return fmt.Errorf("Redemption already exists")
}

type ErrEpochUnbondingRecordNotFound struct{}

func (e ErrEpochUnbondingRecordNotFound) Error() error {
	return fmt.Errorf("Epoch unbonding record not found")
}

type ErrUnknownDepositRecord struct{}

func (e ErrUnknownDepositRecord) Error() error {
	return fmt.Errorf("Unknown deposit record")
}

type ErrUnmarshalFailure struct{}

func (e ErrUnmarshalFailure) Error() error {
	return fmt.Errorf("Unmarshal failure")
}

type ErrAddingHostZone struct{}

func (e ErrAddingHostZone) Error() error {
	return fmt.Errorf("Error adding host zone")
}
