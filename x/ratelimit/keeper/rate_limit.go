package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v14/x/ratelimit/types"
)

// Get the rate limit byte key built from the denom and channelId
func GetRateLimitItemKey(denom string, channelId string) []byte {
	return append(types.KeyPrefix(denom), types.KeyPrefix(channelId)...)
}

// The total value on a given path (aka, the denominator in the percentage calculation)
// is the total supply of the given denom
func (k Keeper) GetChannelValue(ctx sdk.Context, denom string) sdkmath.Int {
	return k.bankKeeper.GetSupply(ctx, denom).Amount
}

// If the rate limit is exceeded or the denom is blacklisted, we emit an event
func EmitTransferDeniedEvent(ctx sdk.Context, reason, denom, channelId string, direction types.PacketDirection, amount sdkmath.Int, err error) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTransferDenied,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyAction, strings.ToLower(direction.String())), // packet_send or packet_recv
			sdk.NewAttribute(types.AttributeKeyDenom, denom),
			sdk.NewAttribute(types.AttributeKeyChannel, channelId),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		),
	)
}

// Adds an amount to the flow in either the SEND or RECV direction
func (k Keeper) UpdateFlow(rateLimit types.RateLimit, direction types.PacketDirection, amount sdkmath.Int) error {
	switch direction {
	case types.PACKET_SEND:
		return rateLimit.Flow.AddOutflow(amount, *rateLimit.Quota)
	case types.PACKET_RECV:
		return rateLimit.Flow.AddInflow(amount, *rateLimit.Quota)
	default:
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid packet direction (%s)", direction.String())
	}
}

// Checks whether the given packet will exceed the rate limit
// Called by OnRecvPacket and OnSendPacket
func (k Keeper) CheckRateLimitAndUpdateFlow(
	ctx sdk.Context,
	direction types.PacketDirection,
	packetInfo RateLimitedPacketInfo,
) (updatedFlow bool, err error) {
	denom := packetInfo.Denom
	channelId := packetInfo.ChannelID
	amount := packetInfo.Amount

	// First check if the denom is blacklisted
	if k.IsDenomBlacklisted(ctx, denom) {
		err := errorsmod.Wrapf(types.ErrDenomIsBlacklisted, "denom %s is blacklisted", denom)
		EmitTransferDeniedEvent(ctx, types.EventBlacklistedDenom, denom, channelId, direction, amount, err)
		return false, err
	}

	// If there's no rate limit yet for this denom, no action is necessary
	rateLimit, found := k.GetRateLimit(ctx, denom, channelId)
	if !found {
		return false, nil
	}

	// Check if the sender/receiver pair is whitelisted
	// If so, return a success without modifying the quota
	if k.IsAddressPairWhitelisted(ctx, packetInfo.Sender, packetInfo.Receiver) {
		return false, nil
	}

	// Update the flow object with the change in amount
	if err := k.UpdateFlow(rateLimit, direction, amount); err != nil {
		// If the rate limit was exceeded, emit an event
		EmitTransferDeniedEvent(ctx, types.EventRateLimitExceeded, denom, channelId, direction, amount, err)
		return false, err
	}

	// If there's no quota error, update the rate limit object in the store with the new flow
	k.SetRateLimit(ctx, rateLimit)

	return true, nil
}

// If a SendPacket fails or times out, undo the outflow increment that happened during the send
func (k Keeper) UndoSendPacket(ctx sdk.Context, channelId string, sequence uint64, denom string, amount sdkmath.Int) error {
	rateLimit, found := k.GetRateLimit(ctx, denom, channelId)
	if !found {
		return nil
	}

	// If the packet was sent during this quota, decrement the outflow
	// Otherwise, it can be ignored
	if k.CheckPacketSentDuringCurrentQuota(ctx, channelId, sequence) {
		rateLimit.Flow.Outflow = rateLimit.Flow.Outflow.Sub(amount)
		k.SetRateLimit(ctx, rateLimit)

		k.RemovePendingSendPacket(ctx, channelId, sequence)
	}

	return nil
}

// Reset the rate limit after expiration
// The inflow and outflow should get reset to 0, the channelValue should be updated,
// and all pending send packet sequence numbers should be removed
func (k Keeper) ResetRateLimit(ctx sdk.Context, denom string, channelId string) error {
	rateLimit, found := k.GetRateLimit(ctx, denom, channelId)
	if !found {
		return types.ErrRateLimitNotFound
	}

	flow := types.Flow{
		Inflow:       sdkmath.ZeroInt(),
		Outflow:      sdkmath.ZeroInt(),
		ChannelValue: k.GetChannelValue(ctx, denom),
	}
	rateLimit.Flow = &flow

	k.SetRateLimit(ctx, rateLimit)
	k.RemoveAllChannelPendingSendPackets(ctx, channelId)
	return nil
}

// Stores/Updates a rate limit object in the store
func (k Keeper) SetRateLimit(ctx sdk.Context, rateLimit types.RateLimit) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)

	rateLimitKey := GetRateLimitItemKey(rateLimit.Path.Denom, rateLimit.Path.ChannelId)
	rateLimitValue := k.cdc.MustMarshal(&rateLimit)

	store.Set(rateLimitKey, rateLimitValue)
}

// Removes a rate limit object from the store using denom and channel-id
func (k Keeper) RemoveRateLimit(ctx sdk.Context, denom string, channelId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)
	rateLimitKey := GetRateLimitItemKey(denom, channelId)
	store.Delete(rateLimitKey)
}

