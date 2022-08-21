package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"

	"github.com/Stride-Labs/stride/x/icacallbacks/types"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
)

type (
	Keeper struct {
		cdc                 codec.BinaryCodec
		storeKey            sdk.StoreKey
		memKey              sdk.StoreKey
		paramstore          paramtypes.Subspace
		scopedKeeper        capabilitykeeper.ScopedKeeper
		icacallbacks        map[string]types.ICACallbackHandler
		IBCKeeper           ibckeeper.Keeper
		ICAControllerKeeper icacontrollerkeeper.Keeper
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
		cdc:                 cdc,
		storeKey:            storeKey,
		memKey:              memKey,
		paramstore:          ps,
		scopedKeeper:        scopedKeeper,
		icacallbacks:        make(map[string]types.ICACallbackHandler),
		IBCKeeper:           ibcKeeper,
		ICAControllerKeeper: icacontrollerkeeper,
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

func (k Keeper) CallRegisteredICACallback(ctx sdk.Context, modulePacket channeltypes.Packet, txMsgData *sdk.TxMsgData) error {
	// get the relevant module from the channel and port
	portID := modulePacket.GetSourcePort()
	channelID := modulePacket.GetSourceChannel()
	module, _, err := k.IBCKeeper.ChannelKeeper.LookupModuleByChannel(ctx, portID, channelID)
	if err != nil {
		return err
	}
	// fetch the callback data
	callbackDataKey := types.PacketID(portID, channelID, modulePacket.Sequence)
	callbackData, found := k.GetCallbackData(ctx, callbackDataKey)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("callback data not found for portID: %s, channelID: %s, sequence: %d", portID, channelID, modulePacket.Sequence))
		return nil
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("callback data found for portID: %s, channelID: %s, sequence: %d", portID, channelID, modulePacket.Sequence))
	}

	// fetch the callback function
	callbackHandler, err := k.GetICACallbackHandler(module)
	if err != nil {
		errMsg := fmt.Sprintf("Callback handler does not exist for module %s | err: %s", module, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrCallbackHandlerNotFound, errMsg)
	}

	// call the callback
	if callbackHandler.HasICACallback(callbackData.CallbackId) {
		// if acknowledgement is empty, then it is a timeout
		err := callbackHandler.CallICACallback(ctx, callbackData.CallbackId, modulePacket, txMsgData, callbackData.CallbackArgs)
		if err != nil {
			errMsg := fmt.Sprintf("Error occured while calling ICACallback (%s) | err: %s", callbackData.CallbackId, err.Error())
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(types.ErrCallbackFailed, errMsg)
		}
	} else {
		k.Logger(ctx).Error(fmt.Sprintf("Callback %v has no associated callback", callbackData))
	}
	// QUESTION: Do we want to catch the case where the callback ID has not been registered?
	// Maybe just as an info log if it's expected that some acks do not have an associated callback?

	// remove the callback data
	k.RemoveCallbackData(ctx, callbackDataKey)
	return nil
}
