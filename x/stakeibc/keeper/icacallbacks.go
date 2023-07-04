package keeper

import (
	"github.com/Stride-Labs/stride/v11/x/icacallbacks/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v11/x/icacallbacks/types"
)

const (
	ICACallbackID_Delegate   = "delegate"
	ICACallbackID_Claim      = "claim"
	ICACallbackID_Undelegate = "undelegate"
	ICACallbackID_Reinvest   = "reinvest"
	ICACallbackID_Redemption = "redemption"
	ICACallbackID_Rebalance  = "rebalance"
)

func (k Keeper) Callbacks() icacallbackstypes.ModuleCallbacks {
	return []types.ICACallback{
		{CallbackId: ICACallbackID_Delegate, CallbackFunc: types.ICACallbackFunction(k.DelegateCallback)},
		{CallbackId: ICACallbackID_Claim, CallbackFunc: types.ICACallbackFunction(k.ClaimCallback)},
		{CallbackId: ICACallbackID_Undelegate, CallbackFunc: types.ICACallbackFunction(k.UndelegateCallback)},
		{CallbackId: ICACallbackID_Reinvest, CallbackFunc: types.ICACallbackFunction(k.ReinvestCallback)},
		{CallbackId: ICACallbackID_Redemption, CallbackFunc: types.ICACallbackFunction(k.RedemptionCallback)},
		{CallbackId: ICACallbackID_Rebalance, CallbackFunc: types.ICACallbackFunction(k.RebalanceCallback)},
	}
}
