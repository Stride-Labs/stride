package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type ChangeValidatorTestCase struct {
	hostZone          stakeibctypes.HostZone
	validMsgs         stakeibctypes.MsgChangeValidatorWeight
	initialValidators []*stakeibctypes.Validator
}

func (s *KeeperTestSuite) SetupChangeValidator() ChangeValidatorTestCase {
	initialValidators := []*stakeibctypes.Validator{
		{
			Name:           "val1",
			Address:        "stride_VAL1",
			CommissionRate: 1,
			Weight:         0,
			Status:         stakeibctypes.Validator_ACTIVE,
			DelegationAmt:  0,
		},
		{
			Name:           "val2",
			Address:        "stride_VAL2",
			CommissionRate: 2,
			Weight:         0,
			Status:         stakeibctypes.Validator_ACTIVE,
			DelegationAmt:  0,
		},
	}

	hostZone := stakeibctypes.HostZone{
		ChainId:    "GAIA",
		Validators: initialValidators,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	validMsgs := stakeibctypes.MsgChangeValidatorWeight{
		Creator:  "stride_ADMIN",
		HostZone: "GAIA",
		ValAddr:  "stride_VAL1",
		Weight:   1,
	}

	return ChangeValidatorTestCase{
		hostZone:          hostZone,
		validMsgs:         validMsgs,
		initialValidators: initialValidators,
	}
}

func (s *KeeperTestSuite) TestChangeValidatorWeight_Successful() {
	tc := s.SetupChangeValidator()

	_, err := s.GetMsgServer().ChangeValidatorWeight(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs)
	s.Require().NoError(err)

	HZ, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "HostZone found")

	var IsEqualWeight bool
	for _, val := range HZ.Validators {
		if val.Address == "stride_VAL1" {
			IsEqualWeight = (val.Weight == 1)
		}
	}
	s.Require().True(IsEqualWeight, "Change Successful")

}
func (s *KeeperTestSuite) TestChangeValidatorWeight_HostZoneNotFound() {
	tc := s.SetupChangeValidator()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsgs
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.GetMsgServer().ChangeValidatorWeight(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	s.Require().EqualError(err, "host zone not registered")
}
func (s *KeeperTestSuite) TestChangeValidatorWeight_ValNoFound() {
	tc := s.SetupChangeValidator()
	tc.validMsgs.ValAddr = "stride_VAL3"
	tc.validMsgs.Weight = 1

	_, err := s.GetMsgServer().ChangeValidatorWeight(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs)
	s.Require().EqualError(err, "validator not found")
}
