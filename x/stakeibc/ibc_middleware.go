package stakeibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	ratelimittypes "github.com/Stride-Labs/stride/v14/x/ratelimit/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

var _ porttypes.Middleware = &IBCMiddleware{}

type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
}

func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// OnChanOpenInit simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck stores the new ICA acccount addresses on the host zone and then passes to the next middleware stack
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck: portID %s, channelID %s, counterpartyChannelID %s, counterpartyVersion %s", portID, channelID, counterpartyChannelID, counterpartyVersion))
	controllerConnectionId, err := im.keeper.GetConnectionId(ctx, portID)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to get connection for port: %s", portID))
	}
	address, found := im.keeper.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find an address for %s/%s", controllerConnectionId, portID))
		return nil
	}
	// get host chain id from connection
	// fetch counterparty connection
	hostChainId, err := im.keeper.GetChainID(ctx, controllerConnectionId)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to obtain counterparty chain for connection: %s, port: %s, err: %s", controllerConnectionId, portID, err.Error()))
		return nil
	}
	//  get zone info
	zoneInfo, found := im.keeper.GetHostZone(ctx, hostChainId)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find zone info for %v", hostChainId))
		return nil
	}
	ctx.Logger().Info(fmt.Sprintf("Found matching address for chain: %s, address %s, port %s", zoneInfo.ChainId, address, portID))

	// expected port IDs for each ICA account type
	withdrawalPortID, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_WITHDRAWAL))
	if err != nil {
		return err
	}
	feePortID, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_FEE))
	if err != nil {
		return err
	}
	delegationPortID, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_DELEGATION))
	if err != nil {
		return err
	}
	redemptionPortID, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_REDEMPTION))
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
		zoneInfo.WithdrawalIcaAddress = address
	case portID == feePortID:
		zoneInfo.FeeIcaAddress = address
	case portID == delegationPortID:
		zoneInfo.DelegationIcaAddress = address
	case portID == redemptionPortID:
		zoneInfo.RedemptionIcaAddress = address
	case portID == communityPoolDepositPortID:
		zoneInfo.CommunityPoolDepositIcaAddress = address
	case portID == communityPoolReturnPortID:
		zoneInfo.CommunityPoolReturnIcaAddress = address
	default:
		ctx.Logger().Error(fmt.Sprintf("Missing portId: %s", portID))
	}

	// Once the delegation channel is registered, whitelist epochly transfers so they're not rate limited
	// Epochly transfers go from the deposit address to the delegation address
	if portID == delegationPortID {
		im.keeper.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   zoneInfo.DepositAddress,
			Receiver: zoneInfo.DelegationIcaAddress,
		})
	}

	// Once the fee channel is registered, whitelist reward transfers so they're not rate limited
	// Reward transfers go from the fee address to the reward collector
	if portID == feePortID {
		rewardCollectorAddress := im.keeper.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
		im.keeper.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   zoneInfo.FeeIcaAddress,
			Receiver: rewardCollectorAddress.String(),
		})
	}

	// Once the community pool deposit ICA is registered, whitelist epochly community pool transfers
	// from the deposit ICA to the community pool holding accounts
	if portID == communityPoolDepositPortID {
		im.keeper.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   zoneInfo.CommunityPoolDepositIcaAddress,
			Receiver: zoneInfo.CommunityPoolStakeHoldingAddress,
		})
		im.keeper.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   zoneInfo.CommunityPoolDepositIcaAddress,
			Receiver: zoneInfo.CommunityPoolRedeemHoldingAddress,
		})
	}

	// Once the community pool return ICA is registered, whitelist epochly community pool transfers
	// from the community pool stake holding account to the community pool return ICA
	if portID == communityPoolReturnPortID {
		im.keeper.RatelimitKeeper.SetWhitelistedAddressPair(ctx, ratelimittypes.WhitelistedAddressPair{
			Sender:   zoneInfo.CommunityPoolStakeHoldingAddress,
			Receiver: zoneInfo.CommunityPoolReturnIcaAddress,
		})
	}

	im.keeper.SetHostZone(ctx, zoneInfo)

	// call underlying app's OnChanOpenAck
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanCloseConfirm simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnChanCloseInit simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanOpenConfirm simply passes down the to next middleware stack
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnAcknowledgementPacket simply passes down the to next middleware stack
// The Ack handling and routing is managed by icacallbacks
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket simply passes down the to next middleware stack
// The Ack handling and routing is managed by icacallbacks
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// OnRecvPacket simply passes down the to next middleware stack
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// SendPacket implements the ICS4 Wrapper interface but is not utilized in the ICA stack
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	panic("UNIMPLEMENTED")
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
// but is not utilized in the bottom of ICA stack
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	panic("UNIMPLEMENTED")
}

// GetAppVersion implements the ICS4 Wrapper interface
// but is not utilized in the bottom of ICA stack
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	panic("UNIMPLEMENTED")
}
