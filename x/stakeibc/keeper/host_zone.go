package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// GetHostZoneCount get the total number of hostZone
func (k Keeper) GetHostZoneCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.HostZoneCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetHostZoneCount set the total number of hostZone
func (k Keeper) SetHostZoneCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.HostZoneCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// SetHostZone set a specific hostZone in the store
func (k Keeper) SetHostZone(ctx sdk.Context, hostZone types.HostZone) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	b := k.cdc.MustMarshal(&hostZone)
	store.Set([]byte(hostZone.ChainId), b)
}

// GetHostZone returns a hostZone from its id
func (k Keeper) GetHostZone(ctx sdk.Context, chain_id string) (val types.HostZone, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	b := store.Get([]byte(chain_id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetHostZoneFromHostDenom returns a HostZone from a HostDenom
func (k Keeper) GetHostZoneFromHostDenom(ctx sdk.Context, denom string) (*types.HostZone, error) {
	var matchZone types.HostZone
	inDenom := strings.ToUpper(denom)
	k.IterateHostZones(ctx, func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		zoneDenom := strings.ToUpper(zoneInfo.HostDenom)
		if zoneDenom == inDenom {
			matchZone = zoneInfo
			return nil
		}
		return nil
	})
	if matchZone.ChainId != "" {
		return &matchZone, nil
	}
	return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "No HostZone for %s found", denom)
}

// RemoveHostZone removes a hostZone from the store
func (k Keeper) RemoveHostZone(ctx sdk.Context, chain_id string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	store.Delete([]byte(chain_id))
}

// GetAllHostZone returns all hostZone
func (k Keeper) GetAllHostZone(ctx sdk.Context) (list []types.HostZone) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.HostZone
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) AddDelegationToValidator(ctx sdk.Context, hostZone types.HostZone, valAddr string, amt int64) (success bool) {
	for _, val := range hostZone.GetValidators() {
		k.Logger(ctx).Info(fmt.Sprintf("Validator %s %d %d", val.GetAddress(), val.GetDelegationAmt(), amt))
		if val.GetAddress() == valAddr {
			if amt >= 0 {
				val.DelegationAmt = val.GetDelegationAmt() + cast.ToUint64(amt)
				return true
			} else {
				absAmt := cast.ToUint64(-amt)
				if absAmt > val.GetDelegationAmt() {
					k.Logger(ctx).Error(fmt.Sprintf("Delegation amount %d is greater than validator %s delegation amount %d", absAmt, valAddr, val.GetDelegationAmt()))
					return false
				}
				val.DelegationAmt = val.GetDelegationAmt() - absAmt
				return true
			}
		}
	}
	k.Logger(ctx).Info(fmt.Sprintf("Could not find validator %s on host zone %s", valAddr, hostZone.GetChainId()))
	return false
}

func (k Keeper) RemoveValidatorFromHostZone(ctx sdk.Context, chainId string, validatorAddress string) (success bool) {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("HostZone not found %s", chainId))
		return false
	}
	for i, val := range hostZone.Validators {
		if val.GetAddress() == validatorAddress {
			if val.GetDelegationAmt() == 0 && val.GetWeight() == 0 {
				hostZone.Validators = append(hostZone.Validators[:i], hostZone.Validators[i+1:]...)
				return true
			} else {
				k.Logger(ctx).Error(fmt.Sprintf("Validator %s has non-zero delegation (%d) or weight (%d)", validatorAddress, val.GetDelegationAmt(), val.GetWeight()))

				return false
			}
		}
	}
	k.SetHostZone(ctx, hostZone)
	k.Logger(ctx).Error(fmt.Sprintf("Validator %s not found on the host zone %s", validatorAddress, chainId))
	return false
}

// GetHostZoneIDFromBytes returns ID in uint64 format from a byte array
func GetHostZoneIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}

// GetHostZoneFromIBCDenom returns a HostZone from a IBCDenom
func (k Keeper) GetHostZoneFromIBCDenom(ctx sdk.Context, denom string) (*types.HostZone, error) {
	var matchZone types.HostZone
	k.IterateHostZones(ctx, func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		if zoneInfo.IBCDenom == denom {
			matchZone = zoneInfo
			return nil
		}
		return nil
	})
	if matchZone.ChainId != "" {
		return &matchZone, nil
	}
	return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "No HostZone for %s found", denom)
}

// IterateHostZones iterates zones
func (k Keeper) IterateHostZones(ctx sdk.Context, fn func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))

	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		zone := types.HostZone{}
		k.cdc.MustUnmarshal(iterator.Value(), &zone)

		error := fn(ctx, i, zone)

		if error != nil {
			break
		}
		i++
	}
}

// GetRedemptionAccount gest the redemption account from a HostZone
func (k Keeper) GetRedemptionAccount(ctx sdk.Context, hostZone types.HostZone) (*types.ICAAccount, bool) {
	if hostZone.RedemptionAccount == nil {
		return nil, false
	}
	return hostZone.RedemptionAccount, true
}
