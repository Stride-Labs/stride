package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/v13/x/icacallbacks/types"
)

const TRANSFER = "transfer"

func (k Keeper) Callbacks() icacallbackstypes.ModuleCallbacks {
	return []icacallbackstypes.ICACallback{
		{CallbackId: TRANSFER, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.TransferCallback)},
	}
}
