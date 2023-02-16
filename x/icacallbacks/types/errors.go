package types

// DONTCOVER

import errorsmod "cosmossdk.io/errors"

// x/icacallbacks module sentinel errors
var (
	ErrSample                  = errorsmod.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout    = errorsmod.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion          = errorsmod.Register(ModuleName, 1501, "invalid version")
	ErrCallbackHandlerNotFound = errorsmod.Register(ModuleName, 1502, "icacallback handler not found")
	ErrCallbackIdNotFound      = errorsmod.Register(ModuleName, 1503, "icacallback ID not found")
	ErrCallbackFailed          = errorsmod.Register(ModuleName, 1504, "icacallback failed")
	ErrCallbackDataNotFound    = errorsmod.Register(ModuleName, 1505, "icacallback data not found")
	ErrTxMsgData               = errorsmod.Register(ModuleName, 1506, "txMsgData fetch failed")
	ErrInvalidAcknowledgement  = errorsmod.Register(ModuleName, 1507, "invalid acknowledgement")
)
