package types

import (
	sdkmath "cosmossdk.io/math"
)

// DefaultGenesis returns the default genesis state
// Only the host zone and accumulator record are needed at default genesis,
// other record should be empty
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		HostZone: HostZone{
			ChainId:                DymensionChainId,
			NativeTokenDenom:       DymensionNativeTokenDenom,
			NativeTokenIbcDenom:    DymensionNativeTokenIBCDenom,
			TransferChannelId:      StrideToDymensionTransferChannelId,
			UnbondingPeriodSeconds: DymensionUnbondingPeriodSeconds,

			// on dymension
			DelegationAddress: DelegationAddressOnDymension,
			RewardAddress:     RewardAddressOnDymension,

			// functional accounts on stride
			DepositAddress:    DepositAddress,
			RedemptionAddress: RedemptionAddress,
			ClaimAddress:      ClaimAddress,

			// management accounts on stride
			SafeAddressOnStride:     SafeAddressOnStride,
			OperatorAddressOnStride: OperatorAddressOnStride,

			RedemptionRate:         sdkmath.LegacyOneDec(),
			LastRedemptionRate:     sdkmath.LegacyOneDec(),
			MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.95"),
			MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.1"),
			MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.95"),
			MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1"),
			DelegatedBalance:       sdkmath.ZeroInt(),
			Halted:                 false,
		},
		UnbondingRecords: []UnbondingRecord{
			{
				Id:            1,
				Status:        ACCUMULATING_REDEMPTIONS,
				NativeAmount:  sdkmath.ZeroInt(),
				StTokenAmount: sdkmath.ZeroInt(),
			},
		},
	}
}

// Validates the host zone and records in the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.HostZone.ValidateGenesis(); err != nil {
		return err
	}
	if err := ValidateDelegationRecordGenesis(gs.DelegationRecords); err != nil {
		return err
	}
	if err := ValidateUnbondingRecordGenesis(gs.UnbondingRecords); err != nil {
		return err
	}
	if err := ValidateRedemptionRecordGenesis(gs.RedemptionRecords); err != nil {
		return err
	}
	if err := ValidateSlashRecordGenesis(gs.SlashRecords); err != nil {
		return err
	}
	return nil
}
