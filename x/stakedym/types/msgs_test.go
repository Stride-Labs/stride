package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	"github.com/Stride-Labs/stride/v26/x/stakedym/types"
)

// ----------------------------------------------
//               MsgLiquidStake
// ----------------------------------------------

func TestMsgLiquidStake_ValidateBasic(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgLiquidStake
		err  error
	}{
		{
			name: "invalid address",
			msg: types.MsgLiquidStake{
				Staker:       "invalid_address",
				NativeAmount: sdkmath.NewInt(1000000),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address: wrong chain's bech32prefix",
			msg: types.MsgLiquidStake{
				Staker:       "tia1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				NativeAmount: sdkmath.NewInt(1000000),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid inputs",
			msg: types.MsgLiquidStake{
				Staker:       validAddress,
				NativeAmount: sdkmath.NewInt(1200000),
			},
		},
		{
			name: "amount below threshold",
			msg: types.MsgLiquidStake{
				Staker:       validAddress,
				NativeAmount: sdkmath.NewInt(20000),
			},
			err: types.ErrInvalidAmountBelowMinimum,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// check validatebasic()
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgLiquidStake_GetSignBytes(t *testing.T) {
	addr := "stride1v9jxgu33kfsgr5"
	msg := types.NewMsgLiquidStake(addr, sdkmath.NewInt(1000))
	res := msg.GetSignBytes()

	expected := `{"type":"stakedym/MsgLiquidStake","value":{"native_amount":"1000","staker":"stride1v9jxgu33kfsgr5"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgRedeemStake
// ----------------------------------------------

func TestMsgRedeemStake_ValidateBasic(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgRedeemStake
		err  error
	}{
		{
			name: "success",
			msg: types.MsgRedeemStake{
				Redeemer:      validAddress,
				StTokenAmount: sdkmath.NewInt(1000000),
			},
		},
		{
			name: "invalid creator",
			msg: types.MsgRedeemStake{
				Redeemer:      "invalid_address",
				StTokenAmount: sdkmath.NewInt(1000000),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "amount below threshold",
			msg: types.MsgRedeemStake{
				Redeemer:      validAddress,
				StTokenAmount: sdkmath.NewInt(20000),
			},
			err: types.ErrInvalidAmountBelowMinimum,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgRedeemStake_GetSignBytes(t *testing.T) {
	addr := "stride1v9jxgu33kfsgr5"
	msg := types.NewMsgRedeemStake(addr, sdkmath.NewInt(1000000))
	res := msg.GetSignBytes()

	expected := `{"type":"stakedym/MsgRedeemStake","value":{"redeemer":"stride1v9jxgu33kfsgr5","st_token_amount":"1000000"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//             MsgConfirmDelegation
// ----------------------------------------------

func TestMsgConfirmDelegation_ValidateBasic(t *testing.T) {
	validTxHash := "BBD978ADDBF580AC2981E351A3EA34AA9D7B57631E9CE21C27C2C63A5B13BDA9"
	validRecordId := uint64(35)
	validAddress, _ := apptesting.GenerateTestAddrs()

	tests := []struct {
		name          string
		msg           types.MsgConfirmDelegation
		expectedError string
	}{
		{
			name: "success",
			msg: types.MsgConfirmDelegation{
				Operator: validAddress,
				RecordId: validRecordId,
				TxHash:   validTxHash,
			},
		},
		{
			name: "empty tx hash",
			msg: types.MsgConfirmDelegation{
				Operator: validAddress,
				RecordId: validRecordId,
				TxHash:   "",
			},
			expectedError: "tx hash is empty",
		},
		{
			name: "invalid tx hash",
			msg: types.MsgConfirmDelegation{
				Operator: validAddress,
				RecordId: validRecordId,
				TxHash:   "invalid_tx-hash",
			},
			expectedError: "tx hash is invalid",
		},
		{
			name: "invalid sender",
			msg: types.MsgConfirmDelegation{
				Operator: "strideinvalid",
				RecordId: validRecordId,
				TxHash:   validTxHash,
			},
			expectedError: "invalid address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectedError)
			}
		})
	}
}

func TestMsgConfirmDelegation_GetSignBytes(t *testing.T) {
	addr := "stride1v9jxgu33kfsgr5"
	msg := types.NewMsgConfirmDelegation(addr, 100, "valid_hash")
	res := msg.GetSignBytes()

	expected := `{"type":"stakedym/MsgConfirmDelegation","value":{"operator":"stride1v9jxgu33kfsgr5","record_id":"100","tx_hash":"valid_hash"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//           MsgConfirmUndelegation
// ----------------------------------------------

func TestMsgConfirmUndelegation_ValidateBasic(t *testing.T) {
	validTxHash := "BBD978ADDBF580AC2981E351A3EA34AA9D7B57631E9CE21C27C2C63A5B13BDA9"
	validRecordId := uint64(35)
	validAddress, _ := apptesting.GenerateTestAddrs()

	tests := []struct {
		name          string
		msg           types.MsgConfirmUndelegation
		expectedError string
	}{
		{
			name: "success",
			msg: types.MsgConfirmUndelegation{
				Operator: validAddress,
				RecordId: validRecordId,
				TxHash:   validTxHash,
			},
		},
		{
			name: "empty tx hash",
			msg: types.MsgConfirmUndelegation{
				Operator: validAddress,
				RecordId: validRecordId,
				TxHash:   "",
			},
			expectedError: "tx hash is empty",
		},
		{
			name: "invalid tx hash",
			msg: types.MsgConfirmUndelegation{
				Operator: validAddress,
				RecordId: validRecordId,
				TxHash:   "invalid_tx-hash",
			},
			expectedError: "tx hash is invalid",
		},
		{
			name: "invalid sender",
			msg: types.MsgConfirmUndelegation{
				Operator: "strideinvalid",
				RecordId: validRecordId,
				TxHash:   validTxHash,
			},
			expectedError: "invalid address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectedError)
			}
		})
	}
}

func TestMsgConfirmUndelegation_GetSignBytes(t *testing.T) {
	addr := "stride1v9jxgu33kfsgr5"
	msg := types.NewMsgConfirmUndelegation(addr, 100, "valid_hash")
	res := msg.GetSignBytes()

	expected := `{"type":"stakedym/MsgConfirmUndelegation","value":{"operator":"stride1v9jxgu33kfsgr5","record_id":"100","tx_hash":"valid_hash"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgConfirmUnbondedTokenSweep
// ----------------------------------------------

func TestMsgConfirmUnbondedTokenSweep_ValidateBasic(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgConfirmUnbondedTokenSweep
		err  error
	}{
		{
			name: "success",
			msg: types.MsgConfirmUnbondedTokenSweep{
				Operator: validAddress,
				RecordId: 35,
				TxHash:   "BBD978ADDBF580AC2981E351A3EA34AA9D7B57631E9CE21C27C2C63A5B13BDA9",
			},
		},
		{
			name: "empty tx hash",
			msg: types.MsgConfirmUnbondedTokenSweep{
				Operator: validAddress,
				RecordId: 35,
				TxHash:   "",
			},
			err: sdkerrors.ErrTxDecode,
		},
		{
			name: "invalid tx hash",
			msg: types.MsgConfirmUnbondedTokenSweep{
				Operator: validAddress,
				RecordId: 35,
				TxHash:   "invalid_tx-hash",
			},
			err: sdkerrors.ErrTxDecode,
		},
		{
			name: "invalid sender",
			msg: types.MsgConfirmUnbondedTokenSweep{
				Operator: "strideinvalid",
				RecordId: 35,
				TxHash:   "BBD978ADDBF580AC2981E351A3EA34AA9D7B57631E9CE21C27C2C63A5B13BDA9",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgConfirmUnbondedTokenSweep_GetSignBytes(t *testing.T) {
	addr := "stride1v9jxgu33kfsgr5"
	msg := types.NewMsgConfirmUnbondedTokenSweep(addr, 100, "valid_hash")
	res := msg.GetSignBytes()

	expected := `{"type":"stakedym/MsgConfirmUnbondedTokenSweep","value":{"operator":"stride1v9jxgu33kfsgr5","record_id":"100","tx_hash":"valid_hash"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//         MsgAdjustDelegatedBalance
// ----------------------------------------------

func TestMsgAdjustDelegatedBalance_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validValidatorAddress := "valoper"
	validDelegationOffset := sdkmath.NewInt(10)

	tests := []struct {
		name string
		msg  types.MsgAdjustDelegatedBalance
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgAdjustDelegatedBalance{
				Operator:         validAddress,
				DelegationOffset: validDelegationOffset,
				ValidatorAddress: validValidatorAddress,
			},
		},
		{
			name: "successful message, negative offset",
			msg: types.MsgAdjustDelegatedBalance{
				Operator:         validAddress,
				DelegationOffset: sdkmath.NewInt(-1),
				ValidatorAddress: validValidatorAddress,
			},
		},
		{
			name: "invalid signer address",
			msg: types.MsgAdjustDelegatedBalance{
				Operator:         invalidAddress,
				DelegationOffset: validDelegationOffset,
				ValidatorAddress: validValidatorAddress,
			},
			err: "invalid address",
		},
		{
			name: "invalid delegation offset",
			msg: types.MsgAdjustDelegatedBalance{
				Operator:         validAddress,
				ValidatorAddress: validValidatorAddress,
			},
			err: "delegation offset must be specified",
		},
		{
			name: "invalid validator address",
			msg: types.MsgAdjustDelegatedBalance{
				Operator:         validAddress,
				DelegationOffset: validDelegationOffset,
				ValidatorAddress: "",
			},
			err: "validator address must be specified",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddress)

				require.Equal(t, test.msg.Type(), "adjust_delegated_balance", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}

// ----------------------------------------------
//        MsgUpdateInnerRedemptionRateBounds
// ----------------------------------------------

func TestMsgUpdateInnerRedemptionRateBounds_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validUpperBound := sdkmath.LegacyNewDec(2)
	validLowerBound := sdkmath.LegacyNewDec(1)
	invalidLowerBound := sdkmath.LegacyNewDec(2)

	tests := []struct {
		name string
		msg  types.MsgUpdateInnerRedemptionRateBounds
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgUpdateInnerRedemptionRateBounds{
				Creator:                validAdminAddress,
				MaxInnerRedemptionRate: validUpperBound,
				MinInnerRedemptionRate: validLowerBound,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgUpdateInnerRedemptionRateBounds{
				Creator:                invalidAddress,
				MaxInnerRedemptionRate: validUpperBound,
				MinInnerRedemptionRate: validLowerBound,
			},
			err: "invalid address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgUpdateInnerRedemptionRateBounds{
				Creator:                validNotAdminAddress,
				MaxInnerRedemptionRate: validUpperBound,
				MinInnerRedemptionRate: validLowerBound,
			},
			err: "not an admin: invalid address",
		},
		{
			name: "invalid bounds",
			msg: types.MsgUpdateInnerRedemptionRateBounds{
				Creator:                validAdminAddress,
				MaxInnerRedemptionRate: validUpperBound,
				MinInnerRedemptionRate: invalidLowerBound,
			},
			err: "invalid host zone redemption rate inner bounds",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)

				require.Equal(t, test.msg.MaxInnerRedemptionRate, validUpperBound, "MaxInnerRedemptionRate")
				require.Equal(t, test.msg.MinInnerRedemptionRate, validLowerBound, "MaxInnerRedemptionRate")
				require.Equal(t, test.msg.Type(), "redemption_rate_bounds", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}

// ----------------------------------------------
//              MsgResumeHostZone
// ----------------------------------------------

func TestMsgResumeHostZone_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	tests := []struct {
		name string
		msg  types.MsgResumeHostZone
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgResumeHostZone{
				Creator: validAdminAddress,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgResumeHostZone{
				Creator: invalidAddress,
			},
			err: "invalid address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgResumeHostZone{
				Creator: validNotAdminAddress,
			},
			err: "not an admin: invalid address",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)

				require.Equal(t, test.msg.Type(), "resume_host_zone", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}

// ----------------------------------------------
//          MsgRefreshRedemptionRate
// ----------------------------------------------

func TestMsgRefreshRedemptionRate_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validAddress, invalidAddress := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgRefreshRedemptionRate
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgRefreshRedemptionRate{
				Creator: validAddress,
			},
		},
		{
			name: "invalid signer address",
			msg: types.MsgRefreshRedemptionRate{
				Creator: invalidAddress,
			},
			err: "invalid address",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddress)

				require.Equal(t, test.msg.Type(), "refresh_redemption_rate", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}

// ----------------------------------------------
//          OverwriteDelegationRecord
// ----------------------------------------------

func TestMsgOverwriteDelegationRecord_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validTxHash := "69650DCD2D68BBC9310BA4B980187483BFC81452EC99C39DA45633F077157911"
	validAddress, invalidAddress := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgOverwriteDelegationRecord
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgOverwriteDelegationRecord{
				Creator: validAddress,
				DelegationRecord: &types.DelegationRecord{
					Id:           1,
					NativeAmount: sdkmath.NewInt(1),
					Status:       types.DELEGATION_QUEUE,
					TxHash:       validTxHash,
				},
			},
		},
		{
			name: "successful message with blank tx hash",
			msg: types.MsgOverwriteDelegationRecord{
				Creator: validAddress,
				DelegationRecord: &types.DelegationRecord{
					Id:           1,
					NativeAmount: sdkmath.NewInt(1),
					Status:       types.DELEGATION_QUEUE,
					TxHash:       "",
				},
			},
		},
		{
			name: "invalid signer address",
			msg: types.MsgOverwriteDelegationRecord{
				Creator: invalidAddress,
				DelegationRecord: &types.DelegationRecord{
					Id:           1,
					NativeAmount: sdkmath.NewInt(1),
					Status:       types.DELEGATION_QUEUE,
					TxHash:       validTxHash,
				},
			},
			err: "invalid address",
		},
		{
			name: "invalid native amount",
			msg: types.MsgOverwriteDelegationRecord{
				Creator: validAddress,
				DelegationRecord: &types.DelegationRecord{
					Id:           1,
					NativeAmount: sdkmath.NewInt(-1),
					Status:       types.DELEGATION_QUEUE,
					TxHash:       validTxHash,
				},
			},
			err: "amount < 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddress)

				require.Equal(t, test.msg.Type(), "overwrite_delegation_record", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}

// ----------------------------------------------
//           OverwriteUnbondingRecord
// ----------------------------------------------

func TestMsgOverwriteUnbondingRecord_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validTxHash1 := "69650DCD2D68BBC9310BA4B980187483BFC81452EC99C39DA45633F077157911"
	validTxHash2 := "2B1B1D6C4975E1CE562DBF6C2C4D6BEF543A71D6036A0664C9B4EE2C108E6AF8"
	validAddress, invalidAddress := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgOverwriteUnbondingRecord
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgOverwriteUnbondingRecord{
				Creator: validAddress,
				UnbondingRecord: &types.UnbondingRecord{
					Id:                             1,
					Status:                         types.UNBONDED,
					StTokenAmount:                  sdkmath.NewInt(11),
					NativeAmount:                   sdkmath.NewInt(10),
					UnbondingCompletionTimeSeconds: 1705857114, // unixtime (1/21/24)
					UndelegationTxHash:             validTxHash1,
					UnbondedTokenSweepTxHash:       validTxHash2,
				},
			},
		},
		{
			name: "successful message with blank tx hashes",
			msg: types.MsgOverwriteUnbondingRecord{
				Creator: validAddress,
				UnbondingRecord: &types.UnbondingRecord{
					Id:                             1,
					Status:                         types.UNBONDED,
					StTokenAmount:                  sdkmath.NewInt(11),
					NativeAmount:                   sdkmath.NewInt(10),
					UnbondingCompletionTimeSeconds: 1705857114, // unixtime (1/21/24)
					UndelegationTxHash:             "",
					UnbondedTokenSweepTxHash:       "",
				},
			},
		},

		{
			name: "invalid signer address",
			msg: types.MsgOverwriteUnbondingRecord{
				Creator: invalidAddress, // invalid
				UnbondingRecord: &types.UnbondingRecord{
					Id:                             1,
					Status:                         types.UNBONDED,
					StTokenAmount:                  sdkmath.NewInt(11),
					NativeAmount:                   sdkmath.NewInt(10),
					UnbondingCompletionTimeSeconds: 1705857114,
					UndelegationTxHash:             validTxHash1,
					UnbondedTokenSweepTxHash:       validTxHash2,
				},
			},
			err: "invalid address",
		},
		{
			name: "invalid native amount",
			msg: types.MsgOverwriteUnbondingRecord{
				Creator: validAddress,
				UnbondingRecord: &types.UnbondingRecord{
					Id:                             1,
					Status:                         types.UNBONDED,
					StTokenAmount:                  sdkmath.NewInt(11),
					NativeAmount:                   sdkmath.NewInt(-1), // negative
					UnbondingCompletionTimeSeconds: 1705857114,
					UndelegationTxHash:             validTxHash1,
					UnbondedTokenSweepTxHash:       validTxHash2,
				},
			},
			err: "amount < 0",
		},
		{
			name: "invalid sttoken amount",
			msg: types.MsgOverwriteUnbondingRecord{
				Creator: validAddress,
				UnbondingRecord: &types.UnbondingRecord{
					Id:                             1,
					Status:                         types.UNBONDED,
					StTokenAmount:                  sdkmath.NewInt(-1), // negative
					NativeAmount:                   sdkmath.NewInt(10),
					UnbondingCompletionTimeSeconds: 1705857114,
					UndelegationTxHash:             validTxHash1,
					UnbondedTokenSweepTxHash:       validTxHash2,
				},
			},
			err: "amount < 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddress)

				require.Equal(t, test.msg.Type(), "overwrite_unbonding_record", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}

// ----------------------------------------------
//          OverwriteRedemptionRecord
// ----------------------------------------------

func TestMsgOverwriteRedemptionRecord_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validAddress, invalidAddress := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgOverwriteRedemptionRecord
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgOverwriteRedemptionRecord{
				Creator: validAddress,
				RedemptionRecord: &types.RedemptionRecord{
					UnbondingRecordId: 1,
					Redeemer:          validAddress,
					StTokenAmount:     sdkmath.NewInt(11),
					NativeAmount:      sdkmath.NewInt(10),
				},
			},
		},
		{
			name: "invalid signer address",
			msg: types.MsgOverwriteRedemptionRecord{
				Creator: invalidAddress,
				RedemptionRecord: &types.RedemptionRecord{
					UnbondingRecordId: 1,
					Redeemer:          validAddress,
					StTokenAmount:     sdkmath.NewInt(11),
					NativeAmount:      sdkmath.NewInt(10),
				},
			},
			err: "invalid address",
		},
		{
			name: "invalid redeemer address",
			msg: types.MsgOverwriteRedemptionRecord{
				Creator: validAddress,
				RedemptionRecord: &types.RedemptionRecord{
					UnbondingRecordId: 1,
					Redeemer:          invalidAddress,
					StTokenAmount:     sdkmath.NewInt(11),
					NativeAmount:      sdkmath.NewInt(10),
				},
			},
			err: "invalid address",
		},
		{
			name: "invalid native amount",
			msg: types.MsgOverwriteRedemptionRecord{
				Creator: validAddress,
				RedemptionRecord: &types.RedemptionRecord{
					UnbondingRecordId: 1,
					Redeemer:          validAddress,
					StTokenAmount:     sdkmath.NewInt(11),
					NativeAmount:      sdkmath.NewInt(-1),
				},
			},
			err: "amount < 0",
		},
		{
			name: "invalid sttoken amount",
			msg: types.MsgOverwriteRedemptionRecord{
				Creator: validAddress,
				RedemptionRecord: &types.RedemptionRecord{
					UnbondingRecordId: 1,
					Redeemer:          validAddress,
					StTokenAmount:     sdkmath.NewInt(-1),
					NativeAmount:      sdkmath.NewInt(10),
				},
			},
			err: "amount < 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddress)

				require.Equal(t, test.msg.Type(), "overwrite_redemption_record", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}

// ----------------------------------------------
//          MsgSetOperatorAddress
// ----------------------------------------------

func TestMsgSetOperatorAddress_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validAddress, invalidAddress := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgSetOperatorAddress
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgSetOperatorAddress{
				Signer:   validAddress,
				Operator: validAddress,
			},
		},
		{
			name: "invalid signer address",
			msg: types.MsgSetOperatorAddress{
				Signer:   invalidAddress,
				Operator: validAddress,
			},
			err: "invalid address",
		},
		{
			name: "invalid operator address",
			msg: types.MsgSetOperatorAddress{
				Signer:   validAddress,
				Operator: invalidAddress,
			},
			err: "invalid address",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddress)

				require.Equal(t, test.msg.Type(), "set_operator_address", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
