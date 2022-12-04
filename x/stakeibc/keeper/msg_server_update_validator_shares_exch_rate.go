package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// This kicks off two ICQs, each with a callback, that will update the number of tokens on a validator
// after being slashed. The flow is:
// 1. QueryValidatorExchangeRate (ICQ)
// 2. ValidatorExchangeRateCallback (CALLBACK)
// 3. QueryDelegationsIcq (ICQ)
// 4. DelegatorSharesCallback (CALLBACK)
func (k msgServer) UpdateValidatorSharesExchRate(goCtx context.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.QueryValidatorExchangeRate(ctx, msg)
}
