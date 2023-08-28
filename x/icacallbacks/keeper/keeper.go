package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/Stride-Labs/stride/v14/x/icacallbacks/types"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeKey     storetypes.StoreKey
		memKey       storetypes.StoreKey
		paramstore   paramtypes.Subspace
		icacallbacks map[string]types.ICACallback
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
		icacallbacks: make(map[string]types.ICACallback),
		IBCKeeper:    ibcKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetICACallbacks(moduleCallbacks ...types.ModuleCallbacks) error {
	for _, callbacks := range moduleCallbacks {
		for _, callback := range callbacks {
			if _, found := k.icacallbacks[callback.CallbackId]; found {
				return fmt.Errorf("callback for ID %s already registered", callback.CallbackId)
			}
			k.icacallbacks[callback.CallbackId] = callback
		}
	}
	return nil
}

func (k Keeper) CallRegisteredICACallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *types.AcknowledgementResponse) error {
	// Get the callback key and associated callback data from the packet
	callbackDataKey := types.PacketID(packet.GetSourcePort(), packet.GetSourceChannel(), packet.Sequence)
	callbackData, found := k.GetCallbackData(ctx, callbackDataKey)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("callback data not found for portID: %s, channelID: %s, sequence: %d",
			packet.SourcePort, packet.SourceChannel, packet.Sequence))
		return nil
	}

	// If there's an associated callback function, execute it
	callback, found := k.icacallbacks[callbackData.CallbackId]
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("No associated callback with callback data %v", callbackData))
		return nil
	}
	if err := callback.CallbackFunc(ctx, packet, ackResponse, callbackData.CallbackArgs); err != nil {
		errMsg := fmt.Sprintf("Error occured while calling ICACallback (%s) | err: %s", callbackData.CallbackId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(types.ErrCallbackFailed, errMsg)
	}

	// remove the callback data
	k.RemoveCallbackData(ctx, callbackDataKey)
	return nil
}
