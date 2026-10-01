package v35_test

import (
	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
)

// setFailedLSMDeposit stores the mainnet deposit with the given status and amount
func (s *UpgradeTestSuite) setFailedLSMDeposit(status recordstypes.LSMTokenDeposit_Status, amount sdkmath.Int) {
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordstypes.LSMTokenDeposit{
		ChainId: v35.FailedLSMDepositChainId,
		Denom:   v35.FailedLSMDepositDenom,
		Amount:  amount,
		Status:  status,
	})
}

func (s *UpgradeTestSuite) mustGetFailedLSMDeposit() recordstypes.LSMTokenDeposit {
	deposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v35.FailedLSMDepositChainId, v35.FailedLSMDepositDenom)
	s.Require().True(found, "LSM deposit should exist")
	return deposit
}

func (s *UpgradeTestSuite) TestResetFailedLSMDeposit() {
	s.setFailedLSMDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED, v35.FailedLSMDepositAmount)
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordstypes.LSMTokenDeposit{
		ChainId: v35.FailedLSMDepositChainId,
		Denom:   "cosmosvaloperX/1",
		Amount:  sdkmath.NewInt(100),
		Status:  recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED,
	})

	v35.ResetFailedLSMDeposit(s.Ctx, s.App.RecordsKeeper)

	deposit := s.mustGetFailedLSMDeposit()
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE, deposit.Status, "status")
	s.Require().Equal(int64(67_850_951), deposit.Amount.Int64(), "amount")

	other, _ := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v35.FailedLSMDepositChainId, "cosmosvaloperX/1")
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED, other.Status, "other deposits untouched")
	s.Require().Equal(int64(100), other.Amount.Int64(), "other deposit amount untouched")

	// A second run finds the record already requeued and leaves it alone
	v35.ResetFailedLSMDeposit(s.Ctx, s.App.RecordsKeeper)
	s.Require().Equal(int64(67_850_951), s.mustGetFailedLSMDeposit().Amount.Int64(), "not decremented twice")
}

func (s *UpgradeTestSuite) TestResetFailedLSMDeposit_Skips() {
	testCases := []struct {
		name   string
		status recordstypes.LSMTokenDeposit_Status
		amount sdkmath.Int
	}{
		{name: "already requeued", status: recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE, amount: v35.FailedLSMDepositAmount},
		{name: "detokenization in progress", status: recordstypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS, amount: v35.FailedLSMDepositAmount},
		{name: "unexpected amount", status: recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED, amount: sdkmath.NewInt(67_850_951)},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.setFailedLSMDeposit(tc.status, tc.amount)

			v35.ResetFailedLSMDeposit(s.Ctx, s.App.RecordsKeeper)

			deposit := s.mustGetFailedLSMDeposit()
			s.Require().Equal(tc.status, deposit.Status, "status untouched")
			s.Require().Equal(tc.amount, deposit.Amount, "amount untouched")
		})
	}
}

func (s *UpgradeTestSuite) TestResetFailedLSMDeposit_MissingDeposit() {
	s.Require().NotPanics(func() { v35.ResetFailedLSMDeposit(s.Ctx, s.App.RecordsKeeper) })
}
