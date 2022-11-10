package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
)

// Keeper of this module maintains collections of registered zones.
type Keeper struct {
	cdc       codec.Codec
	storeKey  sdk.StoreKey
	callbacks map[string]types.QueryCallbacks
	IBCKeeper *ibckeeper.Keeper
}

// NewKeeper returns a new instance of zones Keeper
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, ibckeeper *ibckeeper.Keeper) Keeper {
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

func (k *Keeper) MakeRequest(ctx sdk.Context, connectionId string, chainId string, queryType string, request []byte, period sdk.Int, module string, callbackId string, ttl uint64, height int64) error {
	k.Logger(ctx).Info(
		"MakeRequest",
		"connection_id", connectionId,
		"chain_id", chainId,
		"query_type", queryType,
		"request", request,
		"period", period,
		"module", module,
		"callback", callbackId,
		"ttl", ttl,
		"height", height,
	)

	// Only 0 height queries are currently supported
	if height != 0 {
		k.Logger(ctx).Error("[ICQ Validation Check] Failed! height for interchainquery must be 0 (we exclusively query at the latest height on the host zone)")
	}

	// Confirm the connectionId and chainId are valid
	if connectionId == "" {
		k.Logger(ctx).Error("[ICQ Validation Check] Failed! connection id cannot be empty")
	}
	if !strings.HasPrefix(connectionId, "connection") {
		k.Logger(ctx).Error("[ICQ Validation Check] Failed! connection id must begin with 'connection'")
	}
	if chainId == "" {
		k.Logger(ctx).Error("[ICQ Validation Check] Failed! chain_id cannot be empty")
	}

	// Check to see if the query already exists
	key := GenerateQueryHash(connectionId, chainId, queryType, request, module, height)
	existingQuery, found := k.GetQuery(ctx, key)

	// If the same query is re-requested - reset the TTL
	if found {
		existingQuery.Ttl = ttl
		k.SetQuery(ctx, existingQuery)
		return nil
	}

	// Otherwise, if it's a new query, add it to the store
	if module != "" {
		if _, exists := k.callbacks[module]; !exists {
			err := fmt.Errorf("no callback handler registered for module %s", module)
			k.Logger(ctx).Error(err.Error())
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "no callback handler registered for module")
		}
		if exists := k.callbacks[module].Has(callbackId); !exists {
			err := fmt.Errorf("no callback %s registered for module %s", callbackId, module)
			k.Logger(ctx).Error(err.Error())
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "no callback handler registered for module")
		}
	}
	newQuery := k.NewQuery(ctx, module, connectionId, chainId, queryType, request, period, callbackId, ttl, height)
	k.SetQuery(ctx, *newQuery)

	return nil
}
