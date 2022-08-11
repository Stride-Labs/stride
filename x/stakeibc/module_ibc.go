package stakeibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// IBCModule implements the ICS26 interface for interchain accounts controller chains
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

func (im IBCModule) Hooks() keeper.Hooks {
	return im.keeper.Hooks()
}


func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	// Note: The channel capability must be claimed by the authentication module in OnChanOpenInit otherwise the
	// authentication module will not be able to send packets on the channel created for the associated interchain account.
	if err := im.keeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}
	return nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	controllerConnectionId, err := im.keeper.GetConnectionId(ctx, portID)
	if err != nil {
		ctx.Logger().Error("Unable to get connection for port " + portID)
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
		ctx.Logger().Error(
			"Unable to obtain counterparty chain for given connection and port",
			"ConnectionID", controllerConnectionId,
			"PortID", portID,
			"Error", err,
		)
		return nil
	}
	//  get zone info
	zoneInfo, found := im.keeper.GetHostZone(ctx, hostChainId)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Expected to find zone info for %v", hostChainId))
		return nil
	}
	ctx.Logger().Info("Found matching address", "chain", zoneInfo.ChainId, "address", address, "port", portID)

	// addresses
	withdrawalAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_WITHDRAWAL))
	if err != nil {
		return err
	}
	feeAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_FEE))
	if err != nil {
		return err
	}
	delegationAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_DELEGATION))
	if err != nil {
		return err
	}
	redemptionAddress, err := icatypes.NewControllerPortID(types.FormatICAAccountOwner(hostChainId, types.ICAAccountType_REDEMPTION))
	if err != nil {
		return err
	}

	// Set ICA account addresses
	switch {
	// withdrawal address
	case portID == withdrawalAddress:
		zoneInfo.WithdrawalAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_WITHDRAWAL}
	// fee address
	case portID == feeAddress:
		zoneInfo.FeeAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_FEE}
	// delegation address
	case portID == delegationAddress:
		zoneInfo.DelegationAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_DELEGATION}
	case portID == redemptionAddress:
		zoneInfo.RedemptionAccount = &types.ICAAccount{Address: address, Target: types.ICAAccountType_REDEMPTION}
	default:
		ctx.Logger().Error("Missing portId: ", portID)
	}

	im.keeper.SetHostZone(ctx, zoneInfo)
	return nil
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	im.keeper.Logger(ctx).Info("HANDLING ACK")
	err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, acknowledgement)
	if err != nil {
		return err
	}
	return im.keeper.HandleAcknowledgement(ctx, modulePacket, acknowledgement)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, []byte{})
	if err != nil {
		return err
	}
	return nil
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

func (im IBCModule) NegotiateAppVersion(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionID string,
	portID string,
	counterparty channeltypes.Counterparty,
	proposedVersion string,
) (version string, err error) {
	return proposedVersion, nil
}


// ###################################################################################
// 	Required functions to satisfy interface but not implemented for ICA auth modules
// ###################################################################################

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	panic("UNIMPLEMENTED")
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	panic("UNIMPLEMENTED")
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	panic("UNIMPLEMENTED")
}
