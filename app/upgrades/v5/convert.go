package v5

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	claimv1types "github.com/Stride-Labs/stride/v4/x/claim/types/v1"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	recordv1types "github.com/Stride-Labs/stride/v4/x/records/types/v1"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibcv1types "github.com/Stride-Labs/stride/v4/x/stakeibc/types/v1"
)

func convertToNewClaimParams(oldProp claimv1types.Params) claimtypes.Params {
	var newParams claimtypes.Params
	for _, airDrop := range(oldProp.Airdrops) {
		newAirDrop := claimtypes.Airdrop{
			AirdropIdentifier: airDrop.AirdropIdentifier,
			AirdropStartTime: airDrop.AirdropStartTime,
			AirdropDuration: airDrop.AirdropDuration,
			ClaimDenom: airDrop.ClaimDenom,
			DistributorAddress: airDrop.DistributorAddress,
			ClaimedSoFar: sdk.NewInt(airDrop.ClaimedSoFar),
		}
		newParams.Airdrops = append(newParams.Airdrops, &newAirDrop)
	}
	return newParams
}

func convertToNewUserRedemptionRecord(oldProp recordv1types.UserRedemptionRecord) recordtypes.UserRedemptionRecord {
	return recordtypes.UserRedemptionRecord{
		Id: oldProp.Id,
		Sender: oldProp.Sender,
		Receiver: oldProp.Receiver,
		Amount: sdk.NewIntFromUint64(oldProp.Amount),
		Denom: oldProp.Denom,
		HostZoneId: oldProp.HostZoneId,
		EpochNumber: oldProp.EpochNumber,
		ClaimIsPending: oldProp.ClaimIsPending,
	}
}

func convertToNewDepositRecord(oldProp recordv1types.DepositRecord) recordtypes.DepositRecord {
	return recordtypes.DepositRecord{
		Id: oldProp.Id,
		Amount: sdk.NewInt(oldProp.Amount),
		Denom: oldProp.Denom,
		HostZoneId: oldProp.HostZoneId,
		Status: recordtypes.DepositRecord_Status(oldProp.Status),
		DepositEpochNumber: oldProp.DepositEpochNumber,
		Source: recordtypes.DepositRecord_Source(oldProp.Source),
	}
}

func convertToNewEpochUnbondingRecord(oldProp recordv1types.EpochUnbondingRecord) recordtypes.EpochUnbondingRecord {
	var epochUnbondingRecord recordtypes.EpochUnbondingRecord
	for _, hz := range(oldProp.HostZoneUnbondings) {
		newHz := recordtypes.HostZoneUnbonding{
			StTokenAmount: sdk.NewIntFromUint64(hz.StTokenAmount),
			NativeTokenAmount: sdk.NewIntFromUint64(hz.NativeTokenAmount),
			Denom: hz.Denom,
			HostZoneId: hz.HostZoneId,
			UnbondingTime: hz.UnbondingTime,
			Status: recordtypes.HostZoneUnbonding_Status(hz.Status),
			UserRedemptionRecords: hz.UserRedemptionRecords,
		}
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &newHz)
	}
	return epochUnbondingRecord
}

func convertToNewHostZone(oldProp stakeibcv1types.HostZone) stakeibctypes.HostZone {
	var validators []*stakeibctypes.Validator
	var blacklistValidator []*stakeibctypes.Validator

	for _, val := range(oldProp.Validators) {
		newVal := convertToNewValidator(*val)
		validators = append(validators, &newVal)
	}

	for _, val := range(oldProp.BlacklistedValidators) {
		newVal := convertToNewValidator(*val)
		blacklistValidator = append(blacklistValidator, &newVal)
	}
	return stakeibctypes.HostZone{
		ChainId: oldProp.ChainId,
		ConnectionId: oldProp.ConnectionId,
		Bech32Prefix: oldProp.Bech32Prefix,
		TransferChannelId: oldProp.TransferChannelId,
		Validators: validators,
		BlacklistedValidators: blacklistValidator,
		WithdrawalAccount: oldProp.WithdrawalAccount,
		FeeAccount: oldProp.FeeAccount,
		DelegationAccount: oldProp.DelegationAccount,
		RedemptionAccount: oldProp.RedemptionAccount,
		IbcDenom: oldProp.IbcDenom,
		HostDenom: oldProp.HostDenom,
		LastRedemptionRate: oldProp.LastRedemptionRate,
		RedemptionRate: oldProp.RedemptionRate,
		UnbondingFrequency: oldProp.UnbondingFrequency,
		StakedBal: sdk.NewIntFromUint64(oldProp.StakedBal),
		Address: oldProp.Address,
	}
}

func convertToNewValidator(oldProp stakeibcv1types.Validator) stakeibctypes.Validator {
	return stakeibctypes.Validator{
		Name: oldProp.Name,
		Address: oldProp.Address,
		Status: stakeibctypes.Validator_ValidatorStatus(oldProp.Status),
		CommissionRate: oldProp.CommissionRate,
		DelegationAmt: sdk.NewIntFromUint64(oldProp.DelegationAmt),
		Weight: oldProp.Weight,
		InternalExchangeRate: (*stakeibctypes.ValidatorExchangeRate)(oldProp.InternalExchangeRate),
	}
}

func convertToNewSplitDelegation(oldProp stakeibcv1types.SplitDelegation) stakeibctypes.SplitDelegation {
	return stakeibctypes.SplitDelegation{
		Validator: oldProp.Validator,
		Amount: math.NewIntFromUint64(oldProp.Amount),
	}
}

func convertToNewDelegateCallback(oldProp stakeibcv1types.DelegateCallback) stakeibctypes.DelegateCallback {
	var splitDelegations []*stakeibctypes.SplitDelegation

	for _, sd := range(oldProp.SplitDelegations) {
		newSd := convertToNewSplitDelegation(*sd)
		splitDelegations = append(splitDelegations, &newSd)
	}
	return stakeibctypes.DelegateCallback{
		HostZoneId: oldProp.HostZoneId,
		DepositRecordId: oldProp.DepositRecordId,
		SplitDelegations: splitDelegations,
	}
}

func convertToNewUndelegateCallback(oldProp stakeibcv1types.UndelegateCallback) stakeibctypes.UndelegateCallback {
	var splitDelegations []*stakeibctypes.SplitDelegation

	for _, sd := range(oldProp.SplitDelegations) {
		newSd := convertToNewSplitDelegation(*sd)
		splitDelegations = append(splitDelegations, &newSd)
	}
	return stakeibctypes.UndelegateCallback{
		HostZoneId: oldProp.HostZoneId,
		EpochUnbondingRecordIds: oldProp.EpochUnbondingRecordIds,
		SplitDelegations: splitDelegations,
	}
}

func convertToNewRebalancing(oldProp stakeibcv1types.Rebalancing) stakeibctypes.Rebalancing {
	return stakeibctypes.Rebalancing{
		SrcValidator: oldProp.SrcValidator,
		DstValidator: oldProp.DstValidator,
		Amt: math.NewIntFromUint64(oldProp.Amt),
	}
}

func convertToNewRebalanceCallback(oldProp stakeibcv1types.RebalanceCallback) stakeibctypes.RebalanceCallback {
	var rebalancings []*stakeibctypes.Rebalancing

	for _, rebalancing := range(oldProp.Rebalancings) {
		newRebalancing := convertToNewRebalancing(*rebalancing)
		rebalancings = append(rebalancings, &newRebalancing)
	}
	return stakeibctypes.RebalanceCallback{
		HostZoneId: oldProp.HostZoneId,
		Rebalancings: rebalancings,
	}
}