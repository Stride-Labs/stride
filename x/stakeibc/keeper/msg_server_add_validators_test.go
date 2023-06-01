package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type AddValidatorsTestCase struct {
	hostZone           stakeibctypes.HostZone
	validMsg           stakeibctypes.MsgAddValidators
	expectedValidators []*stakeibctypes.Validator
}

func (s *KeeperTestSuite) SetupAddValidators() AddValidatorsTestCase {
	hostZone := stakeibctypes.HostZone{
		ChainId:    "GAIA",
		Validators: []*stakeibctypes.Validator{},
	}

	validMsg := stakeibctypes.MsgAddValidators{
		Creator:  "stride_ADMIN",
		HostZone: "GAIA",
		Validators: []*types.Validator{
			{Name: "val1", Address: "stride_VAL1", Weight: 1},
			{Name: "val2", Address: "stride_VAL2", Weight: 2},
			{Name: "val3", Address: "stride_VAL3", Weight: 3},
		},
	}

	expectedValidators := []*types.Validator{
		{Name: "val1", Address: "stride_VAL1", Weight: 1, DelegationAmt: sdkmath.ZeroInt()},
		{Name: "val2", Address: "stride_VAL2", Weight: 2, DelegationAmt: sdkmath.ZeroInt()},
		{Name: "val3", Address: "stride_VAL3", Weight: 3, DelegationAmt: sdkmath.ZeroInt()},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return AddValidatorsTestCase{
		hostZone:           hostZone,
		validMsg:           validMsg,
		expectedValidators: expectedValidators,
	}
}

func (s *KeeperTestSuite) TestAddValidators_Successful() {
	tc := s.SetupAddValidators()

	// Add validators
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(3, len(hostZone.Validators), "number of validators")

	for i := 0; i < 3; i++ {
		s.Require().Equal(*tc.expectedValidators[i], *hostZone.Validators[i], "validators %d", i)
	}
}

func (s *KeeperTestSuite) TestAddValidators_HostZoneNotFound() {
	tc := s.SetupAddValidators()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsg
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	s.Require().EqualError(err, "Host Zone (gaia) not found: host zone not found")
}

func (s *KeeperTestSuite) TestAddValidators_AddressAlreadyExists() {
	tc := s.SetupAddValidators()

	// Update host zone so that the name val1 already exists
	hostZone := tc.hostZone
	duplicateVal := stakeibctypes.Validator{Name: "new_val", Address: tc.expectedValidators[0].Address}
	hostZone.Validators = []*stakeibctypes.Validator{&duplicateVal}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Change the validator address to val1 so that the message errors
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "Validator address (stride_VAL1) already exists on Host Zone (GAIA): validator already exists")
}

func (s *KeeperTestSuite) TestAddValidators_NameAlreadyExists() {
	tc := s.SetupAddValidators()

	// Update host zone so that val1's address already exists
	hostZone := tc.hostZone
	duplicateVal := stakeibctypes.Validator{Name: tc.expectedValidators[0].Name, Address: "new_address"}
	hostZone.Validators = []*stakeibctypes.Validator{&duplicateVal}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Change the validator name to val1 so that the message errors
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "Validator name (val1) already exists on Host Zone (GAIA): validator already exists")
}
