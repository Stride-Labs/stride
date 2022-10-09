package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/icacallbacks module sentinel errors
var (
	ErrSample                  = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout    = sdkerrors.Register(ModuleName, 1500, "invalid packet timeout")
	ErrInvalidVersion          = sdkerrors.Register(ModuleName, 1501, "invalid version")
	ErrCallbackHandlerNotFound = sdkerrors.Register(ModuleName, 1502, "icacallback handler not found")
	ErrCallbackIdNotFound      = sdkerrors.Register(ModuleName, 1503, "icacallback ID not found")
	ErrCallbackFailed          = sdkerrors.Register(ModuleName, 1504, "icacallback failed")
	ErrCallbackDataNotFound    = sdkerrors.Register(ModuleName, 1505, "icacallback data not found")
	ErrTxMsgData               = sdkerrors.Register(ModuleName, 1506, "txMsgData fetch failed")
)
