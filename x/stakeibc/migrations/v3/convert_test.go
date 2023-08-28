package v3

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	oldstakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/migrations/v3/types"
	newstakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func TestConvertToNewValidator(t *testing.T) {
	name := "name"
	address := "address"
	weight := uint64(3)
	delegation := sdk.NewInt(4)
	sharesToTokensRate := sdk.NewDec(5)
	tvl := sdk.NewInt(1_000_000)
	slashCheckpoint := sdk.NewInt(10_000) // 1% of TVL

	// First convert a validator with no exchange rate
	// It should get filled in with the default
	oldValidator := oldstakeibctypes.Validator{
		Name:          name,
		Address:       address,
		DelegationAmt: delegation,
		Weight:        weight,
	}
	expectedNewValidator := newstakeibctypes.Validator{
		Name:                        name,
		Address:                     address,
		Weight:                      weight,
		Delegation:                  delegation,
		SlashQueryProgressTracker:   sdkmath.ZeroInt(),
		SlashQueryCheckpoint:        slashCheckpoint,
		DelegationChangesInProgress: 0,
		SharesToTokensRate:          DefaultExchangeRate,
		SlashQueryInProgress:        false,
	}

	actualNewValidator := convertToNewValidator(oldValidator, tvl)
	require.Equal(t, expectedNewValidator, actualNewValidator)

	// Then add an exchange rate and convert again
	oldValidator.InternalExchangeRate = &oldstakeibctypes.ValidatorExchangeRate{
		InternalTokensToSharesRate: sharesToTokensRate,
		EpochNumber:                1,
	}
	expectedNewValidator.SharesToTokensRate = sharesToTokensRate

	actualNewValidator = convertToNewValidator(oldValidator, tvl)
	require.Equal(t, expectedNewValidator, actualNewValidator)
}

func TestConvertToNewHostZone(t *testing.T) {
	chainId := "chain"
	connectionId := "connection"
	bechPrefix := "bech"
	channelId := "channel"
	ibcDenom := "ibc"
	hostDenom := "host"

	depositAddress := "address"
	withdrawalAddress := "withdrawal"
	feeAddress := "fee"
	delegationAddress := "delegation"
	redemptionAddress := "redemption"

	redemptionRate := sdk.NewDec(1)
	lastRedemptionRate := sdk.NewDec(2)
	minRedemptionRate := sdk.MustNewDecFromStr("0.95")
	maxRedemptionRate := sdk.MustNewDecFromStr("1.25")
	unbondingFrequency := uint64(4)
	unbondingPeriod := uint64(21)

	halted := true

	valAddress := "val"
	valDelegation := sdk.NewInt(5)
	valWeight := uint64(6)
	totalDelegations := sdk.NewInt(1_000_000)
	slashCheckpoint := sdk.NewInt(10_000) // 1% of TVL
	sharesToTokensRate := sdk.MustNewDecFromStr("0.99")

	// The stakedBal field and validators get updated on the host zone
	oldHostZone := oldstakeibctypes.HostZone{
		ChainId:           chainId,
		ConnectionId:      connectionId,
		Bech32Prefix:      bechPrefix,
		TransferChannelId: channelId,
		Validators: []*oldstakeibctypes.Validator{
			{
				// Validator with an exchange rate
				Address:       valAddress,
				DelegationAmt: valDelegation,
				Weight:        valWeight,
				InternalExchangeRate: &oldstakeibctypes.ValidatorExchangeRate{
					InternalTokensToSharesRate: sharesToTokensRate,
					EpochNumber:                1,
				},
			},
			{
				// Validator without an exchange rate
				Address:              valAddress,
				DelegationAmt:        valDelegation,
				Weight:               valWeight,
				InternalExchangeRate: nil,
			},
		},
		BlacklistedValidators: []*oldstakeibctypes.Validator{
			{Address: "black", DelegationAmt: valDelegation},
		},
		WithdrawalAccount: &oldstakeibctypes.ICAAccount{
			Address: withdrawalAddress, Target: oldstakeibctypes.ICAAccountType_WITHDRAWAL,
		},
		FeeAccount: &oldstakeibctypes.ICAAccount{
			Address: feeAddress, Target: oldstakeibctypes.ICAAccountType_FEE,
		},
		DelegationAccount: &oldstakeibctypes.ICAAccount{
			Address: delegationAddress, Target: oldstakeibctypes.ICAAccountType_DELEGATION,
		},
		RedemptionAccount: &oldstakeibctypes.ICAAccount{
			Address: redemptionAddress, Target: oldstakeibctypes.ICAAccountType_REDEMPTION,
		},
		IbcDenom:           ibcDenom,
		HostDenom:          hostDenom,
		RedemptionRate:     redemptionRate,
		LastRedemptionRate: lastRedemptionRate,
		UnbondingFrequency: unbondingFrequency,
		StakedBal:          totalDelegations,
		Address:            depositAddress,
		Halted:             halted,
		MinRedemptionRate:  minRedemptionRate,
		MaxRedemptionRate:  maxRedemptionRate,
	}

	expectedNewHostZone := newstakeibctypes.HostZone{
		ChainId:           chainId,
		ConnectionId:      connectionId,
		Bech32Prefix:      bechPrefix,
		TransferChannelId: channelId,
		IbcDenom:          ibcDenom,
		HostDenom:         hostDenom,
		UnbondingPeriod:   unbondingPeriod,
		Validators: []*newstakeibctypes.Validator{
			{
				// Validator with an exchange rate
				Address:                     valAddress,
				Weight:                      valWeight,
				Delegation:                  valDelegation,
				SlashQueryProgressTracker:   sdkmath.ZeroInt(),
				SlashQueryCheckpoint:        slashCheckpoint,
				SharesToTokensRate:          sharesToTokensRate,
				DelegationChangesInProgress: 0,
				SlashQueryInProgress:        false,
			},
			{
				// Validator with nil exchange rate coalesced with 1
				Address:                     valAddress,
				Weight:                      valWeight,
				Delegation:                  valDelegation,
				SlashQueryProgressTracker:   sdkmath.ZeroInt(),
				SlashQueryCheckpoint:        slashCheckpoint,
				SharesToTokensRate:          DefaultExchangeRate,
				DelegationChangesInProgress: 0,
				SlashQueryInProgress:        false,
			},
		},
		DepositAddress:        depositAddress,
		WithdrawalIcaAddress:  withdrawalAddress,
		FeeIcaAddress:         feeAddress,
		DelegationIcaAddress:  delegationAddress,
		RedemptionIcaAddress:  redemptionAddress,
		TotalDelegations:      totalDelegations,
		RedemptionRate:        redemptionRate,
		LastRedemptionRate:    lastRedemptionRate,
		MinRedemptionRate:     minRedemptionRate,
		MaxRedemptionRate:     maxRedemptionRate,
		LsmLiquidStakeEnabled: false,
		Halted:                halted,
	}

	actualNewHostZone := convertToNewHostZone(oldHostZone)
	require.Equal(t, expectedNewHostZone, actualNewHostZone)
}
