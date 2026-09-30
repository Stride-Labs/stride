package v35

import (
	"fmt"

	ratelimitkeeper "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icaoraclekeeper "github.com/Stride-Labs/stride/v34/x/icaoracle/keeper"
)

// DeactivateICAOracles turns every ICA oracle off (spec §5 "Oracles and rate limits"). The
// redemption rate no longer moves, so there is nothing to post. ToggleOracle(false) skips the
// channel validation the true case does, so a half-registered oracle still deactivates.
func DeactivateICAOracles(ctx sdk.Context, k icaoraclekeeper.Keeper) {
	for _, oracle := range k.GetAllOracles(ctx) {
		if !oracle.Active {
			continue
		}
		if err := k.ToggleOracle(ctx, oracle.ChainId, false); err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: unable to deactivate oracle %s, skipping: %s", oracle.ChainId, err))
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("v35: oracle %s deactivated", oracle.ChainId))
	}
}

// RemoveAllRateLimits empties the rate limiter: every limit, blacklisted denom and whitelisted
// address pair. The token sweep sends most of each stToken's on-Stride supply out over channel-5
// in a few days, which no limit would allow, and there is no mint path left to protect. The
// module and middleware stay in the stack with empty state. Stakedym's BeginBlocker re-adds
// `stadym` to the blacklist every block while its redemption rate stays above its max bound
// (1.1005 against 1.1), which is intended (stakedym is deprecated and out of scope), so the post-upgrade blacklist on mainnet holds exactly `stadym`.
func RemoveAllRateLimits(ctx sdk.Context, k *ratelimitkeeper.Keeper) {
	for _, rateLimit := range k.GetAllRateLimits(ctx) {
		k.RemoveRateLimit(ctx, rateLimit.Path.Denom, rateLimit.Path.ChannelOrClientId)
		ctx.Logger().Info(fmt.Sprintf("v35: rate limit removed for %s on %s", rateLimit.Path.Denom, rateLimit.Path.ChannelOrClientId))
	}
	for _, denom := range k.GetAllBlacklistedDenoms(ctx) {
		k.RemoveDenomFromBlacklist(ctx, denom)
		ctx.Logger().Info(fmt.Sprintf("v35: %s removed from the rate-limit blacklist", denom))
	}
	for _, pair := range k.GetAllWhitelistedAddressPairs(ctx) {
		k.RemoveWhitelistedAddressPair(ctx, pair.Sender, pair.Receiver)
		ctx.Logger().Info(fmt.Sprintf("v35: whitelisted pair %s -> %s removed", pair.Sender, pair.Receiver))
	}
}
