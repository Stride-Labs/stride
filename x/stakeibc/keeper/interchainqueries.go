package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icqtypes "github.com/Stride-Labs/stride/v18/x/interchainquery/types"
)

// Submit a validator sharesToTokens rate ICQ as triggered either manually or epochly with a conservative timeout
func (k Keeper) QueryValidatorSharesToTokensRate(ctx sdk.Context, chainId string, validatorAddress string) error {
	timeoutDuration := time.Hour * 24
	timeoutPolicy := icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE
	callbackData := []byte{}
	return k.SubmitValidatorSharesToTokensRateICQ(ctx, chainId, validatorAddress, callbackData, timeoutDuration, timeoutPolicy)
}
