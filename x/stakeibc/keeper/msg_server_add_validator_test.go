package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type AddValidatorTestCase struct {
	hostZone           stakeibctypes.HostZone
	validMsgs          []stakeibctypes.MsgAddValidator
	expectedValidators []*stakeibctypes.Validator
}

func (s *KeeperTestSuite) SetupAddValidator() AddValidatorTestCase {
	hostZone := stakeibctypes.HostZone{
		ChainId:    "GAIA",
		Validators: []*stakeibctypes.Validator{},
	}

	validMsgs := []stakeibctypes.MsgAddValidator{
		{
			Creator:    "stride_ADMIN",
			HostZone:   "GAIA",
			Name:       "val1",
			Address:    "stride_VAL1",
			Commission: 1,
			Weight:     1,
		},
		{
			Creator:    "stride_ADMIN",
			HostZone:   "GAIA",
			Name:       "val2",
			Address:    "stride_VAL2",
			Commission: 2,
			Weight:     2,
		},
		{
			Creator:    "stride_ADMIN",
			HostZone:   "GAIA",
			Name:       "val3",
			Address:    "stride_VAL3",
			Commission: 3,
			Weight:     3,
		},
	}

	expectedValidators := []*stakeibctypes.Validator{
		{
			Name:           "val1",
			Address:        "stride_VAL1",
			CommissionRate: 1,
			Weight:         1,
			Status:         stakeibctypes.Validator_ACTIVE,
			DelegationAmt:  sdk.ZeroInt(),
		},
		{
			Name:           "val2",
			Address:        "stride_VAL2",
			CommissionRate: 2,
			Weight:         2,
			Status:         stakeibctypes.Validator_ACTIVE,
			DelegationAmt:  sdk.ZeroInt(),
		},
		{
			Name:           "val3",
			Address:        "stride_VAL3",
			CommissionRate: 3,
			Weight:         3,
			Status:         stakeibctypes.Validator_ACTIVE,
			DelegationAmt:  sdk.ZeroInt(),
		},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return AddValidatorTestCase{
		hostZone:           hostZone,
		validMsgs:          validMsgs,
		expectedValidators: expectedValidators,
	}
}

func (s *KeeperTestSuite) TestAddValidator_Successful() {
	tc := s.SetupAddValidator()

	// Add first validator
	_, err := s.GetMsgServer().AddValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(1, len(hostZone.Validators), "number of validators should be 1")
	s.Require().Equal(tc.expectedValidators[:1], hostZone.Validators, "validators list after adding 1 validator")

	// Add second validator
	_, err = s.GetMsgServer().AddValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[1])
	s.Require().NoError(err)

	hostZone, found = s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(2, len(hostZone.Validators), "number of validators should be 2")
	s.Require().Equal(tc.expectedValidators[:2], hostZone.Validators, "validators list after adding 2 validators")

	// Add third validator
	_, err = s.GetMsgServer().AddValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[2])
	s.Require().NoError(err)

	hostZone, found = s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(3, len(hostZone.Validators), "number of validators should be 3")
	s.Require().Equal(tc.expectedValidators, hostZone.Validators, "validators list after adding 3 validators")
}

func (s *KeeperTestSuite) TestAddValidator_HostZoneNotFound() {
	tc := s.SetupAddValidator()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsgs[0]
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.GetMsgServer().AddValidator(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	s.Require().EqualError(err, "Host Zone (gaia) not found: host zone not found")
}

func (s *KeeperTestSuite) TestAddValidator_AddressAlreadyExists() {
	tc := s.SetupAddValidator()

	// Update host zone so that val1 already exists and setup our message with val3
	// With no other changes, you would expect this message to be successful
	hostZone := tc.hostZone
	hostZone.Validators = []*stakeibctypes.Validator{tc.expectedValidators[0]} // val1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	validMsg := tc.validMsgs[2] // val3

	// Change the validator address to val1 so that the message errors
	badMsg := validMsg
	badMsg.Address = "stride_VAL1"
	_, err := s.GetMsgServer().AddValidator(sdk.WrapSDKContext(s.Ctx), &badMsg)
	s.Require().EqualError(err, "Validator address (stride_VAL1) already exists on Host Zone (GAIA): validator already exists")
}

func (s *KeeperTestSuite) TestAddValidator_NameAlreadyExists() {
	tc := s.SetupAddValidator()

	// Update host zone so that val1 already exists and setup our message with val3
	// With no other changes, you would expect this message to be successful
	hostZone := tc.hostZone
	hostZone.Validators = []*stakeibctypes.Validator{tc.expectedValidators[0]} // val1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	validMsg := tc.validMsgs[2] // val3

	// Change the validator name to val1 so that the message errors
	badMsg := validMsg
	badMsg.Name = "val1"
	_, err := s.GetMsgServer().AddValidator(sdk.WrapSDKContext(s.Ctx), &badMsg)
	s.Require().EqualError(err, "Validator name (val1) already exists on Host Zone (GAIA): validator already exists")
}
