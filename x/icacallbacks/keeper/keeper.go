package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/Stride-Labs/stride/v9/x/icacallbacks/types"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeKey     storetypes.StoreKey
		memKey       storetypes.StoreKey
		paramstore   paramtypes.Subspace
		icacallbacks map[string]types.ICACallbackHandler
		IBCKeeper    ibckeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	ibcKeeper ibckeeper.Keeper,
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
		icacallbacks: make(map[string]types.ICACallbackHandler),
		IBCKeeper:    ibcKeeper,
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

func (k Keeper) GetCallbackDataFromPacket(ctx sdk.Context, modulePacket channeltypes.Packet, callbackDataKey string) (cbd *types.CallbackData, found bool) {
	// get the relevant module from the channel and port
	portID := modulePacket.GetSourcePort()
	channelID := modulePacket.GetSourceChannel()
	// fetch the callback data
	callbackData, found := k.GetCallbackData(ctx, callbackDataKey)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("callback data not found for portID: %s, channelID: %s, sequence: %d", portID, channelID, modulePacket.Sequence))
		return nil, false
	} else {
		k.Logger(ctx).Info(fmt.Sprintf("callback data found for portID: %s, channelID: %s, sequence: %d", portID, channelID, modulePacket.Sequence))
	}
	return &callbackData, true
}

func (k Keeper) GetICACallbackHandlerFromPacket(ctx sdk.Context, modulePacket channeltypes.Packet) (*types.ICACallbackHandler, error) {
	module, _, err := k.IBCKeeper.ChannelKeeper.LookupModuleByChannel(ctx, modulePacket.GetSourcePort(), modulePacket.GetSourceChannel())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("error LookupModuleByChannel for portID: %s, channelID: %s, sequence: %d", modulePacket.GetSourcePort(), modulePacket.GetSourceChannel(), modulePacket.Sequence))
		return nil, err
	}
	// fetch the callback function
	callbackHandler, err := k.GetICACallbackHandler(module)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrCallbackHandlerNotFound, "Callback handler does not exist for module %s | err: %s", module, err.Error())
	}
	return &callbackHandler, nil
}

func (k Keeper) CallRegisteredICACallback(ctx sdk.Context, modulePacket channeltypes.Packet, ackResponse *types.AcknowledgementResponse) error {
	callbackDataKey := types.PacketID(modulePacket.GetSourcePort(), modulePacket.GetSourceChannel(), modulePacket.Sequence)
	callbackData, found := k.GetCallbackDataFromPacket(ctx, modulePacket, callbackDataKey)
	if !found {
		return nil
	}
	callbackHandler, err := k.GetICACallbackHandlerFromPacket(ctx, modulePacket)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("GetICACallbackHandlerFromPacket %s", err.Error()))
		return err
	}

	// call the callback
	if (*callbackHandler).HasICACallback(callbackData.CallbackId) {
		k.Logger(ctx).Info(fmt.Sprintf("Calling callback for %s", callbackData.CallbackId))
		// if acknowledgement is empty, then it is a timeout
		err := (*callbackHandler).CallICACallback(ctx, callbackData.CallbackId, modulePacket, ackResponse, callbackData.CallbackArgs)
		if err != nil {
			errMsg := fmt.Sprintf("Error occured while calling ICACallback (%s) | err: %s", callbackData.CallbackId, err.Error())
			k.Logger(ctx).Error(errMsg)
			return errorsmod.Wrapf(types.ErrCallbackFailed, errMsg)
		}
	} else {
		k.Logger(ctx).Error(fmt.Sprintf("Callback %v has no associated callback", callbackData))
	}

	// remove the callback data
	k.RemoveCallbackData(ctx, callbackDataKey)
	return nil
}
