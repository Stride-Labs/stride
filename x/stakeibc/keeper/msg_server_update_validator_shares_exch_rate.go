package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// This kicks off two ICQs, each with a callback, that will update the number of tokens on a validator
// after being slashed. The flow is:
// 1. QueryValidatorSharesToTokensRate (ICQ)
// 2. ValidatorSharesToTokensRate (CALLBACK)
// 3. SubmitDelegationICQ (ICQ)
// 4. DelegatorSharesCallback (CALLBACK)
func (k msgServer) UpdateValidatorSharesExchRate(goCtx context.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.QueryValidatorSharesToTokensRate(ctx, msg.ChainId, msg.Valoper); err != nil {
		return nil, err
	}
	return &types.MsgUpdateValidatorSharesExchRateResponse{}, nil
}
