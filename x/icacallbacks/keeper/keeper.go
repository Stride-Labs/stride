package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"

	"github.com/Stride-Labs/stride/x/icacallbacks/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"

	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeKey     sdk.StoreKey
		memKey       sdk.StoreKey
		paramstore   paramtypes.Subspace
		scopedKeeper capabilitykeeper.ScopedKeeper
		icacallbacks map[string]types.ICACallbackHandler
		IBCKeeper    ibckeeper.Keeper
		ICAControllerKeeper   icacontrollerkeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	ibcKeeper ibckeeper.Keeper,
	icacontrollerkeeper icacontrollerkeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		memKey:       memKey,
		paramstore:   ps,
		scopedKeeper: scopedKeeper,
		icacallbacks: make(map[string]types.ICACallbackHandler),
		IBCKeeper: ibcKeeper,
		ICAControllerKeeper:   icacontrollerkeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Should we add a `AddICACallback`
func (k *Keeper) SetICACallbackHandler(module string, handler types.ICACallbackHandler) error {
	_, found := k.icacallbacks[module]
	if found {
		return fmt.Errorf("callback handler already set for %s", module)
	}
	k.icacallbacks[module] = handler.RegisterICACallbacks()
	return nil
}

func (k *Keeper) GetICACallbackHandler(module string) (types.ICACallbackHandler, error) {
	callback, found := k.icacallbacks[module]
	if !found {
		return nil, fmt.Errorf("no callback handler found for %s", module)
	}
	return callback, nil
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) SubmitICATx(ctx sdk.Context, connectionId string, msgs []sdk.Msg, account stakeibctypes.ICAAccount, timeoutTimestamp uint64, chainId string, callbackId string, callbackArgs []byte) (uint64, error) {
	owner := stakeibctypes.FormatICAAccountOwner(chainId, account.GetTarget())
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return 0, err
	}

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, connectionId, portID)
	if !found {
		return 0, sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return 0, err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	
	sequence, err := k.ICAControllerKeeper.SendTx(ctx, chanCap, connectionId, portID, packetData, timeoutTimestamp)
	if err != nil {
		return 0, err
	}

	callback := types.CallbackData{
		CallbackKey: callbackId,
		PortId: portID,
		ChannelId: channelID,
		Sequence: sequence,
		CallbackId: callbackId,
		CallbackArgs: callbackArgs,
	}
	k.SetCallbackData(ctx, callback)

	return sequence, nil
}

func (k Keeper) CallRegisteredICACallback(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte) error {
	k.Logger(ctx).Info("CallRegisteredICACallback", "dst portID", modulePacket.GetDestPort())
	k.Logger(ctx).Info("CallRegisteredICACallback", "dst channelID", modulePacket.GetDestChannel())
	// get the relevant module from the channel and port
	portID := modulePacket.GetSourcePort()
	k.Logger(ctx).Info("CallRegisteredICACallback", "portID", portID)
	channelID := modulePacket.GetSourceChannel()
	k.Logger(ctx).Info("CallRegisteredICACallback", "channelID", channelID)
	module, _, err := k.IBCKeeper.ChannelKeeper.LookupModuleByChannel(ctx, portID, channelID)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info("CallRegisteredICACallback", "module", module)
	// fetch the callback data
	callbackDataKey := types.PacketID(portID, channelID, modulePacket.Sequence)
	k.Logger(ctx).Info("CallRegisteredICACallback", "callbackDataKey", callbackDataKey)
	callbackData, found := k.GetCallbackData(ctx, callbackDataKey)
	if !found {
		errMsg := fmt.Sprintf("callback data not found for portID: %s, channelID: %s, sequence: %d", portID, channelID, modulePacket.Sequence)
		k.Logger(ctx).Info(errMsg)
		return nil
	}
	k.Logger(ctx).Info("CallRegisteredICACallback", "callbackData", callbackData)
	// fetch the callback function
	callbackHandler, err := k.GetICACallbackHandler(module)
	if err != nil {
		k.Logger(ctx).Info("CallRegisteredICACallback", "err", err)
		return err
	}
	k.Logger(ctx).Info("CallRegisteredICACallback", "callbackHandler", callbackHandler)
	// call the callback
	if callbackHandler.HasICACallback(callbackData.CallbackId) {
		// if acknowledgement is empty, then it is a timeout
		err := callbackHandler.CallICACallback(ctx, callbackData.CallbackId, modulePacket, acknowledgement, callbackData.CallbackArgs)
		if err != nil {
			k.Logger(ctx).Info("CallRegisteredICACallback", "err", err)
			return err
		}
	}
	k.Logger(ctx).Info("CallRegisteredICACallback - HasICACallback")

	// remove the callback data
	// NOTE: Should we remove the callback data here, or above (conditional on HasICACallback == true)?
	k.RemoveCallbackData(ctx, callbackDataKey)
	k.Logger(ctx).Info("CallRegisteredICACallback - RemoveCallbackData")
	return nil
}
