package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/v16/x/icacallbacks/types"
)

const (
	ICACallbackID_Transfer = "transfer"
)

func (k Keeper) Callbacks() icacallbackstypes.ModuleCallbacks {
	return []icacallbackstypes.ICACallback{
		{CallbackId: ICACallbackID_Transfer, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.TransferCallback)},
	}
}
