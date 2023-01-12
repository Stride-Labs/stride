package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	oldstakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/migrations/v2/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func convertToNewValidator(oldValidator oldstakeibctypes.Validator) stakeibctypes.Validator {
	return stakeibctypes.Validator{
		Name:                 oldValidator.Name,
		Address:              oldValidator.Address,
		Status:               stakeibctypes.Validator_ValidatorStatus(oldValidator.Status),
		CommissionRate:       oldValidator.CommissionRate,
		DelegationAmt:        sdk.NewIntFromUint64(oldValidator.DelegationAmt),
		Weight:               oldValidator.Weight,
		InternalExchangeRate: (*stakeibctypes.ValidatorExchangeRate)(oldValidator.InternalExchangeRate),
	}
}

func convertICAAccount(oldAccount oldstakeibctypes.ICAAccount) stakeibctypes.ICAAccount {
	return stakeibctypes.ICAAccount{Address: oldAccount.Address, Target: stakeibctypes.ICAAccountType(oldAccount.Target)}
}

func convertToNewHostZone(oldHostZone oldstakeibctypes.HostZone) stakeibctypes.HostZone {
	var validators []*stakeibctypes.Validator
	var blacklistValidator []*stakeibctypes.Validator

	for _, oldValidator := range oldHostZone.Validators {
		newValidator := convertToNewValidator(*oldValidator)
		validators = append(validators, &newValidator)
	}

	for _, oldValidator := range oldHostZone.BlacklistedValidators {
		newValidator := convertToNewValidator(*oldValidator)
		blacklistValidator = append(blacklistValidator, &newValidator)
	}

	newWithdrawalAccount := convertICAAccount(*oldHostZone.WithdrawalAccount)
	newFeeAccount := convertICAAccount(*oldHostZone.FeeAccount)
	newDelegationAccount := convertICAAccount(*oldHostZone.DelegationAccount)
	newRedemptionAccount := convertICAAccount(*oldHostZone.RedemptionAccount)

	return stakeibctypes.HostZone{
		ChainId:               oldHostZone.ChainId,
		ConnectionId:          oldHostZone.ConnectionId,
		Bech32Prefix:          oldHostZone.Bech32Prefix,
		TransferChannelId:     oldHostZone.TransferChannelId,
		Validators:            validators,
		BlacklistedValidators: blacklistValidator,
		WithdrawalAccount:     &newWithdrawalAccount,
		FeeAccount:            &newFeeAccount,
		DelegationAccount:     &newDelegationAccount,
		RedemptionAccount:     &newRedemptionAccount,
		IbcDenom:              oldHostZone.IbcDenom,
		HostDenom:             oldHostZone.HostDenom,
		LastRedemptionRate:    oldHostZone.LastRedemptionRate,
		RedemptionRate:        oldHostZone.RedemptionRate,
		UnbondingFrequency:    oldHostZone.UnbondingFrequency,
		StakedBal:             sdk.NewIntFromUint64(oldHostZone.StakedBal),
		Address:               oldHostZone.Address,
	}
}

func migrateHostZone(store sdk.KVStore, cdc codec.BinaryCodec) error {
	paramsStore := prefix.NewStore(store, []byte(stakeibctypes.HostZoneKey))

	iterator := paramsStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldHostZone oldstakeibctypes.HostZone
		err := cdc.Unmarshal(iterator.Value(), &oldHostZone)
		if err != nil {
			return err
		}

		// Convert and serialize using the new type
		newHostZone := convertToNewHostZone(oldHostZone)
		newHostZoneBz, err := cdc.Marshal(&newHostZone)
		if err != nil {
			return err
		}

		// Store new type
		paramsStore.Set(iterator.Key(), newHostZoneBz)
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	return migrateHostZone(store, cdc)
}
