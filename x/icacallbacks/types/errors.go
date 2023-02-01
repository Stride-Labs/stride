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
	ErrInvalidAcknowledgement  = fmt.Errorf("invalid acknowledgement")
)
