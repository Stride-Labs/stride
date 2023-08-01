package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/v11/x/icacallbacks/types"
)

const (
	ICACallbackID_CastVotes = "castvotes"
)

func (k Keeper) Callbacks() icacallbackstypes.ModuleCallbacks {
	return []icacallbackstypes.ICACallback{
		{CallbackId: ICACallbackID_CastVotes, CallbackFunc: icacallbackstypes.ICACallbackFunction(k.CastVotesCallback)},
	}
}
