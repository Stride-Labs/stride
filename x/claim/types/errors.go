package types

// DONTCOVER

import (
	"errors"
	"fmt"
)

type ErrorInterface interface {
	Error() error
}

// ErrTotalWeightNotSet
type ErrTotalWeightNotSet struct{}

func (e ErrTotalWeightNotSet) Error() error {
	return errors.Unwrap(fmt.Errorf("total weight not set"))
}

// ErrTotalWeightParse
type ErrTotalWeightParse struct{}

func (e ErrTotalWeightParse) Error() error {
	return errors.Unwrap(fmt.Errorf("total weight parse error"))
}

// ErrFailedToGetTotalWeight
type ErrFailedToGetTotalWeight struct{}

func (e ErrFailedToGetTotalWeight) Error() error {
	return errors.Unwrap(fmt.Errorf("failed to get total weight"))
}

// ErrFailedToParseDec
type ErrFailedToParseDec struct{}

func (e ErrFailedToParseDec) Error() error {
	return errors.Unwrap(fmt.Errorf("failed to parse dec from str"))
}

// ErrAirdropAlreadyExists
type ErrAirdropAlreadyExists struct{}

func (e ErrAirdropAlreadyExists) Error() error {
	return errors.Unwrap(fmt.Errorf("airdrop with same identifier already exists"))
}

// ErrDistributorAlreadyExists
type ErrDistributorAlreadyExists struct{}

func (e ErrDistributorAlreadyExists) Error() error {
	return errors.Unwrap(fmt.Errorf("airdrop with same distributor already exists"))
}

// ErrInvalidAmount
type ErrInvalidAmount struct{}

func (e ErrInvalidAmount) Error() error {
	return errors.Unwrap(fmt.Errorf("cannot claim negative tokens"))
}
