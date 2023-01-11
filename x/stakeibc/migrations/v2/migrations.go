package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
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

	return stakeibctypes.HostZone{
		ChainId:               oldHostZone.ChainId,
		ConnectionId:          oldHostZone.ConnectionId,
		Bech32Prefix:          oldHostZone.Bech32Prefix,
		TransferChannelId:     oldHostZone.TransferChannelId,
		Validators:            validators,
		BlacklistedValidators: blacklistValidator,
		WithdrawalAccount:     oldHostZone.WithdrawalAccount,
		FeeAccount:            oldHostZone.FeeAccount,
		DelegationAccount:     oldHostZone.DelegationAccount,
		RedemptionAccount:     oldHostZone.RedemptionAccount,
		IbcDenom:              oldHostZone.IbcDenom,
		HostDenom:             oldHostZone.HostDenom,
		LastRedemptionRate:    oldHostZone.LastRedemptionRate,
		RedemptionRate:        oldHostZone.RedemptionRate,
		UnbondingFrequency:    oldHostZone.UnbondingFrequency,
		StakedBal:             sdk.NewIntFromUint64(oldHostZone.StakedBal),
		Address:               oldHostZone.Address,
	}
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	return nil
}
