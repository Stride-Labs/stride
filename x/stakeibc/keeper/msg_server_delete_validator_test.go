package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type DeleteValidatorTestCase struct {
	hostZone          stakeibctypes.HostZone
	initialValidators []*stakeibctypes.Validator
	validMsgs         []stakeibctypes.MsgDeleteValidator
}

func (s *KeeperTestSuite) SetupDeleteValidator() DeleteValidatorTestCase {
	initialValidators := []*stakeibctypes.Validator{
		{
			Name:           "val1",
			Address:        "stride_VAL1",
			CommissionRate: 1,
			Weight:         0,
			Status:         stakeibctypes.Validator_ACTIVE,
			DelegationAmt:  sdk.ZeroInt(),
		},
		{
			Name:           "val2",
			Address:        "stride_VAL2",
			CommissionRate: 2,
			Weight:         0,
			Status:         stakeibctypes.Validator_ACTIVE,
			DelegationAmt:  sdk.ZeroInt(),
		},
	}

	hostZone := stakeibctypes.HostZone{
		ChainId:    "GAIA",
		Validators: initialValidators,
	}
	validMsgs := []stakeibctypes.MsgDeleteValidator{
		{
			Creator:  "stride_ADDRESS",
			HostZone: "GAIA",
			ValAddr:  "stride_VAL1",
		},
		{
			Creator:  "stride_ADDRESS",
			HostZone: "GAIA",
			ValAddr:  "stride_VAL2",
		},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return DeleteValidatorTestCase{
		hostZone:          hostZone,
		initialValidators: initialValidators,
		validMsgs:         validMsgs,
	}
}

func (s *KeeperTestSuite) TestDeleteValidator_Successful() {
	tc := s.SetupDeleteValidator()

	// Delete first validator
	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(1, len(hostZone.Validators), "number of validators should be 1")
	s.Require().Equal(tc.initialValidators[1:], hostZone.Validators, "validators list after removing 1 validator")

	// Delete second validator
	_, err = s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[1])
	s.Require().NoError(err)

	hostZone, found = s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(0, len(hostZone.Validators), "number of validators should be 0")
}

func (s *KeeperTestSuite) TestDeleteValidator_HostZoneNotFound() {
	tc := s.SetupDeleteValidator()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsgs[0]
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	errMsg := "Validator (stride_VAL1) not removed from host zone (gaia) "
	errMsg += "| err: HostZone (gaia) not found: host zone not found: validator not removed"
	s.Require().EqualError(err, errMsg)
}

func (s *KeeperTestSuite) TestDeleteValidator_AddressNotFound() {
	tc := s.SetupDeleteValidator()

	// Build message with a validator address that does not exist
	badAddressMsg := tc.validMsgs[0]
	badAddressMsg.ValAddr = "stride_VAL5"
	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &badAddressMsg)

	errMsg := "Validator (stride_VAL5) not removed from host zone (GAIA) "
	errMsg += "| err: Validator address (stride_VAL5) not found on host zone (GAIA): "
	errMsg += "validator not found: validator not removed"
	s.Require().EqualError(err, errMsg)
}

func (s *KeeperTestSuite) TestDeleteValidator_NonZeroDelegation() {
	tc := s.SetupDeleteValidator()

	// Update val1 to have a non-zero delegation
	hostZone := tc.hostZone
	hostZone.Validators[0].DelegationAmt = sdk.NewInt(1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	errMsg := "Validator (stride_VAL1) not removed from host zone (GAIA) "
	errMsg += "| err: Validator (stride_VAL1) has non-zero delegation (1) or weight (0): "
	errMsg += "validator not removed"
	s.Require().EqualError(err, errMsg)
}

func (s *KeeperTestSuite) TestDeleteValidator_NonZeroWeight() {
	tc := s.SetupDeleteValidator()

	// Update val1 to have a non-zero weight
	hostZone := tc.hostZone
	hostZone.Validators[0].Weight = 1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	errMsg := "Validator (stride_VAL1) not removed from host zone (GAIA) "
	errMsg += "| err: Validator (stride_VAL1) has non-zero delegation (0) or weight (1): "
	errMsg += "validator not removed"
	s.Require().EqualError(err, errMsg)
}
