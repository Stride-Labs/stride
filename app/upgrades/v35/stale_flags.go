package v35

import (
	"fmt"

	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// ResetStaleDelegationChangesInProgress zeroes DelegationChangesInProgress on every validator
// of each in-scope zone whose delegation ICA channel is open with no unacked packet: exactly
// what RestoreInterchainAccount does after a channel restore (spec §5 "Stale in-progress
// flags"). A flag with no ICA behind it is stale by definition, since the callback that clears
// it can never fire, and stale flags are what has the Cosmos Hub pipeline stuck. A zone with a
// packet in flight, no open channel, or the Deprecated flag is logged and left alone.
func ResetStaleDelegationChangesInProgress(ctx sdk.Context, k stakeibckeeper.Keeper) {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if hostZone.Deprecated {
			continue
		}

		owner := stakeibctypes.FormatHostZoneICAOwner(hostZone.ChainId, stakeibctypes.ICAAccountType_DELEGATION)
		portId, err := icatypes.NewControllerPortID(owner)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("v35: %s: unable to build the delegation port id, skipping flag reset: %s", hostZone.ChainId, err))
			continue
		}

		channelId, found := k.ICAControllerKeeper.GetOpenActiveChannel(ctx, hostZone.ConnectionId, portId)
		if !found {
			ctx.Logger().Info(fmt.Sprintf("v35: %s: no open delegation channel, skipping flag reset (restore the channel after the upgrade)", hostZone.ChainId))
			continue
		}
		if unacked := k.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctx, portId, channelId); len(unacked) > 0 {
			ctx.Logger().Info(fmt.Sprintf("v35: %s: %d unacked packet(s) on %s, skipping flag reset", hostZone.ChainId, len(unacked), channelId))
			continue
		}

		// Validators are stored as pointers, so clearing the flag updates hostZone in place
		numReset := 0
		for _, validator := range hostZone.Validators {
			if validator.DelegationChangesInProgress == 0 {
				continue
			}
			ctx.Logger().Info(fmt.Sprintf("v35: %s: resetting %s DelegationChangesInProgress %d -> 0",
				hostZone.ChainId, validator.Address, validator.DelegationChangesInProgress))
			validator.DelegationChangesInProgress = 0
			numReset++
		}
		k.SetHostZone(ctx, hostZone)
		ctx.Logger().Info(fmt.Sprintf("v35: %s: %d stale in-progress flag(s) reset on %s", hostZone.ChainId, numReset, channelId))
	}
}
