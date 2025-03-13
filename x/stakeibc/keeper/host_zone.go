package keeper

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v26/utils"
	recordstypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

const (
	MinValidatorsBeforeWeightCapCheck = 10
)

// SetHostZone set a specific hostZone in the store
func (k Keeper) SetHostZone(ctx sdk.Context, hostZone types.HostZone) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	b := k.cdc.MustMarshal(&hostZone)
	store.Set([]byte(hostZone.ChainId), b)
}

// GetHostZone returns a hostZone from its id
func (k Keeper) GetHostZone(ctx sdk.Context, chainId string) (val types.HostZone, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	b := store.Get([]byte(chainId))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetActiveHostZone returns an error if the host zone is not found or if it's found, but is halted
func (k Keeper) GetActiveHostZone(ctx sdk.Context, chainId string) (hostZone types.HostZone, err error) {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return hostZone, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", chainId)
	}
	if hostZone.Halted {
		return hostZone, types.ErrHaltedHostZone.Wrapf("host zone %s is halted", chainId)
	}
	return hostZone, nil
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
	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "No HostZone for %s denom found", denom)
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
func (k Keeper) RemoveHostZone(ctx sdk.Context, chainId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.HostZoneKey))
	store.Delete([]byte(chainId))
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

// Unregisters a host zone, including removing the module accounts and records
func (k Keeper) UnregisterHostZone(ctx sdk.Context, chainId string) error {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return types.ErrHostZoneNotFound.Wrapf("host zone %s not found", chainId)
	}

	// Burn all outstanding stTokens
	stTokenDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	for _, account := range k.AccountKeeper.GetAllAccounts(ctx) {
		stTokenBalance := k.bankKeeper.GetBalance(ctx, account.GetAddress(), stTokenDenom)
		stTokensToBurn := sdk.NewCoins(stTokenBalance)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account.GetAddress(), types.ModuleName, stTokensToBurn); err != nil {
			return err
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, stTokensToBurn); err != nil {
			return err
		}
	}

	// Set the escrow'd tokens to 0 (all the escrowed tokens should have been burned from the above)
	k.RecordsKeeper.TransferKeeper.SetTotalEscrowForDenom(ctx, sdk.NewCoin(stTokenDenom, sdkmath.ZeroInt()))

	// Remove module accounts
	depositAddress := types.NewHostZoneDepositAddress(chainId)
	communityPoolStakeAddress := types.NewHostZoneModuleAddress(chainId, CommunityPoolStakeHoldingAddressKey)
	communityPoolRedeemAddress := types.NewHostZoneModuleAddress(chainId, CommunityPoolRedeemHoldingAddressKey)

	k.AccountKeeper.RemoveAccount(ctx, k.AccountKeeper.GetAccount(ctx, depositAddress))
	k.AccountKeeper.RemoveAccount(ctx, k.AccountKeeper.GetAccount(ctx, communityPoolStakeAddress))
	k.AccountKeeper.RemoveAccount(ctx, k.AccountKeeper.GetAccount(ctx, communityPoolRedeemAddress))

	// Remove all deposit records for the host zone
	for _, depositRecord := range k.RecordsKeeper.GetAllDepositRecord(ctx) {
		if depositRecord.HostZoneId == chainId {
			k.RecordsKeeper.RemoveDepositRecord(ctx, depositRecord.Id)
		}
	}

	// Remove all epoch unbonding records for the host zone
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		updatedHostZoneUnbondings := []*recordstypes.HostZoneUnbonding{}
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			if hostZoneUnbonding.HostZoneId != chainId {
				updatedHostZoneUnbondings = append(updatedHostZoneUnbondings, hostZoneUnbonding)
			}
		}
		epochUnbondingRecord.HostZoneUnbondings = updatedHostZoneUnbondings
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	}

	// Remove all user redemption records for the host zone
	for _, userRedemptionRecord := range k.RecordsKeeper.GetAllUserRedemptionRecord(ctx) {
		if userRedemptionRecord.HostZoneId == chainId {
			k.RecordsKeeper.RemoveUserRedemptionRecord(ctx, userRedemptionRecord.Id)
		}
	}

	// Remove whitelisted address pairs from rate limit module
	rewardCollectorAddress := k.AccountKeeper.GetModuleAccount(ctx, types.RewardCollectorName).GetAddress()
	k.RatelimitKeeper.RemoveWhitelistedAddressPair(ctx, hostZone.DepositAddress, hostZone.DelegationIcaAddress)
	k.RatelimitKeeper.RemoveWhitelistedAddressPair(ctx, hostZone.FeeIcaAddress, rewardCollectorAddress.String())
	k.RatelimitKeeper.RemoveWhitelistedAddressPair(ctx, hostZone.CommunityPoolDepositIcaAddress, hostZone.CommunityPoolStakeHoldingAddress)
	k.RatelimitKeeper.RemoveWhitelistedAddressPair(ctx, hostZone.CommunityPoolDepositIcaAddress, hostZone.CommunityPoolRedeemHoldingAddress)
	k.RatelimitKeeper.RemoveWhitelistedAddressPair(ctx, hostZone.CommunityPoolStakeHoldingAddress, hostZone.CommunityPoolReturnIcaAddress)

	// Remove any blacklisted denoms from the rate limit module (may not be applicable)
	k.RatelimitKeeper.RemoveDenomFromBlacklist(ctx, utils.StAssetDenomFromHostZoneDenom(hostZone.HostDenom))

	// Finally, remove the host zone struct
	k.RemoveHostZone(ctx, chainId)

	return nil
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
// TODO [cleanup]: Remove this in favor of GetAllHostZones
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

