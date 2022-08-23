package keeper_test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type DeleteValidatorTestCase struct {
	hostZone          types.HostZone
	initialValidators []*types.Validator
	validMsgs         []types.MsgDeleteValidator
}

func (s *KeeperTestSuite) SetupDeleteValidator() DeleteValidatorTestCase {
	initialValidators := []*types.Validator{
		{
			Name:           "val1",
			Address:        "stride_VAL1",
			CommissionRate: 1,
			Weight:         1,
			Status:         types.Validator_Active,
			DelegationAmt:  0,
		},
		{
			Name:           "val2",
			Address:        "stride_VAL2",
			CommissionRate: 2,
			Weight:         2,
			Status:         types.Validator_Active,
			DelegationAmt:  0,
		},
	}

	hostZone := stakeibc.HostZone{
		ChainId:    "GAIA",
		Validators: initialValidators,
	}
	validMsgs := []types.MsgDeleteValidator{
		{
			Creator:  "stride_ADDRESS",
			HostZone: "GAIA",
			ValAddr:  "stride_VAL2",
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
	_, err := s.msgServer.DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(1, len(hostZone.Validators), "number of validators should be 1")
	s.Require().Equal(tc.initialValidators[1:], hostZone.Validators)

	// Delete second validator
	_, err = s.msgServer.DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[1])
	s.Require().NoError(err)

	hostZone, found = s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(0, len(hostZone.Validators), "number of validators should be 0")
	s.Require().Equal([]*types.Validator{}, hostZone.Validators)
}

func (s *KeeperTestSuite) TestDeleteValidator_HostZoneNotFound() {
	tc := s.SetupDeleteValidator()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsgs[0]
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.msgServer.DeleteValidator(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	s.Require().EqualError(err, fmt.Sprintf("Host Zone (gaia) not found: host zone not found"))
}

func (s *KeeperTestSuite) TestAddValidator_AddressNotFound() {
	_ = s.SetupDeleteValidator()
}

func (s *KeeperTestSuite) TestAddValidator_NonZeroDelegations() {
	_ = s.SetupDeleteValidator()
}
