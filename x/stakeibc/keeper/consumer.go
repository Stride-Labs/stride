package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Register new stTokens to the consumer reward denom whitelist
func (k Keeper) RegisterStTokenDenomsToWhitelist(ctx sdk.Context, denoms []string) error {
	hostZones := k.GetAllHostZone(ctx)
	allDenomsMap := make(map[string]bool)
	registeredDenomsMap := make(map[string]bool)

	// get all stToken denoms
	for _, zone := range hostZones {
		allDenomsMap[types.StAssetDenomFromHostZoneDenom(zone.HostDenom)] = true
	}

	// get registered denoms in the consumer reward denom whitelist
	consumerParams := k.ConsumerKeeper.GetConsumerParams(ctx)
	for _, denom := range consumerParams.RewardDenoms {
		registeredDenomsMap[denom] = true
	}

	// register new denoms to the whitelist
	for _, denom := range denoms {
		if !allDenomsMap[denom] {
			return types.ErrStTokenNotFound
		} else if registeredDenomsMap[denom] {
			continue
		} else {
			consumerParams.RewardDenoms = append(consumerParams.RewardDenoms, denom)
		}
	}

	k.ConsumerKeeper.SetParams(ctx, consumerParams)
	return nil
}
