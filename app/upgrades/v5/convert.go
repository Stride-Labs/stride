package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	claimtypes "github.com/Stride-Labs/stride/v5/x/claim/types"
	claimv1types "github.com/Stride-Labs/stride/v5/x/claim/types/v1"
	recordtypes "github.com/Stride-Labs/stride/v5/x/records/types"
	recordv1types "github.com/Stride-Labs/stride/v5/x/records/types/v1"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"
	stakeibcv1types "github.com/Stride-Labs/stride/v5/x/stakeibc/types/v1"
)

func convertToNewClaimParams(oldProp claimv1types.Params) claimtypes.Params {
	var newParams claimtypes.Params
	for _, airDrop := range oldProp.Airdrops {
		newAirDrop := claimtypes.Airdrop{
			AirdropIdentifier:  airDrop.AirdropIdentifier,
			AirdropStartTime:   airDrop.AirdropStartTime,
			AirdropDuration:    airDrop.AirdropDuration,
			ClaimDenom:         airDrop.ClaimDenom,
			DistributorAddress: airDrop.DistributorAddress,
			ClaimedSoFar:       sdk.NewInt(airDrop.ClaimedSoFar),
		}
		newParams.Airdrops = append(newParams.Airdrops, &newAirDrop)
	}
	return newParams
}

func convertToNewUserRedemptionRecord(oldProp recordv1types.UserRedemptionRecord) recordtypes.UserRedemptionRecord {
	return recordtypes.UserRedemptionRecord{
		Id:             oldProp.Id,
		Sender:         oldProp.Sender,
		Receiver:       oldProp.Receiver,
		Amount:         sdk.NewIntFromUint64(oldProp.Amount),
		Denom:          oldProp.Denom,
		HostZoneId:     oldProp.HostZoneId,
		EpochNumber:    sdk.NewIntFromUint64(oldProp.EpochNumber),
		ClaimIsPending: oldProp.ClaimIsPending,
	}
}

func convertToNewDepositRecord(oldProp recordv1types.DepositRecord) recordtypes.DepositRecord {
	return recordtypes.DepositRecord{
		Id:                 sdk.NewIntFromUint64(oldProp.Id),
		Amount:             sdk.NewInt(oldProp.Amount),
		Denom:              oldProp.Denom,
		HostZoneId:         oldProp.HostZoneId,
		Status:             recordtypes.DepositRecord_Status(oldProp.Status),
		DepositEpochNumber: sdk.NewIntFromUint64(oldProp.DepositEpochNumber),
		Source:             recordtypes.DepositRecord_Source(oldProp.Source),
	}
}

func convertToNewEpochUnbondingRecord(oldProp recordv1types.EpochUnbondingRecord) recordtypes.EpochUnbondingRecord {
	var epochUnbondingRecord recordtypes.EpochUnbondingRecord
	for _, hz := range oldProp.HostZoneUnbondings {
		newHz := recordtypes.HostZoneUnbonding{
			StTokenAmount:         sdk.NewIntFromUint64(hz.StTokenAmount),
			NativeTokenAmount:     sdk.NewIntFromUint64(hz.NativeTokenAmount),
			Denom:                 hz.Denom,
			HostZoneId:            hz.HostZoneId,
			UnbondingTime:         sdk.NewIntFromUint64(hz.UnbondingTime),
			Status:                recordtypes.HostZoneUnbonding_Status(hz.Status),
			UserRedemptionRecords: hz.UserRedemptionRecords,
		}
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &newHz)
	}
	return epochUnbondingRecord
}

// func convertToNewEpoch(oldProp epochv1types.EpochInfo) epochtypes.EpochInfo {
// 	return epochtypes.EpochInfo{
// 		Identifier:              oldProp.Identifier,
// 		StartTime:               oldProp.StartTime,
// 		Duration:                oldProp.Duration,
// 		CurrentEpoch:            sdk.NewInt(oldProp.CurrentEpoch),
// 		CurrentEpochStartTime:   oldProp.CurrentEpochStartTime,
// 		EpochCountingStarted:    oldProp.EpochCountingStarted,
// 		CurrentEpochStartHeight: sdk.NewInt(oldProp.CurrentEpochStartHeight),
// 	}
// }

func convertToNewDelegation(oldProp stakeibcv1types.Delegation) stakeibctypes.Delegation {
	internalExchangeRate := &stakeibctypes.ValidatorExchangeRate{
		InternalTokensToSharesRate: oldProp.Validator.InternalExchangeRate.InternalTokensToSharesRate,
		EpochNumber:                sdk.NewIntFromUint64(oldProp.Validator.InternalExchangeRate.EpochNumber),
	}
	return stakeibctypes.Delegation{
		DelegateAcctAddress: oldProp.DelegateAcctAddress,
		Validator: &stakeibctypes.Validator{
			Name:                 oldProp.Validator.Name,
			Address:              oldProp.Validator.Address,
			Status:               stakeibctypes.Validator_ValidatorStatus(oldProp.Validator.Status),
			CommissionRate:       sdk.NewIntFromUint64(oldProp.Validator.CommissionRate),
			DelegationAmt:        sdk.NewIntFromUint64(oldProp.Validator.DelegationAmt),
			Weight:               sdk.NewIntFromUint64(oldProp.Validator.Weight),
			InternalExchangeRate: internalExchangeRate,
		},
		Amt: sdk.NewInt(oldProp.Amt),
	}
}