// Grabs and returns a rate limit object from the store using denom and channel-id
func (k Keeper) GetRateLimit(ctx sdk.Context, denom string, channelId string) (rateLimit types.RateLimit, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)

	rateLimitKey := GetRateLimitItemKey(denom, channelId)
	rateLimitValue := store.Get(rateLimitKey)

	if len(rateLimitValue) == 0 {
		return rateLimit, false
	}

	k.cdc.MustUnmarshal(rateLimitValue, &rateLimit)
	return rateLimit, true
}

// Returns all rate limits stored
func (k Keeper) GetAllRateLimits(ctx sdk.Context) []types.RateLimit {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allRateLimits := []types.RateLimit{}
	for ; iterator.Valid(); iterator.Next() {

		rateLimit := types.RateLimit{}
		k.cdc.MustUnmarshal(iterator.Value(), &rateLimit)
		allRateLimits = append(allRateLimits, rateLimit)
	}

	return allRateLimits
}

// Sets the sequence number of a packet that was just sent
func (k Keeper) SetPendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)
	key := types.GetPendingSendPacketKey(channelId, sequence)
	store.Set(key, []byte{1})
}

// Remove a pending packet sequence number from the store
// Used after the ack or timeout for a packet has been received
func (k Keeper) RemovePendingSendPacket(ctx sdk.Context, channelId string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)
	key := types.GetPendingSendPacketKey(channelId, sequence)
	store.Delete(key)
}

// Checks whether the packet sequence number is in the store - indicating that it was
// sent during the current quota
func (k Keeper) CheckPacketSentDuringCurrentQuota(ctx sdk.Context, channelId string, sequence uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)
	key := types.GetPendingSendPacketKey(channelId, sequence)
	valueBz := store.Get(key)
	found := len(valueBz) != 0
	return found
}

// Get all pending packet sequence numbers
func (k Keeper) GetAllPendingSendPackets(ctx sdk.Context) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	pendingPackets := []string{}
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		channelId := string(key[:types.PendingSendPacketChannelLength])
		channelId = strings.TrimRight(channelId, "\x00") // removes null bytes from suffix
		sequence := binary.BigEndian.Uint64(key[types.PendingSendPacketChannelLength:])

		packetId := fmt.Sprintf("%s/%d", channelId, sequence)
		pendingPackets = append(pendingPackets, packetId)
	}

	return pendingPackets
}

// Remove all pending sequence numbers from the store
// This is executed when the quota resets
func (k Keeper) RemoveAllChannelPendingSendPackets(ctx sdk.Context, channelId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingSendPacketPrefix)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefix(channelId))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

// Adds a denom to a blacklist to prevent all IBC transfers with this denom
func (k Keeper) AddDenomToBlacklist(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)
	key := types.KeyPrefix(denom)
	store.Set(key, []byte{1})
}

// Removes a denom from a blacklist to re-enable IBC transfers for that denom
func (k Keeper) RemoveDenomFromBlacklist(ctx sdk.Context, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)
	key := types.KeyPrefix(denom)
	store.Delete(key)
}

// Check if a denom is currently blacklisted
func (k Keeper) IsDenomBlacklisted(ctx sdk.Context, denom string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)

	key := types.KeyPrefix(denom)
	value := store.Get(key)
	found := len(value) != 0

	return found
}

// Get all the blacklisted denoms
func (k Keeper) GetAllBlacklistedDenoms(ctx sdk.Context) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomBlacklistKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allBlacklistedDenoms := []string{}
	for ; iterator.Valid(); iterator.Next() {
		allBlacklistedDenoms = append(allBlacklistedDenoms, string(iterator.Key()))
	}

	return allBlacklistedDenoms
}

// Adds an pair of sender and receiver addresses to the whitelist to allow all
// IBC transfers between those addresses to skip all flow calculations
func (k Keeper) SetWhitelistedAddressPair(ctx sdk.Context, whitelist types.WhitelistedAddressPair) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)
	key := types.GetAddressWhitelistKey(whitelist.Sender, whitelist.Receiver)
	value := k.cdc.MustMarshal(&whitelist)
	store.Set(key, value)
}

// Removes a whitelisted address pair so that it's transfers are counted in the quota
func (k Keeper) RemoveWhitelistedAddressPair(ctx sdk.Context, sender, receiver string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)
	key := types.GetAddressWhitelistKey(sender, receiver)
	store.Delete(key)
}

// Check if a sender/receiver address pair is currently whitelisted
func (k Keeper) IsAddressPairWhitelisted(ctx sdk.Context, sender, receiver string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)

	key := types.GetAddressWhitelistKey(sender, receiver)
	value := store.Get(key)
	found := len(value) != 0

	return found
}

// Get all the whitelisted addresses
func (k Keeper) GetAllWhitelistedAddressPairs(ctx sdk.Context) []types.WhitelistedAddressPair {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AddressWhitelistKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allWhitelistedAddresses := []types.WhitelistedAddressPair{}
	for ; iterator.Valid(); iterator.Next() {
		whitelist := types.WhitelistedAddressPair{}
		k.cdc.MustUnmarshal(iterator.Value(), &whitelist)
		allWhitelistedAddresses = append(allWhitelistedAddresses, whitelist)
	}

	return allWhitelistedAddresses
}
