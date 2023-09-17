package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type EnableStrictUnbondingCapTestCase struct {
	validMsg stakeibctypes.MsgEnableStrictUnbondingCap
}

func (s *KeeperTestSuite) SetupEnableStrictUnbondingCap() EnableStrictUnbondingCapTestCase {
	return EnableStrictUnbondingCapTestCase{
		validMsg: stakeibctypes.MsgEnableStrictUnbondingCap{},
	}
}

func (s *KeeperTestSuite) TestEnableStrictUnbondingCap_CapNotSet() {
	s.SetupEnableStrictUnbondingCap()

	// make sure StrictUnbondingCap is not set
	s.Require().False(s.App.StakeibcKeeper.IsStrictUnbondingEnabled(s.Ctx), "strict unbonding cap not set by default")
}

func (s *KeeperTestSuite) TestEnableStrictUnbondingCap_CapSet() {
	tc := s.SetupEnableStrictUnbondingCap()

	// set StrictUnbondingCap
	msgServer := s.GetMsgServer()
	_, err := msgServer.EnableStrictUnbondingCap(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "enable strict unbonding cap should not error")

	// make sure StrictUnbondingCap is set
	s.Require().True(s.App.StakeibcKeeper.IsStrictUnbondingEnabled(s.Ctx), "strict unbonding cap set to true")
}
