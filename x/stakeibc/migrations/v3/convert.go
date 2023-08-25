package v3

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oldstakeibctypes "github.com/Stride-Labs/stride/v13/x/stakeibc/migrations/v3/types"
	newstakeibctypes "github.com/Stride-Labs/stride/v13/x/stakeibc/types"
)

const (
	ValidatorSlashQueryThreshold uint64 = 1 // denominated in percentage of TVL (1 => 1%)
)

// Converts an old validator data type to the new schema
// Changes are as follows:
//   - Added SlashQueryProgressTracker field
//   - Added SlashQueryCheckpoint field
//   - Added DelegationsInProgress field
//   - Added SlashQueryInProgress field
//   - InternalExchangeRate is now a decimal named SharesToTokensRate
//   - DelegationAmt renamed to Delegation
func convertToNewValidator(oldValidator oldstakeibctypes.Validator, totalDelegations sdkmath.Int) newstakeibctypes.Validator {
	queryThreshold := sdk.NewDecWithPrec(int64(ValidatorSlashQueryThreshold), 2) // percentage
	slashQueryCheckpoint := queryThreshold.Mul(sdk.NewDecFromInt(totalDelegations)).TruncateInt()

	return newstakeibctypes.Validator{
		Name:                        oldValidator.Name,
		Address:                     oldValidator.Address,
		Weight:                      oldValidator.Weight,
		Delegation:                  oldValidator.DelegationAmt,
		SlashQueryProgressTracker:   sdkmath.ZeroInt(),
		SlashQueryCheckpoint:        slashQueryCheckpoint,
		SharesToTokensRate:          oldValidator.InternalExchangeRate.InternalTokensToSharesRate,
		DelegationChangesInProgress: 0,
		SlashQueryInProgress:        false,
	}
}

// Converts an old host zone data type to the new schema
// Changes are as follows:
//   - ICA Accounts are now strings
//   - Address has been renamed to DepositAddress
//   - UnbondingFrequency has been changed to UnbondingPeriod
//   - StakedBal has been renamed to TotalDelegations
//   - Removed blacklisted validators
func convertToNewHostZone(oldHostZone oldstakeibctypes.HostZone) newstakeibctypes.HostZone {
	var validators []*newstakeibctypes.Validator
	for _, oldValidator := range oldHostZone.Validators {
		newValidator := convertToNewValidator(*oldValidator, oldHostZone.StakedBal)
		validators = append(validators, &newValidator)
	}

	return newstakeibctypes.HostZone{
		ChainId:               oldHostZone.ChainId,
		ConnectionId:          oldHostZone.ConnectionId,
		Bech32Prefix:          oldHostZone.Bech32Prefix,
		TransferChannelId:     oldHostZone.TransferChannelId,
		IbcDenom:              oldHostZone.IbcDenom,
		HostDenom:             oldHostZone.HostDenom,
		UnbondingPeriod:       (oldHostZone.UnbondingFrequency - 1) * 7,
		Validators:            validators,
		DepositAddress:        oldHostZone.Address,
		WithdrawalIcaAddress:  oldHostZone.WithdrawalAccount.GetAddress(),
		FeeIcaAddress:         oldHostZone.FeeAccount.GetAddress(),
		DelegationIcaAddress:  oldHostZone.DelegationAccount.GetAddress(),
		RedemptionIcaAddress:  oldHostZone.RedemptionAccount.GetAddress(),
		TotalDelegations:      oldHostZone.StakedBal,
		LastRedemptionRate:    oldHostZone.LastRedemptionRate,
		RedemptionRate:        oldHostZone.RedemptionRate,
		MinRedemptionRate:     oldHostZone.MinRedemptionRate,
		MaxRedemptionRate:     oldHostZone.MaxRedemptionRate,
		LsmLiquidStakeEnabled: false,
		Halted:                oldHostZone.Halted,
	}
}
