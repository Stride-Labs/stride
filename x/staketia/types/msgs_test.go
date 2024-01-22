package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v17/app/apptesting"
	"github.com/Stride-Labs/stride/v17/x/staketia/types"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v17/testutil/sample"
)

// ----------------------------------------------
//        MsgResumeHostZone
// ----------------------------------------------

func TestMsgResumeHostZone(t *testing.T) {
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
//        MsgUpdateInnerRedemptionRateBounds
// ----------------------------------------------

func TestMsgUpdateInnerRedemptionRateBounds(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validUpperBound := sdk.NewDec(2)
	validLowerBound := sdk.NewDec(1)
	invalidLowerBound := sdk.NewDec(2)

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
//               MsgLiquidStake
// ----------------------------------------------

func TestMsgLiquidStake_ValidateBasic(t *testing.T) {
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
				Staker:       "celestia1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				NativeAmount: sdkmath.NewInt(1000000),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid inputs",
			msg: types.MsgLiquidStake{
				Staker:       sample.AccAddress(),
				NativeAmount: sdkmath.NewInt(1200000),
			},
		},
		{
			name: "amount below threshold",
			msg: types.MsgLiquidStake{
				Staker:       sample.AccAddress(),
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

	expected := `{"type":"staketia/MsgLiquidStake","value":{"native_amount":"1000","staker":"stride1v9jxgu33kfsgr5"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgRedeemStake
// ----------------------------------------------

func TestMsgRedeemStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgRedeemStake
		err  error
	}{
		{
			name: "success",
			msg: types.MsgRedeemStake{
				Redeemer:      sample.AccAddress(),
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
				Redeemer:      sample.AccAddress(),
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

	expected := `{"type":"staketia/MsgRedeemStake","value":{"redeemer":"stride1v9jxgu33kfsgr5","st_token_amount":"1000000"}}`
	require.Equal(t, expected, string(res))
}

func TestMsgSetOperatorAddress(t *testing.T) {
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

// ----------------------------------------------
//               OverwriteDelegationRecord
// ----------------------------------------------

func TestMsgOverwriteDelegationRecord(t *testing.T) {
	apptesting.SetupConfig()

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
					TxHash:       "TXHASH",
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
					TxHash:       "TXHASH",
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
					TxHash:       "TXHASH",
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
//               OverwriteUnbondingRecord
// ----------------------------------------------

func TestMsgOverwriteUnbondingRecord(t *testing.T) {
	apptesting.SetupConfig()

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
					UndelegationTxHash:             "TXHASH1",
					UnbondedTokenSweepTxHash:       "TXHASH2",
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
					UndelegationTxHash:             "TXHASH1",
					UnbondedTokenSweepTxHash:       "TXHASH2",
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
					UndelegationTxHash:             "TXHASH1",
					UnbondedTokenSweepTxHash:       "TXHASH2",
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
					UndelegationTxHash:             "TXHASH1",
					UnbondedTokenSweepTxHash:       "TXHASH2",
				},
			},
			err: "amount < 0",
		},
		{
			name: "invalid undelegation txhash",
			msg: types.MsgOverwriteUnbondingRecord{
				Creator: validAddress,
				UnbondingRecord: &types.UnbondingRecord{
					Id:                             1,
					Status:                         types.UNBONDING_ARCHIVE, // should work
					StTokenAmount:                  sdkmath.NewInt(11),
					NativeAmount:                   sdkmath.NewInt(10),
					UnbondingCompletionTimeSeconds: 1705857114,
					UndelegationTxHash:             "",
					UnbondedTokenSweepTxHash:       "TXHASH2",
				},
			},
			err: "transaction hash cannot be empty",
		},
		{
			name: "invalid unbonded sweep txhash",
			msg: types.MsgOverwriteUnbondingRecord{
				Creator: validAddress,
				UnbondingRecord: &types.UnbondingRecord{
					Id:                             1,
					Status:                         types.UNBONDING_ARCHIVE, // should work
					StTokenAmount:                  sdkmath.NewInt(11),
					NativeAmount:                   sdkmath.NewInt(10),
					UnbondingCompletionTimeSeconds: 1705857114,
					UndelegationTxHash:             "TXHASH1",
					UnbondedTokenSweepTxHash:       "",
				},
			},
			err: "transaction hash cannot be empty",
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
//               OverwriteRedemptionRecord
// ----------------------------------------------

func TestMsgOverwriteRedemptionRecord(t *testing.T) {
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
