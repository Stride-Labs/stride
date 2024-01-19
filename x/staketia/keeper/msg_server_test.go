package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	types "github.com/Stride-Labs/stride/v17/x/staketia/types"
)

const (
	ChainId = "CELESTIA"
)

type UpdateInnerRedemptionRateBoundsTestCase struct {
	initialMsg types.MsgUpdateInnerRedemptionRateBounds
	updateMsg  types.MsgUpdateInnerRedemptionRateBounds
	invalidMsg types.MsgUpdateInnerRedemptionRateBounds
}

func (s *KeeperTestSuite) SetupUpdateInnerRedemptionRateBounds() UpdateInnerRedemptionRateBoundsTestCase {
	// Register a host zone
	zone := types.HostZone{
		ChainId: ChainId,
		// Upper bound 1.5
		MaxRedemptionRate: sdk.NewDec(3).Quo(sdk.NewDec(2)),
		// Lower bound 0.9
		MinRedemptionRate: sdk.NewDec(9).Quo(sdk.NewDec(10)),
	}

	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)

	initialMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                s.TestAccs[0].String(),
		MinInnerRedemptionRate: sdk.NewDec(90).Quo(sdk.NewDec(100)),
		MaxInnerRedemptionRate: sdk.NewDec(105).Quo(sdk.NewDec(100)),
	}

	updateMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                s.TestAccs[0].String(),
		MinInnerRedemptionRate: sdk.NewDec(95).Quo(sdk.NewDec(100)),
		MaxInnerRedemptionRate: sdk.NewDec(11).Quo(sdk.NewDec(10)),
	}

	invalidMsg := types.MsgUpdateInnerRedemptionRateBounds{
		Creator:                s.TestAccs[0].String(),
		MinInnerRedemptionRate: sdk.NewDec(0),
		MaxInnerRedemptionRate: sdk.NewDec(2),
	}

	return UpdateInnerRedemptionRateBoundsTestCase{
		initialMsg: initialMsg,
		updateMsg:  updateMsg,
		invalidMsg: invalidMsg,
	}
}

// Verify that bounds can be set successfully
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the inner bounds on the host zone for the first time
	_, err := s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.initialMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "should not throw an error")

	s.Require().Equal(tc.initialMsg.MinInnerRedemptionRate, zone.MinInnerRedemptionRate, "min inner redemption rate should be set")
	s.Require().Equal(tc.initialMsg.MaxInnerRedemptionRate, zone.MaxInnerRedemptionRate, "max inner redemption rate should be set")

	// Update the inner bounds on the host zone
	_, err = s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.updateMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone, err = s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "should not throw an error")

	s.Require().Equal(tc.updateMsg.MinInnerRedemptionRate, zone.MinInnerRedemptionRate, "min inner redemption rate should be set")
	s.Require().Equal(tc.updateMsg.MaxInnerRedemptionRate, zone.MaxInnerRedemptionRate, "max inner redemption rate should be set")

	// Set the inner bounds on the host zone for the first time
	_, err = s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.invalidMsg)
	s.Require().ErrorContains(err, "invalid inner bounds")
}
