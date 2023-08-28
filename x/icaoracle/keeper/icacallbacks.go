package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
)

const (
	ICACallbackID_InstantiateOracle = "instantiate_oracle"
	ICACallbackID_UpdateOracle      = "update_oracle"
)

func (k Keeper) Callbacks() icacallbackstypes.ModuleCallbacks {
	return []icacallbackstypes.ICACallback{
		{
			CallbackId:   ICACallbackID_InstantiateOracle,
			CallbackFunc: icacallbackstypes.ICACallbackFunction(k.InstantiateOracleCallback),
		},
		{
			CallbackId:   ICACallbackID_UpdateOracle,
			CallbackFunc: icacallbackstypes.ICACallbackFunction(k.UpdateOracleCallback),
		},
	}
}
