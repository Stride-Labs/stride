package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	"fmt"

	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_Success() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	// verify the total amount is expected
	unbond := sdk.NewInt(1_000_000)
	totalAmt, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, tc.hostZone, unbond)
	s.Require().Nil(err)

	// sum up totalAmt
	actualAmount := sdk.ZeroInt()
	for _, amt := range totalAmt {
		actualAmount = actualAmount.Add(amt)
	}
	s.Require().Equal(unbond, actualAmount, "total amount unbonded matches input")

	// verify the order of the validators is expected
	// GetTargetValAmtsForHostZone first reverses the list, then sorts by weight using SliceStable
	// E.g. given A:1, B:2, C:2
	// 1. C:2, B:2, A:1
	// 2. A:1, C:2, B:2
	s.Require().Equal(tc.valNames[0], tc.hostZone.Validators[0].Address)
	s.Require().Equal(tc.valNames[1], tc.hostZone.Validators[2].Address)
	s.Require().Equal(tc.valNames[2], tc.hostZone.Validators[1].Address)
}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_InvalidAmountOfDelegation() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	// if unbond/ finalDelegation is zero then return err
	unbond := sdk.ZeroInt()
	_, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, tc.hostZone, unbond)
	s.Require().EqualError(err, stakeibc.ErrNoValidatorWeights.Error(), "Delegate zero amount should fail")

}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_ErrNoValidatorsWeight() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	unbond := sdk.NewInt(1_000_000)

	// assign zero amount to all validators's weights
	validators := tc.hostZone.GetValidators()
	for _, validator := range validators {
		validator.Weight = 0
	}

	// if weight is zero then return err
	_, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, tc.hostZone, unbond)
	s.Require().EqualError(err, stakeibc.ErrNoValidatorWeights.Error(), "Delegate zero amount should fail")
}

func (s *KeeperTestSuite) SetupGetValidatorDelegationAmtDifferences(validators []*stakeibc.Validator) stakeibc.HostZone {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddr := "cosmos_DELEGATION"

	delegationAccount := stakeibc.ICAAccount{
		Address: delegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}

	hostZone := stakeibc.HostZone{
		ChainId:           "GAIA",
		HostDenom:         "uatom",
		Bech32Prefix:      "cosmos",
		Validators:        validators,
		DelegationAccount: &delegationAccount,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	return hostZone
}

func (s *KeeperTestSuite) TestGetValidatorDelegationAmtDifferences_Successful() {
	validators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: sdk.NewInt(1_000_000),
			Weight:        uint64(1),
		},
	}
	hostZone := s.SetupGetValidatorDelegationAmtDifferences(validators)

	_, err := s.App.StakeibcKeeper.GetValidatorDelegationAmtDifferences(s.Ctx, hostZone)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestGetValidatorDelegationAmtDifferences_ErrorGetTargetValAtmsForHostZone() {
	validators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: sdk.ZeroInt(),
			Weight:        uint64(2),
		},
	}
	hostZone := s.SetupGetValidatorDelegationAmtDifferences(validators)
	_, err := s.App.StakeibcKeeper.GetValidatorDelegationAmtDifferences(s.Ctx, hostZone)
	s.Require().Error(err)
	s.Require().Equal(err, types.ErrNoValidatorWeights)
}
