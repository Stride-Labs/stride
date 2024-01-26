package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	stakeibctypes "github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

type UpdateInnerRedemptionRateBoundsTestCase struct {
	validMsg stakeibctypes.MsgUpdateInnerRedemptionRateBounds
	zone     stakeibctypes.HostZone
}

func (s *KeeperTestSuite) SetupUpdateInnerRedemptionRateBounds() UpdateInnerRedemptionRateBoundsTestCase {
	// Register a host zone
	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		RedemptionRate:    sdk.NewDec(1.0),
		MinRedemptionRate: sdk.NewDec(9).Quo(sdk.NewDec(10)),
		MaxRedemptionRate: sdk.NewDec(15).Quo(sdk.NewDec(10)),
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	defaultMsg := stakeibctypes.MsgUpdateInnerRedemptionRateBounds{
		// TODO: does this need to be the admin address?
		Creator:                s.TestAccs[0].String(),
		ChainId:                HostChainId,
		MinInnerRedemptionRate: sdk.NewDec(1),
		MaxInnerRedemptionRate: sdk.NewDec(11).Quo(sdk.NewDec(10)),
	}

	return UpdateInnerRedemptionRateBoundsTestCase{
		validMsg: defaultMsg,
		zone:     hostZone,
	}
}

// Verify that bounds can be set successfully
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds_Success() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should be in the store")
	s.Require().Equal(tc.validMsg.MinInnerRedemptionRate, zone.MinInnerRedemptionRate, "min inner redemption rate should be set")
	s.Require().Equal(tc.validMsg.MaxInnerRedemptionRate, zone.MaxInnerRedemptionRate, "max inner redemption rate should be set")
}

// Setting inner bounds outside of outer bounds should throw an error
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds_OutOfBounds() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the min inner bound to be less than the min outer bound
	tc.validMsg.MinInnerRedemptionRate = sdk.NewDec(0)

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	// verify it throws an error
	errMsg := fmt.Sprintf("inner min safety threshold (%s) is less than outer min safety threshold (%s)", tc.validMsg.MinInnerRedemptionRate, sdk.NewDec(9).Quo(sdk.NewDec(10)))
	s.Require().ErrorContains(err, errMsg)

	// Set the min inner bound to be valid, but the max inner bound to be greater than the max outer bound
	tc.validMsg.MinInnerRedemptionRate = sdk.NewDec(1)
	tc.validMsg.MaxInnerRedemptionRate = sdk.NewDec(3)
	// Set the inner bounds on the host zone
	_, err = s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	// verify it throws an error
	errMsg = fmt.Sprintf("inner max safety threshold (%s) is greater than outer max safety threshold (%s)", tc.validMsg.MaxInnerRedemptionRate, sdk.NewDec(15).Quo(sdk.NewDec(10)))
	s.Require().ErrorContains(err, errMsg)
}

// Validate basic tests
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds_InvalidMsg() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the min inner bound to be greater than than the max inner bound
	invalidMsg := tc.validMsg
	invalidMsg.MinInnerRedemptionRate = sdk.NewDec(2)

	err := invalidMsg.ValidateBasic()

	// Verify the error
	errMsg := fmt.Sprintf("Inner max safety threshold (%s) is less than inner min safety threshold (%s)", invalidMsg.MaxInnerRedemptionRate, invalidMsg.MinInnerRedemptionRate)
	s.Require().ErrorContains(err, errMsg)
}

// Verify that if inner bounds end up outside of outer bounds (somehow), the outer bounds are returned
func (s *KeeperTestSuite) TestGetInnerSafetyBounds() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the inner bounds outside the outer bounds on the host zone directly
	tc.zone.MinInnerRedemptionRate = sdk.NewDec(0)
	tc.zone.MaxInnerRedemptionRate = sdk.NewDec(3)
	// Set the host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.zone)

	// Get the inner bounds and verify the outer bounds are used
	innerMinSafetyThreshold, innerMaxSafetyThreshold := s.App.StakeibcKeeper.GetInnerSafetyBounds(s.Ctx, tc.zone)
	s.Require().Equal(tc.zone.MinRedemptionRate, innerMinSafetyThreshold, "min inner redemption rate should be set")
	s.Require().Equal(tc.zone.MaxRedemptionRate, innerMaxSafetyThreshold, "max inner redemption rate should be set")
}
