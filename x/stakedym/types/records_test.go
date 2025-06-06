package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v27/x/stakedym/types"
)

func TestValidateDelegationRecordGenesis(t *testing.T) {
	testCases := []struct {
		name              string
		delegationRecords []types.DelegationRecord
		expectedError     string
	}{
		{
			name: "valid records",
			delegationRecords: []types.DelegationRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1)},
				{Id: 2, NativeAmount: sdkmath.NewInt(2)},
				{Id: 3, NativeAmount: sdkmath.NewInt(3)},
			},
		},
		{
			name: "duplicate records",
			delegationRecords: []types.DelegationRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1)},
				{Id: 2, NativeAmount: sdkmath.NewInt(2)},
				{Id: 1, NativeAmount: sdkmath.NewInt(3)},
			},
			expectedError: "duplicate delegation record 1",
		},
		{
			name: "uninitialized native amount",
			delegationRecords: []types.DelegationRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1)},
				{Id: 2},
				{Id: 3, NativeAmount: sdkmath.NewInt(3)},
			},
			expectedError: "uninitialized native amount in delegation record 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateDelegationRecordGenesis(tc.delegationRecords)
			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}

func TestValidateUnbondingRecordGenesis(t *testing.T) {
	testCases := []struct {
		name             string
		unbondingRecords []types.UnbondingRecord
		expectedError    string
	}{
		{
			name: "valid records",
			unbondingRecords: []types.UnbondingRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{Id: 2, NativeAmount: sdkmath.NewInt(2), StTokenAmount: sdkmath.NewInt(2)},
				{Id: 3, NativeAmount: sdkmath.NewInt(3), StTokenAmount: sdkmath.NewInt(3)},
			},
		},
		{
			name: "duplicate records",
			unbondingRecords: []types.UnbondingRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{Id: 2, NativeAmount: sdkmath.NewInt(2), StTokenAmount: sdkmath.NewInt(2)},
				{Id: 1, NativeAmount: sdkmath.NewInt(3), StTokenAmount: sdkmath.NewInt(3)},
			},
			expectedError: "duplicate unbonding record 1",
		},
		{
			name: "uninitialized native amount",
			unbondingRecords: []types.UnbondingRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{Id: 2, StTokenAmount: sdkmath.NewInt(2)},
				{Id: 3, NativeAmount: sdkmath.NewInt(3), StTokenAmount: sdkmath.NewInt(3)},
			},
			expectedError: "uninitialized native amount in unbonding record 2",
		},
		{
			name: "uninitialized st amount",
			unbondingRecords: []types.UnbondingRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{Id: 2, NativeAmount: sdkmath.NewInt(2), StTokenAmount: sdkmath.NewInt(2)},
				{Id: 3, NativeAmount: sdkmath.NewInt(3)},
			},
			expectedError: "uninitialized sttoken amount in unbonding record 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateUnbondingRecordGenesis(tc.unbondingRecords)
			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}

func TestValidateRedemptionRecordGenesis(t *testing.T) {
	testCases := []struct {
		name             string
		redemptionRecord []types.RedemptionRecord
		expectedError    string
	}{
		{
			name: "valid records",
			redemptionRecord: []types.RedemptionRecord{
				{UnbondingRecordId: 1, Redeemer: "A", NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{UnbondingRecordId: 2, Redeemer: "A", NativeAmount: sdkmath.NewInt(2), StTokenAmount: sdkmath.NewInt(2)},
				{UnbondingRecordId: 3, Redeemer: "A", NativeAmount: sdkmath.NewInt(3), StTokenAmount: sdkmath.NewInt(3)},
			},
		},
		{
			name: "duplicate records",
			redemptionRecord: []types.RedemptionRecord{
				{UnbondingRecordId: 1, Redeemer: "A", NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{UnbondingRecordId: 2, Redeemer: "A", NativeAmount: sdkmath.NewInt(2), StTokenAmount: sdkmath.NewInt(2)},
				{UnbondingRecordId: 1, Redeemer: "A", NativeAmount: sdkmath.NewInt(3), StTokenAmount: sdkmath.NewInt(3)},
			},
			expectedError: "duplicate redemption record 1-A",
		},
		{
			name: "uninitialized native amount",
			redemptionRecord: []types.RedemptionRecord{
				{UnbondingRecordId: 1, Redeemer: "A", NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{UnbondingRecordId: 2, Redeemer: "A", StTokenAmount: sdkmath.NewInt(2)},
				{UnbondingRecordId: 3, Redeemer: "A", NativeAmount: sdkmath.NewInt(3), StTokenAmount: sdkmath.NewInt(3)},
			},
			expectedError: "uninitialized native amount in redemption record 2-A",
		},
		{
			name: "uninitialized st amount",
			redemptionRecord: []types.RedemptionRecord{
				{UnbondingRecordId: 1, Redeemer: "A", NativeAmount: sdkmath.NewInt(1), StTokenAmount: sdkmath.NewInt(1)},
				{UnbondingRecordId: 2, Redeemer: "A", NativeAmount: sdkmath.NewInt(2), StTokenAmount: sdkmath.NewInt(2)},
				{UnbondingRecordId: 3, Redeemer: "A", NativeAmount: sdkmath.NewInt(3)},
			},
			expectedError: "uninitialized sttoken amount in redemption record 3-A",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateRedemptionRecordGenesis(tc.redemptionRecord)
			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}

func TestValidateSlashRecordGenesis(t *testing.T) {
	testCases := []struct {
		name          string
		slashRecords  []types.SlashRecord
		expectedError string
	}{
		{
			name: "valid records",
			slashRecords: []types.SlashRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1)},
				{Id: 2, NativeAmount: sdkmath.NewInt(2)},
				{Id: 3, NativeAmount: sdkmath.NewInt(3)},
			},
		},
		{
			name: "duplicate records",
			slashRecords: []types.SlashRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1)},
				{Id: 2, NativeAmount: sdkmath.NewInt(2)},
				{Id: 1, NativeAmount: sdkmath.NewInt(3)},
			},
			expectedError: "duplicate slash record 1",
		},
		{
			name: "uninitialized native amount",
			slashRecords: []types.SlashRecord{
				{Id: 1, NativeAmount: sdkmath.NewInt(1)},
				{Id: 2},
				{Id: 3, NativeAmount: sdkmath.NewInt(3)},
			},
			expectedError: "uninitialized native amount in slash record 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateSlashRecordGenesis(tc.slashRecords)
			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}
