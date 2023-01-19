package v2

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	oldstakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/migrations/v2/types"
	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func TestConvertToNewValidator(t *testing.T) {
	name := "name"
	address := "address"
	commmissionRate := uint64(2)
	weight := uint64(3)
	epochNumber := uint64(4)
	tokensToShares := sdk.NewDec(5)

	// Only the DelegationAmt field of the Validator should change
	oldValidator := oldstakeibctypes.Validator{
		Name:           name,
		Address:        address,
		Status:         oldstakeibctypes.Validator_ACTIVE,
		CommissionRate: commmissionRate,
		DelegationAmt:  uint64(1),
		Weight:         weight,
		InternalExchangeRate: &oldstakeibctypes.ValidatorExchangeRate{
			InternalTokensToSharesRate: tokensToShares,
			EpochNumber:                epochNumber,
		},
	}
	expectedNewValidator := stakeibctypes.Validator{
		Name:           name,
		Address:        address,
		Status:         stakeibctypes.Validator_ACTIVE,
		CommissionRate: commmissionRate,
		DelegationAmt:  sdkmath.NewInt(1),
		Weight:         weight,
		InternalExchangeRate: &stakeibctypes.ValidatorExchangeRate{
			InternalTokensToSharesRate: tokensToShares,
			EpochNumber:                epochNumber,
		},
	}

	actualNewValidator := convertToNewValidator(oldValidator)
	require.Equal(t, expectedNewValidator, actualNewValidator)
}

func TestConvertToNewICAAccount(t *testing.T) {
	oldAccount := oldstakeibctypes.ICAAccount{Address: "address", Target: oldstakeibctypes.ICAAccountType_FEE}
	expectedNewAccount := stakeibctypes.ICAAccount{Address: "address", Target: stakeibctypes.ICAAccountType_FEE}
	actualNewAccount := convertToNewICAAccount(&oldAccount)
	require.Equal(t, expectedNewAccount, *actualNewAccount)
}

func TestConvertToNewICAAccount_Nil(t *testing.T) {
	actualNewAccount := convertToNewICAAccount(nil)
	require.Nil(t, actualNewAccount)
}

func TestConvertToNewHostZone(t *testing.T) {
	chainId := "chain"
	connectionId := "connection"
	bechPrefix := "bech"
	channelId := "channel"
	valAddress := "val"
	blacklistedValAddress := "black_val"
	withdrawalAddress := "withdrawal"
	feeAddress := "fee"
	delegationAddress := "delegation"
	redemptionAddress := "redemption"
	ibcDenom := "ibc"
	hostDenom := "host"
	redemptionRate := sdk.NewDec(1)
	lastRedemptionRate := sdk.NewDec(2)
	unbondingFrequency := uint64(3)
	hostAddress := "address"

	// The stakedBal field and validators get updated on the host zone
	oldHostZone := oldstakeibctypes.HostZone{
		ChainId:           chainId,
		ConnectionId:      connectionId,
		Bech32Prefix:      bechPrefix,
		TransferChannelId: channelId,
		Validators: []*oldstakeibctypes.Validator{
			{Address: valAddress, DelegationAmt: uint64(1)},
		},
		BlacklistedValidators: []*oldstakeibctypes.Validator{
			{Address: blacklistedValAddress, DelegationAmt: uint64(2)},
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
		StakedBal:          uint64(3),
		Address:            hostAddress,
	}

	expectedNewHostZone := stakeibctypes.HostZone{
		ChainId:           chainId,
		ConnectionId:      connectionId,
		Bech32Prefix:      bechPrefix,
		TransferChannelId: channelId,
		Validators: []*stakeibctypes.Validator{
			{Address: valAddress, DelegationAmt: sdkmath.NewInt(1)},
		},
		BlacklistedValidators: []*stakeibctypes.Validator{
			{Address: blacklistedValAddress, DelegationAmt: sdkmath.NewInt(2)},
		},
		WithdrawalAccount: &stakeibctypes.ICAAccount{
			Address: withdrawalAddress, Target: stakeibctypes.ICAAccountType_WITHDRAWAL,
		},
		FeeAccount: &stakeibctypes.ICAAccount{
			Address: feeAddress, Target: stakeibctypes.ICAAccountType_FEE,
		},
		DelegationAccount: &stakeibctypes.ICAAccount{
			Address: delegationAddress, Target: stakeibctypes.ICAAccountType_DELEGATION,
		},
		RedemptionAccount: &stakeibctypes.ICAAccount{
			Address: redemptionAddress, Target: stakeibctypes.ICAAccountType_REDEMPTION,
		},
		IbcDenom:           ibcDenom,
		HostDenom:          hostDenom,
		RedemptionRate:     redemptionRate,
		LastRedemptionRate: lastRedemptionRate,
		UnbondingFrequency: unbondingFrequency,
		StakedBal:          sdkmath.NewInt(3),
		Address:            hostAddress,
	}

	actualNewHostZone := convertToNewHostZone(oldHostZone)
	require.Equal(t, expectedNewHostZone, actualNewHostZone)
}
