package types

// DONTCOVER

import (
	"fmt"
)

var (
	ErrSample                  = fmt.Errorf("Sample error")
	ErrInvalidPacketTimeout    = fmt.Errorf("Invalid packet timeout")
	ErrInvalidVersion          = fmt.Errorf("Invalid version")
	ErrCallbackHandlerNotFound = fmt.Errorf("Icacallback handler not found")
	ErrCallbackIdNotFound      = fmt.Errorf("Icacallback ID not found")
	ErrCallbackFailed          = fmt.Errorf("Icacallback failed")
	ErrCallbackDataNotFound    = fmt.Errorf("Icacallback data not found")
	ErrTxMsgData               = fmt.Errorf("TxMsgData fetch failed")
)

// type ErrorInterface interface {
// 	Error() error
// }

// // ErrTotaErrSamplelWeightNotSet
// type ErrSample struct{}

// func (e ErrSample) Error() error {
// 	return fmt.Errorf("Sample error")
// }

// ErrInvalidPacketTimeout
// type ErrInvalidPacketTimeout struct{}

// func (e ErrInvalidPacketTimeout) Error() error {
// 	return fmt.Errorf("Invalid packet timeout")
// }

// ErrInvalidVersion
// type ErrInvalidVersion struct{}

// func (e ErrInvalidVersion) Error() error {
// 	return fmt.Errorf("Invalid version")
// }

// ErrCallbackHandlerNotFound
// type ErrCallbackHandlerNotFound struct{}

// func (e ErrCallbackHandlerNotFound) Error() error {
// 	return fmt.Errorf("Icacallback handler not found")
// }

// ErrCallbackIdNotFound
// type ErrCallbackIdNotFound struct{}

// func (e ErrCallbackIdNotFound) Error() error {
// 	return fmt.Errorf("Icacallback ID not found")
// }

// ErrCallbackFailed
// type ErrCallbackFailed struct{}

// func (e ErrCallbackFailed) Error() error {
// 	return fmt.Errorf("Icacallback failed")
// }

// ErrCallbackDataNotFound
// type ErrCallbackDataNotFound struct{}

// func (e ErrCallbackDataNotFound) Error() error {
// 	return fmt.Errorf("Icacallback data not found")
// }

// ErrTxMsgData
// type ErrTxMsgData struct{}

// func (e ErrTxMsgData) Error() error {
// 	return fmt.Errorf("TxMsgData fetch failed")
// }
