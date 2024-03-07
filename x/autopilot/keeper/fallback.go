package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v19/x/autopilot/types"
)

// Stores a fallback address for an outbound transfer
func (k Keeper) SetTransferFallbackAddress(ctx sdk.Context, channelId string, sequence uint64, address string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TransferFallbackAddressPrefix)
	key := types.GetTransferFallbackAddressKey(channelId, sequence)
	value := []byte(address)
	store.Set(key, value)
}

// Removes a fallback address from the store
// This is used after the ack or timeout for a packet has been received
func (k Keeper) RemoveTransferFallbackAddress(ctx sdk.Context, channelId string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TransferFallbackAddressPrefix)
	key := types.GetTransferFallbackAddressKey(channelId, sequence)
	store.Delete(key)
}

// Returns a fallback address, given the channel ID and sequence number of the packet
// If no fallback address has been stored, return false
func (k Keeper) GetTransferFallbackAddress(ctx sdk.Context, channelId string, sequence uint64) (address string, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TransferFallbackAddressPrefix)

	key := types.GetTransferFallbackAddressKey(channelId, sequence)
	valueBz := store.Get(key)

	if len(valueBz) == 0 {
		return "", false
	}

	return string(valueBz), true
}
