package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	ratelimittypes "github.com/Stride-Labs/stride/v16/x/ratelimit/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (k Keeper) OnChanOpenAck(ctx sdk.Context, portID, channelID string) error {
	controllerConnectionId, found := k.GetConnectionIdFromICAPortId(ctx, portID)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("portId %s has no associated ICA account", portID))
		return nil
	}
	address, found := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("No ICA address associated with connection %s and port %s", controllerConnectionId, portID))
		return nil
	}

	// get host chain id from connection
	// fetch counterparty connection
	hostChainId, err := k.GetChainIdFromConnectionId(ctx, controllerConnectionId)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to obtain counterparty chain for connection: %s, port: %s, err: %s", controllerConnectionId, portID, err.Error()))
		return nil
	}
	//  get zone info
	hostZone, found := k.GetHostZone(ctx, hostChainId)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find zone info for %v", hostChainId))
		return nil
	}
	ctx.Logger().Info(fmt.Sprintf("Found matching address for chain: %s, address %s, port %s", hostZone.ChainId, address, portID))

	// expected port IDs for each ICA account type
	withdrawalOwner := types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_WITHDRAWAL)
	withdrawalPortID, err := icatypes.NewControllerPortID(withdrawalOwner)
	if err != nil {
		return err
	}
	feeOwner := types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_FEE)
	feePortID, err := icatypes.NewControllerPortID(feeOwner)
	if err != nil {
		return err
	}
	delegationOwner := types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_DELEGATION)
	delegationPortID, err := icatypes.NewControllerPortID(delegationOwner)
	if err != nil {
		return err
	}
	rewardOwner := types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_REDEMPTION)
	redemptionPortID, err := icatypes.NewControllerPortID(rewardOwner)
	if err != nil {
		return err
	}
	communityPoolDepositOwner := types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	communityPoolDepositPortID, err := icatypes.NewControllerPortID(communityPoolDepositOwner)
	if err != nil {
		return err
	}
	communityPoolReturnOwner := types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_COMMUNITY_POOL_RETURN)
	communityPoolReturnPortID, err := icatypes.NewControllerPortID(communityPoolReturnOwner)
	if err != nil {
		return err
	}

	// Set ICA account addresses
	switch {
	case portID == withdrawalPortID:
		hostZone.WithdrawalIcaAddress = address
	case portID == feePortID:
		hostZone.FeeIcaAddress = address
	case portID == delegationPortID:
		hostZone.DelegationIcaAddress = address
	case portID == redemptionPortID:
		hostZone.RedemptionIcaAddress = address
	case portID == communityPoolDepositPortID:
		hostZone.CommunityPoolDepositIcaAddress = address
	case portID == communityPoolReturnPortID:
		hostZone.CommunityPoolReturnIcaAddress = address
	default:
		ctx.Logger().Error(fmt.Sprintf("Missing portId: %s", portID))
	}

	k.SetHostZone(ctx, hostZone)

	// Once the delegation channel is registered, whitelist epochly transfers so they're not rate limited
	// Epochly transfers go from the deposit address to the delegation address
	if portID == delegationPortID {
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.DepositAddress,
			Receiver: hostZone.DelegationIcaAddress,
		})
	}

	// Once the fee channel is registered, whitelist reward transfers so they're not rate limited
	// Reward transfers go from the fee address to the reward collector
	if portID == feePortID {
		rewardCollectorAddress := k.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.FeeIcaAddress,
			Receiver: rewardCollectorAddress.String(),
		})
	}

	// Once the community pool deposit ICA is registered, whitelist epochly community pool transfers
	// from the deposit ICA to the community pool holding accounts
	if portID == communityPoolDepositPortID {
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
	if portID == communityPoolReturnPortID {
		k.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   hostZone.CommunityPoolStakeHoldingAddress,
			Receiver: hostZone.CommunityPoolReturnIcaAddress,
		})
	}

	return nil
}
