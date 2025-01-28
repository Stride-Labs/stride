package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/spf13/cast"

	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"

	"github.com/Stride-Labs/stride/v25/x/stakeibc/types"
)

func (k Keeper) OnChanOpenAck(ctx sdk.Context, portID, channelID string) error {
	// Lookup connection ID, counterparty chain ID, and ICA address from the channel ID
	controllerConnectionId, _, err := k.IBCKeeper.ChannelKeeper.GetChannelConnection(ctx, portID, channelID)
	if err != nil {
		return err
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
	if err := k.StoreHostZoneIcaAddress(ctx, chainId, portID, address); err != nil {
		return err
	}

	// Check if the chainId matches any ICAs from trade routes, and if so,
	// store the relevant ICA addresses in the trade route structs
	if err := k.StoreTradeRouteIcaAddress(ctx, chainId, portID, address); err != nil {
		return err
	}

	return nil
}

// Checks if the chainId matches a given host zone, and the address matches a relevant ICA account
// If so, stores the ICA address on the host zone struct
// Also whitelists ICA addresses from rate limiting
func (k Keeper) StoreHostZoneIcaAddress(ctx sdk.Context, chainId, portId, address string) error {
	// Check if the chainId matches a host zone
	// If the chainId does not match (for instance, a reward zone in a trade route is not a host zone)
	// then we can ignore the ICA address checks
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("chainId %s has no associated host zone", chainId))
		return nil
	}

	// expected port IDs for each ICA account type
	delegationOwner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_DELEGATION)
	delegationPortID, err := icatypes.NewControllerPortID(delegationOwner)
	if err != nil {
		return err
	}
	withdrawalOwner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_WITHDRAWAL)
	withdrawalPortID, err := icatypes.NewControllerPortID(withdrawalOwner)
	if err != nil {
		return err
	}
	feeOwner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_FEE)
	feePortID, err := icatypes.NewControllerPortID(feeOwner)
	if err != nil {
		return err
	}
	redemptionOwner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_REDEMPTION)
	redemptionPortID, err := icatypes.NewControllerPortID(redemptionOwner)
	if err != nil {
		return err
	}
	communityPoolDepositOwner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT)
	communityPoolDepositPortID, err := icatypes.NewControllerPortID(communityPoolDepositOwner)
	if err != nil {
		return err
	}
	communityPoolReturnOwner := types.FormatHostZoneICAOwner(chainId, types.ICAAccountType_COMMUNITY_POOL_RETURN)
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
		k.Logger(ctx).Info(fmt.Sprintf("portId %s has an associated host zone, but does not match any ICA accounts", portId))
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

// Checks if the port matches an ICA account on the trade route, and if so, stores the
// relevant ICA address on the trade route
func (k Keeper) StoreTradeRouteIcaAddress(ctx sdk.Context, callbackChainId, callbackPortId, address string) error {
	// Check if the port Id matches either the trade or unwind ICA on the tradeRoute
	// If the chainId and port Id from the callback match the account
	// on a trade route, set the ICA address in the relevant places,
	// including the from/to addresses on each hop
	for _, route := range k.GetAllTradeRoutes(ctx) {
		// Build the expected port ID for the reward and trade accounts,
		// using the chainId and route ID
		rewardAccount := route.RewardAccount
		rewardOwner := types.FormatTradeRouteICAOwnerFromRouteId(rewardAccount.ChainId, route.GetRouteId(), rewardAccount.Type)
		rewardPortId, err := icatypes.NewControllerPortID(rewardOwner)
		if err != nil {
			return err
		}

		tradeAccount := route.TradeAccount
		tradeOwner := types.FormatTradeRouteICAOwnerFromRouteId(tradeAccount.ChainId, route.GetRouteId(), tradeAccount.Type)
		tradePortId, err := icatypes.NewControllerPortID(tradeOwner)
		if err != nil {
			return err
		}

		// Check if route IDs match the callback chainId/portId
		if route.RewardAccount.ChainId == callbackChainId && callbackPortId == rewardPortId {
			k.Logger(ctx).Info(fmt.Sprintf("ICA Address %s found for Unwind ICA on %s", address, route.Description()))
			route.RewardAccount.Address = address

		} else if route.TradeAccount.ChainId == callbackChainId && callbackPortId == tradePortId {
			k.Logger(ctx).Info(fmt.Sprintf("ICA Address %s found for Trade ICA on %s", address, route.Description()))
			route.TradeAccount.Address = address
		}

		k.SetTradeRoute(ctx, route)
	}

	return nil
}

// Given a connection ID, returns the light client time
func (k Keeper) GetLightClientTime(ctx sdk.Context, connectionID string) (clientTime uint64, err error) {
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return 0, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection-id: %s", connectionID)
	}

	latestConsensusClientState, found := k.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(ctx, connection.ClientId)
	if !found {
		return 0, errorsmod.Wrapf(clienttypes.ErrConsensusStateNotFound, "client-id: %s", connection.ClientId)
	}

	return latestConsensusClientState.GetTimestamp(), nil
}

// Given a connection ID, returns the light client height
func (k Keeper) GetLightClientHeight(ctx sdk.Context, connectionID string) (height uint64, err error) {
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return 0, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection-id: %s", connectionID)
	}

	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return 0, errorsmod.Wrapf(clienttypes.ErrConsensusStateNotFound, "client-id: %s", connection.ClientId)
	}

	latestHeight, err := cast.ToUint64E(clientState.GetLatestHeight().GetRevisionHeight())
	if err != nil {
		return 0, err
	}
	return latestHeight, nil
}

// Lookup a chain ID from a connection ID by looking up the client state
func (k Keeper) GetChainIdFromConnectionId(ctx sdk.Context, connectionID string) (string, error) {
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return "", errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection %s not found", connectionID)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return "", errorsmod.Wrapf(clienttypes.ErrClientNotFound, "client %s not found", connection.ClientId)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", types.ErrClientStateNotTendermint
	}

	return client.ChainId, nil
}
