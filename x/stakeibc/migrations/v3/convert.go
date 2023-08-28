package v3

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oldstakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/migrations/v3/types"
	newstakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

var (
	// The threshold, denominated in percentage of TVL, of when a slash query should
	// be submitted (1 => 1%)
	ValidatorSlashQueryThreshold uint64 = 1
	// The exchange rate here does not matter since it will be updated after the slash query
	// Setting it to this value makes it easier to verify that we've submitted the query
	DefaultExchangeRate = sdk.MustNewDecFromStr("0.999999999999999999")
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

	// Note: The old name of "TokensToShares" was slightly misleading - it represents the conversion of shares to tokens
	sharesToTokensRate := DefaultExchangeRate
	if oldValidator.InternalExchangeRate != nil && !oldValidator.InternalExchangeRate.InternalTokensToSharesRate.IsNil() {
		sharesToTokensRate = oldValidator.InternalExchangeRate.InternalTokensToSharesRate
	}

	return newstakeibctypes.Validator{
		Name:                        oldValidator.Name,
		Address:                     oldValidator.Address,
		Weight:                      oldValidator.Weight,
		Delegation:                  oldValidator.DelegationAmt,
		SlashQueryProgressTracker:   sdkmath.ZeroInt(),
		SlashQueryCheckpoint:        slashQueryCheckpoint,
		SharesToTokensRate:          sharesToTokensRate,
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
		Bech32Prefix:          oldHostZone.Bech32Prefix,
		ConnectionId:          oldHostZone.ConnectionId,
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
