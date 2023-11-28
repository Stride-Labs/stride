package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	ratelimittypes "github.com/Stride-Labs/stride/v16/x/ratelimit/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (k Keeper) OnChanOpenAck(ctx sdk.Context, portID, channelID string) error {
	// Lookup the ICA address and chainId from the port and connection
	controllerConnectionId, found := k.GetConnectionIdFromICAPortId(ctx, portID)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("portId %s has no associated ICA account", portID))
		return nil
	}
	address, found := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("No ICA address associated with connection %s and port %s", controllerConnectionId, portID))
		return nil
	}
	chainId, err := k.GetChainIdFromConnectionId(ctx, controllerConnectionId)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("Found matching address for chain: %s, address %s, port %s", chainId, address, portID))

	// Check if the chainId matches one of the host zones, and if so,
	// store the relevant ICA address on the host zone struct
	if err := k.StoreHostZoneIcaAddresses(ctx, chainId, portID, address); err != nil {
		return err
	}

	return nil
}

// Checks if the chainId matches a given host zone, and the address matches a relevant ICA account
// If so, stores the ICA address on the host zone struct
// Also whitelists ICA addresses from rate limiting
func (k Keeper) StoreHostZoneIcaAddresses(ctx sdk.Context, chainId, portId, address string) error {
	// Check if the chainId matches a host zone
	// If the chainId does not match (for instance, a reward zone in a trade route is not a host zone)
	// then we can ignore the ICA address checks
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("chainId %s has no associated host zone", chainId))
		return nil
	}

	// expected port IDs for each ICA account type
	withdrawalOwner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	withdrawalPortID, err := icatypes.NewControllerPortID(withdrawalOwner)
	if err != nil {
		return err
	}
	feeOwner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_FEE)
	feePortID, err := icatypes.NewControllerPortID(feeOwner)
	if err != nil {
		return err
	}
	delegationOwner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_DELEGATION)
	delegationPortID, err := icatypes.NewControllerPortID(delegationOwner)
	if err != nil {
		return err
	}
	rewardOwner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_REDEMPTION)
	redemptionPortID, err := icatypes.NewControllerPortID(rewardOwner)
	if err != nil {
		return err
	}
	communityPoolDepositOwner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	communityPoolDepositPortID, err := icatypes.NewControllerPortID(communityPoolDepositOwner)
	if err != nil {
		return err
	}
	communityPoolReturnOwner := types.FormatICAAccountOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_RETURN)
	communityPoolReturnPortID, err := icatypes.NewControllerPortID(communityPoolReturnOwner)
	if err != nil {
		return err
	}

	// Set ICA account addresses
	switch {
	case portId == withdrawalPortID:
		hostZone.WithdrawalIcaAddress = address
	case portId == feePortID:
		hostZone.FeeIcaAddress = address
	case portId == delegationPortID:
		hostZone.DelegationIcaAddress = address
	case portId == redemptionPortID:
		hostZone.RedemptionIcaAddress = address
	case portId == communityPoolDepositPortID:
		hostZone.CommunityPoolDepositIcaAddress = address
	case portId == communityPoolReturnPortID:
		hostZone.CommunityPoolReturnIcaAddress = address
	default:
		k.Logger(ctx).Info(fmt.Sprintf("chainId %s has an associated host zone, but does not match any ICA accounts", chainId))
		return nil
	}

	k.SetHostZone(ctx, hostZone)

	// Once the delegation channel is registered, whitelist epochly transfers so they're not rate limited
	// Epochly transfers go from the deposit address to the delegation address
	if portId == delegationPortID {
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.DepositAddress,
			Receiver: hostZone.DelegationIcaAddress,
		})
	}

	// Once the fee channel is registered, whitelist reward transfers so they're not rate limited
	// Reward transfers go from the fee address to the reward collector
	if portId == feePortID {
		rewardCollectorAddress := k.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.FeeIcaAddress,
			Receiver: rewardCollectorAddress.String(),
		})
	}

	// Once the community pool deposit ICA is registered, whitelist epochly community pool transfers
	// from the deposit ICA to the community pool holding accounts
	if portId == communityPoolDepositPortID {
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.CommunityPoolDepositIcaAddress,
			Receiver: hostZone.CommunityPoolStakeHoldingAddress,
		})
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.CommunityPoolDepositIcaAddress,
			Receiver: hostZone.CommunityPoolRedeemHoldingAddress,
		})
	}

	// Once the community pool return ICA is registered, whitelist epochly community pool transfers
	// from the community pool stake holding account to the community pool return ICA
	if portId == communityPoolReturnPortID {
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.CommunityPoolStakeHoldingAddress,
			Receiver: hostZone.CommunityPoolReturnIcaAddress,
		})
	}

	return nil
}
