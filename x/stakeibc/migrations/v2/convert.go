package v2

import (
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
