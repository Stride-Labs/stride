package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/v16/x/icacallbacks/types"
)

const (
	ICACallbackID_TransferFallback = "transfer-fallback"
)

func (k Keeper) Callbacks() icacallbackstypes.ModuleCallbacks {
	return []icacallbackstypes.ICACallback{
		{CallbackId: ICACallbackID_TransferFallback, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.TransferFallbackCallback)},
	}
}
