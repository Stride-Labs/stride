package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/utils"
)

// ibc transfers tokens from the foreign hub community pool deposit ICA address onto Stride hub
// then as an atomic action, liquid stake the tokens with Autopilot commands in the ibc message memos
func (k Keeper) IBCTransferCommunityPoolICATokensToStride(ctx sdk.Context, communityPoolHostZoneId string, token sdk.Coin) error {
	k.Logger(ctx).Info(fmt.Sprintf("Transfering %d %s tokens from community pool deposit ICA to Stride hub holding address", token.Amount.Int64(), token.Denom))

	// TODO: add safety check here on if the amount is greater than a threshold to avoid many, frequent, small transfers
	//       threshold might be a config we tune to avoid ddos type attacks, for now using 0 as hard coded threshold
	if token.Amount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] The amount %v to transfer did not meet the minimum threshold!", token.Amount))
		return errors.New("Transfer Amount below threshold!")
	}

	hostZone, hostZoneFound := k.GetHostZone(ctx, communityPoolHostZoneId)
	if !hostZoneFound || hostZone.Halted {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Host zone not found or halted!"))
		return errors.New("No active host zone found!")
	}

	if hostZone.CommunityPoolDepositIcaAddress == "" || hostZone.CommunityPoolHoldingAddress == "" {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Unknown send or recieve address! DepositICAAddress: %s HoldingAddress: %s", 
			hostZone.CommunityPoolDepositIcaAddress, 
			hostZone.CommunityPoolHoldingAddress))
		return errors.New("Critical addresses missing from hostZone config!")
	}

	// TODO: add more detailed check logic that the denom is something we know how to liquid stake -- here is where a token whitelist could filter
	// Could also add logic for some coins which could be liquid staked to be distributed instead, traded to different denom, etc.

	// Some denoms can be liquid staked, others might not be but we still want to return all of those to the community pool for distribution.  
	// Everything in the deposit ICA needs to come over to the Stride holding account but only liquid-stakable assets will need autopilot 
	shouldLiquidStake := k.DenomCanLiquidStake(ctx, token.Denom)

	// ibc transfer tokens from foreign hub deposit ICA address to Stride hub "holding" address
	err := k.IBCTransferCommunityPoolTokens(ctx, token, hostZone, shouldLiquidStake)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Failed to submit transfer to host zone, HostZone: %v, Channel: %v, Coin: %v, SendAddress: %v, RecAddress: %v",
			hostZone.ChainId, hostZone.TransferChannelId, token, hostZone.CommunityPoolDepositIcaAddress, hostZone.CommunityPoolHoldingAddress))
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] err {%s}", err.Error()))
		return err
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "[IBCTransferCommunityPoolICATokensToStride] Transfer community pool tokens to Stride successfully initiated!"))

	return nil
}

// Check if incoming denom is the base denom of a known hostZone for which Stride is able to liquid stake
func (k Keeper) DenomCanLiquidStake(ctx sdk.Context, denom string) bool {
	hostZones := k.GetAllHostZone(ctx)
	for _, zone := range hostZones {
		if denom == zone.HostDenom {
			return true
		}
	}
	return false
}
