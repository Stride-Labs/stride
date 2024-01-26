package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	stakeibctypes "github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

type ResumeHostZoneTestCase struct {
	validMsg stakeibctypes.MsgResumeHostZone
	zone     stakeibctypes.HostZone
}

func (s *KeeperTestSuite) SetupResumeHostZone() ResumeHostZoneTestCase {
	// Register a host zone
	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		RedemptionRate:    sdk.NewDec(1.0),
		MinRedemptionRate: sdk.NewDec(9).Quo(sdk.NewDec(10)),
		MaxRedemptionRate: sdk.NewDec(15).Quo(sdk.NewDec(10)),
		Halted:            true,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	defaultMsg := stakeibctypes.MsgResumeHostZone{
		Creator: s.TestAccs[0].String(),
		ChainId: HostChainId,
	}

	return ResumeHostZoneTestCase{
		validMsg: defaultMsg,
		zone:     hostZone,
	}
}

// Verify that bounds can be set successfully
func (s *KeeperTestSuite) TestResumeHostZone_Success() {
	tc := s.SetupResumeHostZone()

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().ResumeHostZone(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should be in the store")

	s.Require().False(zone.Halted, "host zone should not be halted")
}

// verify that non-admins can't call the tx
func (s *KeeperTestSuite) TestResumeHostZone_NonAdmin() {
	tc := s.SetupResumeHostZone()

	invalidMsg := tc.validMsg
	invalidMsg.Creator = s.TestAccs[1].String()

	err := invalidMsg.ValidateBasic()
	s.Require().Error(err, "nonadmins shouldn't be able to call this tx")
}

// verify that the function can't be called on missing zones
func (s *KeeperTestSuite) TestResumeHostZone_MissingZones() {
	tc := s.SetupResumeHostZone()

	invalidMsg := tc.validMsg
	invalidChainId := "invalid-chain"
	invalidMsg.ChainId = invalidChainId

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().ResumeHostZone(s.Ctx, &invalidMsg)

	s.Require().Error(err, "shouldn't be able to call tx on missing zones")
	expectedErrorMsg := fmt.Sprintf("invalid chain id, zone for %s not found: host zone not found", invalidChainId)
	s.Require().Equal(expectedErrorMsg, err.Error(), "should return correct error msg")
}

// verify that the function can't be called on unhalted zones
func (s *KeeperTestSuite) TestResumeHostZone_UnhaltedZones() {
	tc := s.SetupResumeHostZone()

	zone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should be in the store")
	s.Require().True(zone.Halted, "host zone should be halted")
	zone.Halted = false
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, zone)

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().ResumeHostZone(s.Ctx, &tc.validMsg)
	s.Require().Error(err, "shouldn't be able to call tx on unhalted zones")
	expectedErrorMsg := fmt.Sprintf("invalid chain id, zone for %s not halted: host zone is not halted", HostChainId)
	s.Require().Equal(expectedErrorMsg, err.Error(), "should return correct error msg")
}
