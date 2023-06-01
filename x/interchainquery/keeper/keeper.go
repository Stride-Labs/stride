package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/interchainquery/types"
)

// Keeper of this module maintains collections of registered zones.
type Keeper struct {
	cdc       codec.Codec
	storeKey  storetypes.StoreKey
	callbacks map[string]types.QueryCallbacks
	IBCKeeper *ibckeeper.Keeper
}

// NewKeeper returns a new instance of zones Keeper
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, ibckeeper *ibckeeper.Keeper) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		callbacks: make(map[string]types.QueryCallbacks),
		IBCKeeper: ibckeeper,
	}
}

func (k *Keeper) SetCallbackHandler(module string, handler types.QueryCallbacks) error {
	_, found := k.callbacks[module]
	if found {
		return fmt.Errorf("callback handler already set for %s", module)
	}
	k.callbacks[module] = handler.RegisterICQCallbacks()
	return nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) MakeRequest(ctx sdk.Context, module string, callbackId string, chainId string, connectionId string, queryType string, request []byte, ttl uint64) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId,
		"Submitting ICQ Request - module=%s, callbackId=%s, connectionId=%s, queryType=%s, ttl=%d", module, callbackId, connectionId, queryType, ttl))

	// Confirm the connectionId and chainId are valid
	if connectionId == "" {
		errMsg := "[ICQ Validation Check] Failed! connection id cannot be empty"
		k.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}
	if !strings.HasPrefix(connectionId, "connection") {
		errMsg := "[ICQ Validation Check] Failed! connection id must begin with 'connection'"
		k.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}
	if chainId == "" {
		errMsg := "[ICQ Validation Check] Failed! chain_id cannot be empty"
		k.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	// Confirm the module and callbackId exist
	if module != "" {
		if _, exists := k.callbacks[module]; !exists {
			err := fmt.Errorf("no callback handler registered for module %s", module)
			k.Logger(ctx).Error(err.Error())
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no callback handler registered for module")
		}
		if exists := k.callbacks[module].HasICQCallback(callbackId); !exists {
			err := fmt.Errorf("no callback %s registered for module %s", callbackId, module)
			k.Logger(ctx).Error(err.Error())
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "no callback handler registered for module")
		}
	}

	// Save the query to the store
	// If the same query is re-requested, it will get replace in the store with an updated TTL
	//  and the RequestSent bool reset to false
	query := k.NewQuery(ctx, module, callbackId, chainId, connectionId, queryType, request, ttl)
	k.SetQuery(ctx, *query)

	return nil
}
