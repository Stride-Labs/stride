package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

// Adds a user link record to the store
func (k Keeper) AddUserLink(ctx sdk.Context, airdropId, strideAddress, hostAddress string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserLinksKeyPrefix)
	// Fetch current user links
	key := types.UserLinksKey(airdropId, strideAddress)
	userLinksBz := store.Get(key)

	var userLinks types.UserLinks

	if len(userLinksBz) != 0 {
		// If there are user links, unmarshal them and append the new host address

		k.cdc.MustUnmarshal(userLinksBz, &userLinks)

		// Check that the new host address is not already linked to this airdrop
		for _, existingHostAddress := range userLinks.HostAddresses {
			if existingHostAddress == hostAddress {
				return
			}
		}
		// Append the new host address
		userLinks.HostAddresses = append(userLinks.HostAddresses, hostAddress)
	} else {
		// If there are no user links yet, create a new one with the given host address

		userLinks = types.UserLinks{
			AirdropId:     airdropId,
			StrideAddress: strideAddress,
			HostAddresses: []string{hostAddress},
		}
	}

	// Reset the claim type for all allocations involed with this operation
	// Make sure this happens only when we also update the link
	for _, address := range append(userLinks.HostAddresses, userLinks.StrideAddress) {
		k.ResetClaimType(ctx, airdropId, address)
	}

	// Marshal and set the new link record in the store
	userLinksBz = k.cdc.MustMarshal(&userLinks)
	store.Set(key, userLinksBz)
}

// Removes a user links record from the store
func (k Keeper) RemoveUserLinks(ctx sdk.Context, airdropId, strideAddress string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserLinksKeyPrefix)
	key := types.UserLinksKey(airdropId, strideAddress)

	userLinksBz := store.Get(key)
	if len(userLinksBz) == 0 {
		return
	}

	var userLinks types.UserLinks
	k.cdc.MustUnmarshal(userLinksBz, &userLinks)

	// Reset the claim type for all allocations involed with this operation
	// Make sure this happens only when we also update the link
	for _, address := range append(userLinks.HostAddresses, userLinks.StrideAddress) {
		k.ResetClaimType(ctx, airdropId, address)
	}

	store.Delete(key)
}

// Removes a single user link from the store, while maintaining the rest of the links for that user
func (k Keeper) RemoveUserLink(ctx sdk.Context, airdropId, strideAddress, hostAddress string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserLinksKeyPrefix)
	// Fetch current user links
	key := types.UserLinksKey(airdropId, strideAddress)
	userLinksBz := store.Get(key)

	// If there are no user links yet, create a new one with the given host address
	userLinks := types.UserLinks{}

	// If there are user links, unmarshal them and append the new host address
	// Otherwise, return (nothing to remove)
	if len(userLinksBz) != 0 {
		k.cdc.MustUnmarshal(userLinksBz, &userLinks)
	} else {
		return
	}

	// Check that the new host address is not already linked to this airdrop
	newHostAddresses := []string{}
	for _, existingHostAddress := range userLinks.HostAddresses {
		if existingHostAddress != hostAddress {
			newHostAddresses = append(newHostAddresses, existingHostAddress)
		}
	}

	// Return if hostAddress wasn't actually linked in the first place
	if len(newHostAddresses) == len(userLinks.HostAddresses) {
		return
	}

	// Reset the claim type for all allocations involed with this operation
	// Make sure this happens only when we also update the link
	// Also make sure we use the list before the deletion
	for _, address := range append(userLinks.HostAddresses, userLinks.StrideAddress) {
		k.ResetClaimType(ctx, airdropId, address)
	}

	// Set the new host address
	userLinks.HostAddresses = newHostAddresses

	// Marshal and set the new link record in the store
	userLinksBz = k.cdc.MustMarshal(&userLinks)
	store.Set(key, userLinksBz)
}

// Retrieves a user link record from the store
func (k Keeper) GetUserLinks(ctx sdk.Context, airdropId, address string) (link types.UserLinks, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserLinksKeyPrefix)

	key := types.UserLinksKey(airdropId, address)
	linksBz := store.Get(key)

	if len(linksBz) == 0 {
		return link, false
	}

	k.cdc.MustUnmarshal(linksBz, &link)
	return link, true
}

// Retrieves a all user links for an airdrop from the store
func (k Keeper) GetAllLinks(ctx sdk.Context, airdropId string) (links []types.UserLinks) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserLinksKeyPrefix)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefix(airdropId))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		link := types.UserLinks{}
		k.cdc.MustUnmarshal(iterator.Value(), &link)
		links = append(links, link)
	}

	return links
}
