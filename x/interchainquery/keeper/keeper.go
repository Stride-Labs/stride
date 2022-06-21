package keeper

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of this module maintains collections of registered zones.
type Keeper struct {
	cdc       codec.Codec
	storeKey  sdk.StoreKey
	callbacks map[string]types.QueryCallbacks
}

// NewKeeper returns a new instance of zones Keeper
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		callbacks: make(map[string]types.QueryCallbacks),
	}
}

func (k *Keeper) SetCallbackHandler(module string, handler types.QueryCallbacks) error {
	_, found := k.callbacks[module]
	if found {
		return fmt.Errorf("callback handler already set for %s", module)
	}
	k.callbacks[module] = handler
	return nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) SetDatapointForId(ctx sdk.Context, id string, result []byte, height sdk.Int) error {
	mapping := types.DataPoint{Id: id, RemoteHeight: height, LocalHeight: sdk.NewInt(ctx.BlockHeight()), Value: result}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixData)
	bz := k.cdc.MustMarshal(&mapping)
	store.Set([]byte(id), bz)
	return nil
}

func (k *Keeper) GetDatapointForId(ctx sdk.Context, id string) (types.DataPoint, error) {
	mapping := types.DataPoint{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixData)
	bz := store.Get([]byte(id))
	if len(bz) == 0 {
		return types.DataPoint{}, fmt.Errorf("unable to find data for id %s", id)
	}

	k.cdc.MustUnmarshal(bz, &mapping)
	return mapping, nil
}

func (k *Keeper) GetDatapoint(ctx sdk.Context, connection_id string, chain_id string, query_type string, request []byte, height int64) (types.DataPoint, error) {
	id := GenerateQueryHash(connection_id, chain_id, query_type, request, strconv.FormatInt(height, 10))
	return k.GetDatapointForId(ctx, id)
}

func (k *Keeper) GetDatapointOrRequest(ctx sdk.Context, connection_id string, chain_id string, query_type string, request []byte, max_age int64, height int64) (types.DataPoint, error) {
	val, err := k.GetDatapoint(ctx, connection_id, chain_id, query_type, request, height)
	if err != nil {
		// no datapoint
		k.MakeRequest(ctx, connection_id, chain_id, query_type, request, sdk.NewInt(-1), strconv.FormatInt(height, 10), "", nil)
		return types.DataPoint{}, fmt.Errorf("no data; query submitted")
	}

	if val.LocalHeight.LT(sdk.NewInt(ctx.BlockHeight() - max_age)) { // this is somewhat arbitrary; TODO: make this better
		k.MakeRequest(ctx, connection_id, chain_id, query_type, request, sdk.NewInt(-1), strconv.FormatInt(height, 10), "", nil)
		return types.DataPoint{}, fmt.Errorf("stale data; query submitted")
	}
	// check ttl
	return val, nil
}

func (k *Keeper) MakeRequest(ctx sdk.Context, connection_id string, chain_id string, query_type string, request []byte, period sdk.Int, height string, module string, callback interface{}) {
	key := GenerateQueryHash(connection_id, chain_id, query_type, request, height)
	_, found := k.GetQuery(ctx, key)
	if !found {
		if module != "" {
			k.callbacks[module].AddCallback(key, callback)
		}
		newQuery := k.NewQuery(ctx, connection_id, chain_id, query_type, request, period, height)
		k.SetQuery(ctx, *newQuery)
	}
}

func (k Keeper) QueryHostZone(ctx sdk.Context, zone stakeibctypes.HostZone, cb Callback, query_type string, address string, height int64) error {

	// note: height=0 queries at latest block header, NOT at height 0
	connectionId := zone.ConnectionId
	chainId := zone.ChainId
	heightStr := strconv.FormatInt(height, 10)

	// populate query based on query type
	var bz []byte
	if query_type == "cosmos.bank.v1beta1.Query/AllBalances" {
		query := banktypes.QueryAllBalancesRequest{Address: address}
		bz = k.cdc.MustMarshal(&query)
	} else if query_type == "cosmos.staking.v1beta1.Query/DelegatorDelegations" {
		query := stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: address}
		bz = k.cdc.MustMarshal(&query)
	}
	k.Logger(ctx).Info(fmt.Sprintf("\tabout to %s to %s at height %s", query_type, address, heightStr))

	k.MakeRequest(
		ctx,
		connectionId,
		chainId,
		query_type,
		bz,
		// TODO(TEST-79) understand and use proper period
		sdk.NewInt(100000),
		strconv.FormatInt(height, 10),
		types.ModuleName,
		cb,
	)

	// TODO(TEST-119) get gaia LC height, pass to height

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
			sdk.NewAttribute(types.AttributeKeyQueryId, GenerateQueryHash(connectionId, chainId, query_type, bz, heightStr)),
			sdk.NewAttribute(types.AttributeKeyChainId, chainId),
			sdk.NewAttribute(types.AttributeKeyConnectionId, connectionId),
			sdk.NewAttribute(types.AttributeKeyType, query_type),
			// TODO(TEST-119) set height based on gaia LC height
			sdk.NewAttribute(types.AttributeKeyHeight, heightStr),
			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(bz)),
		),
	})
	return nil
}
