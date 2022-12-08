package keeper_test

import (
	"fmt"
	"math"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

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
			DelegationAmt: uint64(1_000_000),
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
			DelegationAmt: uint64(0),
			Weight:        uint64(2),
		},
	}
	hostZone := s.SetupGetValidatorDelegationAmtDifferences(validators)
	_, err := s.App.StakeibcKeeper.GetValidatorDelegationAmtDifferences(s.Ctx, hostZone)
	s.Require().Error(err)
	s.Require().Equal(err, types.ErrNoValidatorWeights)
}

func (s *KeeperTestSuite) TestGetValidatorDelegationAmtDifferences_ErrorGetTargetWeightForHostZone() {
	validators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: math.MaxUint64,
			Weight:        uint64(2),
		},
	}
	hostZone := s.SetupGetValidatorDelegationAmtDifferences(validators)
	_, err := s.App.StakeibcKeeper.GetValidatorDelegationAmtDifferences(s.Ctx, hostZone)
	s.Require().Error(err)

	targetDelForVal := hostZone.Validators[0].DelegationAmt
	msgErr := fmt.Errorf("overflow: unable to cast %v of type %T to int64", targetDelForVal, targetDelForVal)
	s.Require().Equal(err, msgErr)
}

func (s *KeeperTestSuite) TestGetValidatorDelegationAmtDifferences_ErrorGetTargetDelAmtForHostZone() {
	validators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR_1",
			DelegationAmt: uint64(1_000_000),
			Weight:        uint64(1),
		},
		{
			Address:       "cosmos_VALIDATOR_2",
			DelegationAmt: math.MaxUint64,
			Weight:        uint64(1),
		},
	}
	hostZone := s.SetupGetValidatorDelegationAmtDifferences(validators)
	_, err := s.App.StakeibcKeeper.GetValidatorDelegationAmtDifferences(s.Ctx, hostZone)
	s.Require().Error(err)

	targetDelForVal := hostZone.Validators[0].DelegationAmt
	msgErr := fmt.Errorf("overflow: unable to cast %v of type %T to int64", targetDelForVal, targetDelForVal)
	s.Require().Equal(err, msgErr)
}
