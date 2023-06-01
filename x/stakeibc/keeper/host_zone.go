package keeper

import (
	"errors"
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

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
	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "No HostZone for %s found", denom)
}

// GetHostZoneFromTransferChannelID returns a HostZone from a transfer channel ID
func (k Keeper) GetHostZoneFromTransferChannelID(ctx sdk.Context, channelID string) (hostZone types.HostZone, found bool) {
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		if hostZone.TransferChannelId == channelID {
			return hostZone, true
		}
	}
	return types.HostZone{}, false
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

// GetAllActiveHostZone returns all hostZones that are active (halted = false)
func (k Keeper) GetAllActiveHostZone(ctx sdk.Context) (list []types.HostZone) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.HostZone
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if !val.Halted {
			list = append(list, val)
		}
	}

	return
}

func (k Keeper) AddDelegationToValidator(ctx sdk.Context, hostZone types.HostZone, validatorAddress string, amount sdkmath.Int, callbackId string) (success bool) {
	for _, validator := range hostZone.Validators {
		if validator.Address == validatorAddress {
			k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(hostZone.ChainId, callbackId,
				"  Validator %s, Current Delegation: %v, Delegation Change: %v", validator.Address, validator.DelegationAmt, amount))

			if amount.GTE(sdkmath.ZeroInt()) {
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
func (k Keeper) AddValidatorToHostZone(ctx sdk.Context, chainId string, validator types.Validator, fromGovernance bool) error {
	// Get the corresponding host zone
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "Host Zone (%s) not found", chainId)
	}

	// Get max number of validators and confirm we won't exceed it
	err := k.ConfirmValSetHasSpace(ctx, hostZone.Validators)
	if err != nil {
		return errorsmod.Wrap(types.ErrMaxNumValidators, "cannot add validator on host zone")
	}

	// Check that we don't already have this validator
	// Grab the minimum weight in the process (to assign to validator's added through governance)
	var minWeight uint64 = math.MaxUint64
	for _, existingValidator := range hostZone.Validators {
		if existingValidator.Address == validator.Address {
			return errorsmod.Wrapf(types.ErrValidatorAlreadyExists, "Validator address (%s) already exists on Host Zone (%s)", validator.Address, chainId)
		}
		if existingValidator.Name == validator.Name {
			return errorsmod.Wrapf(types.ErrValidatorAlreadyExists, "Validator name (%s) already exists on Host Zone (%s)", validator.Name, chainId)
		}
		// Store the min weight to assign to new validator added through governance (ignore zero-weight validators)
		if existingValidator.Weight < minWeight && existingValidator.Weight > 0 {
			minWeight = existingValidator.Weight
		}
	}

	// If the validator was added via governance, set the weight to the min validator weight of the host zone
	valWeight := validator.Weight
	if fromGovernance {
		valWeight = minWeight
	}

	// Finally, add the validator to the host
	hostZone.Validators = append(hostZone.Validators, &types.Validator{
		Name:          validator.Name,
		Address:       validator.Address,
		Weight:        valWeight,
		DelegationAmt: sdkmath.ZeroInt(),
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
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, errMsg)
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
	return errorsmod.Wrapf(types.ErrValidatorNotFound, errMsg)
}

// Get a validator and its index from a list of validators, by address
func GetValidatorFromAddress(validators []*types.Validator, address string) (val types.Validator, index int64, found bool) {
	for i, v := range validators {
		if v.Address == address {
			return *v, int64(i), true
		}
	}
	return types.Validator{}, 0, false
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
	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "No HostZone for %s found", denom)
}

// Validate whether a denom is a supported liquid staking token
func (k Keeper) CheckIsStToken(ctx sdk.Context, denom string) bool {
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom) == denom {
			return true
		}
	}
	return false
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
