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
			ChainId:                CelestiaChainId,
			NativeTokenDenom:       CelestiaNativeTokenDenom,
			NativeTokenIbcDenom:    CelestiaNativeTokenIBCDenom,
			TransferChannelId:      StrideToCelestiaTransferChannelId,
			UnbondingPeriodSeconds: CelestiaUnbondingPeriodSeconds,

			// on celestia
			DelegationAddress: DelegationAddressOnCelestia,
			RewardAddress:     RewardAddressOnCelestia,

			// functional accounts on stride
			DepositAddress:    DepositAddress,
			RedemptionAddress: RedemptionAddress,
			ClaimAddress:      ClaimAddress,

			// management accounts on stride
			SafeAddressOnStride:     SafeAddressOnStride,
			OperatorAddressOnStride: OperatorAddressOnStride,

			RemainingDelegatedBalance: sdkmath.ZeroInt(),
			Halted:                    false,
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
