package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type AddValidatorTestCase struct {
	initialValidators []*types.Validator
	validMsg          types.MsgAddValidator
}

func (s *KeeperTestSuite) SetupAddValidator() AddValidatorTestCase {
	hostZone := stakeibc.HostZone{
		ChainId: "GAIA",
		Validators: []*types.Validator{
			{
				Name:    "val1",
				Address: "stride_VAL1",
			},
		},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return AddValidatorTestCase{
		initialValidators: hostZone.Validators,
		validMsg: types.MsgAddValidator{
			Creator:    "stride_ADMIN",
			HostZone:   "GAIA",
			Name:       "val2",
			Address:    "stride_VAL2",
			Commission: 10,
			Weight:     10,
		},
	}
}

func (s *KeeperTestSuite) AddValidator_Successful() {
	tc := s.SetupAddValidator()

	_, err := s.msgServer.AddValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err)

	actualValidatorList1, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
}

func (s *KeeperTestSuite) AddValidator_HostZoneNotFound() {

}

func (s *KeeperTestSuite) AddValidator_AddressAlreadyExists() {

}

func (s *KeeperTestSuite) AddValidator_NameAlreadyExists() {

}
