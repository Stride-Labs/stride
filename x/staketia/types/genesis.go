package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state
// Only the host zone and accumulator record are needed at default genesis,
// other record should be empty
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		HostZone: HostZone{
			ChainId:             CelestiaChainId,
			NativeTokenDenom:    CelestiaNativeTokenDenom,
			NativeTokenIbcDenom: CelestiaNativeTokenIBCDenom,
			TransferChannelId:   CelestiaChannelId,

			DelegationAddress: CelestiaDelegationAddress,
			RewardAddress:     CelestiaRewardsAddress,
			DepositAddress:    CelestiaDepositAddress,
			RedemptionAddress: CelestiaRedemptionAddress,
			ClaimAddress:      CelestiaClaimAddress,
			OperatorAddress:   CelestiaOperatorAddress,
			SafeAddress:       CelestiaSafeAddress,

			RedemptionRate:         sdk.OneDec(),
			LastRedemptionRate:     sdk.OneDec(),
			MinRedemptionRate:      sdk.MustNewDecFromStr("0.95"),
			MaxRedemptionRate:      sdk.MustNewDecFromStr("1.1"),
			MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.95"),
			MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
			DelegatedBalance:       sdkmath.ZeroInt(),
			Halted:                 false,
		},
		UnbondingRecords: []UnbondingRecord{
			{
				Id:            0,
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