func convertToNewHostZone(oldProp stakeibcv1types.HostZone) stakeibctypes.HostZone {
	var validators []*stakeibctypes.Validator
	var blacklistValidator []*stakeibctypes.Validator

	for _, val := range oldProp.Validators {
		internalExchangeRate := &stakeibctypes.ValidatorExchangeRate{
			InternalTokensToSharesRate: val.InternalExchangeRate.InternalTokensToSharesRate,
			EpochNumber:                sdk.NewIntFromUint64(val.InternalExchangeRate.EpochNumber),
		}
		newVal := stakeibctypes.Validator{
			Name:                 val.Name,
			Address:              val.Address,
			Status:               stakeibctypes.Validator_ValidatorStatus(val.Status),
			CommissionRate:       sdk.NewIntFromUint64(val.CommissionRate),
			DelegationAmt:        sdk.NewIntFromUint64(val.DelegationAmt),
			Weight:               sdk.NewIntFromUint64(val.Weight),
			InternalExchangeRate: internalExchangeRate,
		}
		validators = append(validators, &newVal)
	}

	for _, val := range oldProp.BlacklistedValidators {
		internalExchangeRate := &stakeibctypes.ValidatorExchangeRate{
			InternalTokensToSharesRate: val.InternalExchangeRate.InternalTokensToSharesRate,
			EpochNumber:                sdk.NewIntFromUint64(val.InternalExchangeRate.EpochNumber),
		}
		newVal := stakeibctypes.Validator{
			Name:                 val.Name,
			Address:              val.Address,
			Status:               stakeibctypes.Validator_ValidatorStatus(val.Status),
			CommissionRate:       sdk.NewIntFromUint64(val.CommissionRate),
			DelegationAmt:        sdk.NewIntFromUint64(val.DelegationAmt),
			Weight:               sdk.NewIntFromUint64(val.Weight),
			InternalExchangeRate: internalExchangeRate,
		}
		blacklistValidator = append(blacklistValidator, &newVal)
	}
	return stakeibctypes.HostZone{
		ChainId:               oldProp.ChainId,
		ConnectionId:          oldProp.ConnectionId,
		Bech32Prefix:          oldProp.Bech32Prefix,
		TransferChannelId:     oldProp.TransferChannelId,
		Validators:            validators,
		BlacklistedValidators: blacklistValidator,
		WithdrawalAccount:     oldProp.WithdrawalAccount,
		FeeAccount:            oldProp.FeeAccount,
		DelegationAccount:     oldProp.DelegationAccount,
		RedemptionAccount:     oldProp.RedemptionAccount,
		IbcDenom:              oldProp.IbcDenom,
		HostDenom:             oldProp.HostDenom,
		LastRedemptionRate:    oldProp.LastRedemptionRate,
		RedemptionRate:        oldProp.RedemptionRate,
		UnbondingFrequency:    sdk.NewIntFromUint64(oldProp.UnbondingFrequency),
		StakedBal:             sdk.NewIntFromUint64(oldProp.StakedBal),
		Address:               oldProp.Address,
	}
}

func convertToNewValidator(oldProp stakeibcv1types.Validator) stakeibctypes.Validator {
	internalExchangeRate := &stakeibctypes.ValidatorExchangeRate{
		InternalTokensToSharesRate: oldProp.InternalExchangeRate.InternalTokensToSharesRate,
		EpochNumber:                sdk.NewIntFromUint64(oldProp.InternalExchangeRate.EpochNumber),
	}
	return stakeibctypes.Validator{
		Name:                 oldProp.Name,
		Address:              oldProp.Address,
		Status:               stakeibctypes.Validator_ValidatorStatus(oldProp.Status),
		CommissionRate:       sdk.NewIntFromUint64(oldProp.CommissionRate),
		DelegationAmt:        sdk.NewIntFromUint64(oldProp.DelegationAmt),
		Weight:               sdk.NewIntFromUint64(oldProp.Weight),
		InternalExchangeRate: internalExchangeRate,
	}
}

func convertToNewStakeIbcParams(oldProp stakeibcv1types.Params) stakeibctypes.Params {
	newParam := stakeibctypes.Params{
		RewardsInterval:                  sdk.NewIntFromUint64(oldProp.RewardsInterval),
		DelegateInterval:                 sdk.NewIntFromUint64(oldProp.DelegateInterval),
		DepositInterval:                  sdk.NewIntFromUint64(oldProp.DepositInterval),
		RedemptionRateInterval:           sdk.NewIntFromUint64(oldProp.RedemptionRateInterval),
		StrideCommission:                 sdk.NewIntFromUint64(oldProp.StrideCommission),
		ZoneComAddress:                   oldProp.ZoneComAddress,
		ReinvestInterval:                 sdk.NewIntFromUint64(oldProp.ReinvestInterval),
		ValidatorRebalancingThreshold:    sdk.NewIntFromUint64(oldProp.ValidatorRebalancingThreshold),
		IcaTimeoutNanos:                  sdk.NewIntFromUint64(oldProp.IcaTimeoutNanos),
		BufferSize:                       sdk.NewIntFromUint64(oldProp.BufferSize),
		IbcTimeoutBlocks:                 sdk.NewIntFromUint64(oldProp.IbcTimeoutBlocks),
		FeeTransferTimeoutNanos:          sdk.NewIntFromUint64(oldProp.FeeTransferTimeoutNanos),
		MaxStakeIcaCallsPerEpoch:         sdk.NewIntFromUint64(oldProp.MaxStakeIcaCallsPerEpoch),
		SafetyMinRedemptionRateThreshold: sdk.NewIntFromUint64(oldProp.SafetyMinRedemptionRateThreshold),
		SafetyMaxRedemptionRateThreshold: sdk.NewIntFromUint64(oldProp.SafetyMaxRedemptionRateThreshold),
		IbcTransferTimeoutNanos:          sdk.NewIntFromUint64(oldProp.IbcTransferTimeoutNanos),
		SafetyNumValidators:              sdk.NewIntFromUint64(oldProp.SafetyNumValidators),
	}
	return newParam
}
