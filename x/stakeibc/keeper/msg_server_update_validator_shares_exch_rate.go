package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// todo rename for clarity (this is the validator query : 1st step of daisy chain)
func (k msgServer) UpdateValidatorSharesExchRate(goCtx context.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.QueryValidatorExchangeRate(ctx, msg)
}
