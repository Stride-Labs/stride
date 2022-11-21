package types

// DONTCOVER

import (
	"errors"
	"fmt"
)

type ErrorInterface interface {
	Error() error
}

// ErrTotaErrSamplelWeightNotSet
type ErrSample struct{}

func (e ErrSample) Error() error {
	return errors.Unwrap(fmt.Errorf("Sample error"))
}

// ErrInvalidPacketTimeout
type ErrInvalidPacketTimeout struct{}

func (e ErrInvalidPacketTimeout) Error() error {
	return errors.Unwrap(fmt.Errorf("Invalid packet timeout"))
}

// ErrInvalidVersion
type ErrInvalidVersion struct{}

func (e ErrInvalidVersion) Error() error {
	return errors.Unwrap(fmt.Errorf("Invalid version"))
}

// ErrCallbackHandlerNotFound
type ErrCallbackHandlerNotFound struct{}

func (e ErrCallbackHandlerNotFound) Error() error {
	return errors.Unwrap(fmt.Errorf("Icacallback handler not found"))
}

// ErrCallbackIdNotFound
type ErrCallbackIdNotFound struct{}

func (e ErrCallbackIdNotFound) Error() error {
	return errors.Unwrap(fmt.Errorf("Icacallback ID not found"))
}

// ErrCallbackFailed
type ErrCallbackFailed struct{}

func (e ErrCallbackFailed) Error() error {
	return errors.Unwrap(fmt.Errorf("Icacallback failed"))
}

// ErrCallbackDataNotFound
type ErrCallbackDataNotFound struct{}

func (e ErrCallbackDataNotFound) Error() error {
	return errors.Unwrap(fmt.Errorf("Icacallback data not found"))
}

// ErrTxMsgData
type ErrTxMsgData struct{}

func (e ErrTxMsgData) Error() error {
	return errors.Unwrap(fmt.Errorf("TxMsgData fetch failed"))
}
