package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Definition for the store key format based on tradeRoute start and end denoms
func (k Keeper) GetTradeRouteKeyFromDenoms(ctx sdk.Context, startDenom string, endDenom string) (key []byte) {
	return []byte(startDenom + "-" + endDenom)
}

func (k Keeper) GetTradeRouteKey(ctx sdk.Context, tradeRoute types.TradeRoute) (key []byte) {
	routeStartDenom := tradeRoute.RewardDenomOnHostZone
	routeEndDenom := tradeRoute.TargetDenomOnHostZone
	return k.GetTradeRouteKeyFromDenoms(ctx, routeStartDenom, routeEndDenom)
}



// SetTradeRoute set a specific tradeRoute in the store
func (k Keeper) SetTradeRoute(ctx sdk.Context, tradeRoute types.TradeRoute) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TradeRouteKey))
	b := k.cdc.MustMarshal(&tradeRoute)
	store.Set(k.GetTradeRouteKey(ctx, tradeRoute), b)
}

// GetTradeRoute returns a tradeRoute from its start and end denoms
func (k Keeper) GetTradeRoute(ctx sdk.Context, startDenom string, endDenom string) (val types.TradeRoute, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TradeRouteKey))
	storeKey := k.GetTradeRouteKeyFromDenoms(ctx, startDenom, endDenom)
	b := store.Get(storeKey)
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
} 

// RemoveTradeRoute removes a tradeRoute from the store
func (k Keeper) RemoveTradeRoute(ctx sdk.Context, startDenom string, endDenom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TradeRouteKey))
	storeKey := k.GetTradeRouteKeyFromDenoms(ctx, startDenom, endDenom)	
	store.Delete(storeKey)
}

// GetAllTradeRoute returns all tradeRoutes
func (k Keeper) GetAllTradeRoute(ctx sdk.Context) (list []types.TradeRoute) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TradeRouteKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.TradeRoute
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetAllTradeRouteForHostZone uses that targetDenomOnHostZone should be the hostDenom for the chain
func (k Keeper) GetAllTradeRouteForHostZone(ctx sdk.Context, chainId string) (list []types.TradeRoute) {
	tradeRoutes := k.GetAllTradeRoute(ctx)
	for _, tradeRoute := range tradeRoutes {
		hostZone, err := k.GetHostZoneFromHostDenom(ctx, tradeRoute.TargetDenomOnHostZone)
		if err == nil && hostZone.ChainId == chainId {
			// Marshal and unmarshal to "deep copy" since subfields contain pointers
			var copied types.TradeRoute
			value := k.cdc.MustMarshal(&tradeRoute)
			k.cdc.MustUnmarshal(value, &copied)
			list = append(list, copied)
		}
		// If err != nil then hostZone wasn't found, return empty list
	}
	return
}