// This will split a total delegation amount across validators, according to weights
// It returns a map of each portion, key'd on validator address
// Validator's with a slash query in progress are excluded
func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, totalDelegation sdkmath.Int) (map[string]sdkmath.Int, error) {
	// Confirm the expected delegation amount is greater than 0
	if !totalDelegation.IsPositive() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"Cannot calculate target delegation if final amount is less than or equal to zero (%v)", totalDelegation)
	}

	// Ignore any validators with a slash query in progress
	validators := []types.Validator{}
	for _, validator := range hostZone.Validators {
		if !validator.SlashQueryInProgress {
			validators = append(validators, *validator)
		}
	}

	// Sum the total weight across all validators
	totalWeight := k.GetTotalValidatorWeight(validators)
	if totalWeight == 0 {
		return nil, errorsmod.Wrapf(types.ErrNoValidatorWeights,
			"No non-zero validators found for host zone %s", hostZone.ChainId)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Total Validator Weight: %d", totalWeight))

	// sort validators by weight ascending
	sort.SliceStable(validators, func(i, j int) bool { // Do not use `Slice` here, it is stochastic
		if validators[i].Weight != validators[j].Weight {
			return validators[i].Weight < validators[j].Weight
		}
		// use name for tie breaker if weights are equal
		return validators[i].Address < validators[j].Address
	})

	// Assign each validator their portion of the delegation (and give any overflow to the last validator)
	targetUnbondingsByValidator := make(map[string]sdkmath.Int)
	totalAllocated := sdkmath.ZeroInt()
	for i, validator := range validators {
		// For the last element, we need to make sure that the totalAllocated is equal to the finalDelegation
		if i == len(validators)-1 {
			targetUnbondingsByValidator[validator.Address] = totalDelegation.Sub(totalAllocated)
		} else {
			delegateAmt := sdkmath.NewIntFromUint64(validator.Weight).Mul(totalDelegation).Quo(sdkmath.NewIntFromUint64(totalWeight))
			totalAllocated = totalAllocated.Add(delegateAmt)
			targetUnbondingsByValidator[validator.Address] = delegateAmt
		}
	}

	return targetUnbondingsByValidator, nil
}

// Enables redemptions by setting the parameter on the host zone to true
// This is used during the staketia/stakedym migrations
func (k Keeper) EnableRedemptions(ctx sdk.Context, chainId string) error {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return types.ErrHostZoneNotFound.Wrap(chainId)
	}

	hostZone.RedemptionsEnabled = true
	k.SetHostZone(ctx, hostZone)
	return nil
}
