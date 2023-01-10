package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/utils"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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
	k.IterateHostZones(ctx, func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		if zoneInfo.HostDenom == denom {
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

func (k Keeper) AddDelegationToValidator(ctx sdk.Context, hostZone types.HostZone, validatorAddress string, amount sdk.Int, callbackId string) (success bool) {
	for _, validator := range hostZone.Validators {
		if validator.Address == validatorAddress {
			k.Logger(ctx).Info(utils.LogCallbackWithHostZone(hostZone.ChainId, callbackId,
				"  Validator %s, Current Delegation: %v, Delegation Change: %v", validator.Address, validator.DelegationAmt, amount))

			if amount.GTE(sdk.ZeroInt()) {
				validator.DelegationAmt = validator.DelegationAmt.Add(amount)
				return true
			}
			absAmt := amount.Abs()
			if absAmt.GT(validator.DelegationAmt) {
				k.Logger(ctx).Error(fmt.Sprintf("Delegation amount %v is greater than validator %s delegation amount %v", absAmt, validatorAddress, validator.DelegationAmt))
				return false
			}
			validator.DelegationAmt = validator.DelegationAmt.Sub(absAmt)
			return true
		}
	}

	k.Logger(ctx).Error(fmt.Sprintf("Could not find validator %s on host zone %s", validatorAddress, hostZone.ChainId))
	return false
}

// Appends a validator to host zone (if the host zone is not already at capacity)
// If the validator is added through governance, the weight is equal to the minimum weight across the set
// If the validator is added through an admin transactions, the weight is specified in the message
func (k Keeper) AddValidatorToHostZone(ctx sdk.Context, msg *types.MsgAddValidator, fromGovernance bool) error {
	// Get the corresponding host zone
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		errMsg := fmt.Sprintf("Host Zone (%s) not found", msg.HostZone)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrap(types.ErrHostZoneNotFound, errMsg)
	}

	// Get max number of validators and confirm we won't exceed it
	err := k.ConfirmValSetHasSpace(ctx, hostZone.Validators)
	if err != nil {
		return sdkerrors.Wrap(types.ErrMaxNumValidators, "cannot add validator on host zone")
	}

	// Check that we don't already have this validator
	// Grab the minimum weight in the process (to assign to validator's added through governance)
	var minWeight uint64 = math.MaxUint64
	for _, validator := range hostZone.Validators {
		if validator.Address == msg.Address {
			errMsg := fmt.Sprintf("Validator address (%s) already exists on Host Zone (%s)", msg.Address, msg.HostZone)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrap(types.ErrValidatorAlreadyExists, errMsg)
		}
		if validator.Name == msg.Name {
			errMsg := fmt.Sprintf("Validator name (%s) already exists on Host Zone (%s)", msg.Name, msg.HostZone)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrap(types.ErrValidatorAlreadyExists, errMsg)
		}
		// Store the min weight to assign to new validator added through governance (ignore zero-weight validators)
		if validator.Weight < minWeight && validator.Weight > 0 {
			minWeight = validator.Weight
		}
	}

	// If the validator was added via governance, set the weight to the min validator weight of the host zone
	valWeight := msg.Weight
	if fromGovernance {
		valWeight = minWeight
	}

	// Finally, add the validator to the host
	hostZone.Validators = append(hostZone.Validators, &types.Validator{
		Name:           msg.Name,
		Address:        msg.Address,
		Status:         types.Validator_ACTIVE,
		CommissionRate: msg.Commission,
		DelegationAmt:  sdk.ZeroInt(),
		Weight:         valWeight,
	})

	k.SetHostZone(ctx, hostZone)

	return nil
}

// Removes a validator from a host zone
// The validator must be zero-weight and have no delegations in order to be removed
func (k Keeper) RemoveValidatorFromHostZone(ctx sdk.Context, chainId string, validatorAddress string) error {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		errMsg := fmt.Sprintf("HostZone (%s) not found", chainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}
	for i, val := range hostZone.Validators {
		if val.GetAddress() == validatorAddress {
			if val.DelegationAmt.IsZero() && val.Weight == 0 {
				hostZone.Validators = append(hostZone.Validators[:i], hostZone.Validators[i+1:]...)
				k.SetHostZone(ctx, hostZone)
				return nil
			}
			errMsg := fmt.Sprintf("Validator (%s) has non-zero delegation (%v) or weight (%d)", validatorAddress, val.DelegationAmt, val.Weight)
			k.Logger(ctx).Error(errMsg)
			return errors.New(errMsg)
		}
	}
	errMsg := fmt.Sprintf("Validator address (%s) not found on host zone (%s)", validatorAddress, chainId)
	k.Logger(ctx).Error(errMsg)
	return sdkerrors.Wrapf(types.ErrValidatorNotFound, errMsg)
}

// GetHostZoneFromIBCDenom returns a HostZone from a IBCDenom
func (k Keeper) GetHostZoneFromIBCDenom(ctx sdk.Context, denom string) (*types.HostZone, error) {
	var matchZone types.HostZone
	k.IterateHostZones(ctx, func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		if zoneInfo.IbcDenom == denom {
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
		k.Logger(ctx).Debug(fmt.Sprintf("Iterating HostZone %d", i))
		zone := types.HostZone{}
		k.cdc.MustUnmarshal(iterator.Value(), &zone)

		error := fn(ctx, i, zone)

		if error != nil {
			break
		}
		i++
	}
}

func (k Keeper) GetRedemptionAccount(ctx sdk.Context, hostZone types.HostZone) (*types.ICAAccount, bool) {
	if hostZone.RedemptionAccount == nil {
		return nil, false
	}
	return hostZone.RedemptionAccount, true
}
