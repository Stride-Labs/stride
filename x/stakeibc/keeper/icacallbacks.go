package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
)

const (
	ICACallbackID_Delegate   = "delegate"
	ICACallbackID_Claim      = "claim"
	ICACallbackID_Undelegate = "undelegate"
	ICACallbackID_Reinvest   = "reinvest"
	ICACallbackID_Redemption = "redemption"
	ICACallbackID_Rebalance  = "rebalance"
	ICACallbackID_Detokenize = "detokenize"
)

func (k Keeper) Callbacks() icacallbackstypes.ModuleCallbacks {
	return []icacallbackstypes.ICACallback{
		{CallbackId: ICACallbackID_Delegate, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.DelegateCallback)},
		{CallbackId: ICACallbackID_Claim, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.ClaimCallback)},
		{CallbackId: ICACallbackID_Undelegate, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.UndelegateCallback)},
		{CallbackId: ICACallbackID_Reinvest, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.ReinvestCallback)},
		{CallbackId: ICACallbackID_Redemption, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.RedemptionCallback)},
		{CallbackId: ICACallbackID_Rebalance, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.RebalanceCallback)},
		{CallbackId: ICACallbackID_Detokenize, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.DetokenizeCallback)},
	}
}
